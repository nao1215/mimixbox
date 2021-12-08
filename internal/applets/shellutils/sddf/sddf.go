//
// mimixbox/internal/applets/shellutils/sddf/sddf.go
//
// Copyright 2021 Naohiro CHIKAMATSU
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package sddf

import (
	"crypto/md5"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cheggaaa/pb/v3"
	"github.com/jessevdk/go-flags"
	mb "github.com/nao1215/mimixbox/internal/lib"
)

const cmdName string = "sddf"
const ext string = ".sddf"

type Paths []string

const version = "1.0.0"

var osExit = os.Exit

// Exit code
const (
	ExitSuccess int = iota // 0
	ExitFailuer
)

type options struct {
	Output  string `short:"o" long:"output" default:"duplicated-file.sddf" description:"Change output file-name without extension"`
	Version bool   `short:"v" long:"version" description:"Show sddf command version"`
}

func Run() (int, error) {
	var opts options
	var args []string
	var err error

	if args, err = parseArgs(&opts); err != nil {
		return ExitFailuer, nil
	}
	return sddfMainSeq(os.ExpandEnv(args[0]), opts)
}

func sddfMainSeq(path string, opts options) (int, error) {
	if !mb.Exists(path) {
		return ExitFailuer, errors.New(path + " does not exists")
	}

	if mb.IsFile(path) {
		return restoreAndDelete(path)
	}
	return search(path, opts)
}

func restoreAndDelete(path string) (int, error) {
	if !strings.HasSuffix(path, ext) {
		return ExitFailuer, errors.New(path + ": file format is not *.sddf")
	}

	df, err := restore(path)
	if err != nil {
		return ExitFailuer, err
	}

	return deleteFiles(df)
}

func deleteFiles(df map[string]Paths) (int, error) {
	var status int = ExitSuccess
	var deleteFileList []string

	fmt.Fprintln(os.Stdout, "Decide delete files")
	for _, v := range df {
		list, err := decideDeleteTarget(v)
		if err != nil {
			return ExitFailuer, err
		}
		deleteFileList = append(deleteFileList, list...)
	}

	fmt.Fprintln(os.Stdout, "Delete files")
	for _, v := range deleteFileList {
		err := mb.RemoveFile(v, false)
		if err != nil {
			status = ExitFailuer
			fmt.Fprintln(os.Stdout, "Delete(Failuer): "+v)
		} else {
			fmt.Fprintln(os.Stdout, "Delete(Success): "+v)
		}
	}
	return status, nil
}

func decideDeleteTarget(paths []string) ([]string, error) {
	fileModTime := map[string]int64{}
	for _, path := range paths {
		f, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		finfo, _ := f.Stat()
		dsec := finfo.ModTime().Unix()
		fileModTime[path] = dsec
	}

	var newest int64 = 0
	var leaveKey string = ""
	for k, v := range fileModTime {
		if newest < v {
			newest = v
			leaveKey = k
		}
	}

	deleteTarget := []string{}
	for k, _ := range fileModTime {
		if k == leaveKey {
			continue
		}
		deleteTarget = append(deleteTarget, k)
	}
	return deleteTarget, nil
}

func restore(path string) (map[string]Paths, error) {
	df := map[string]Paths{}
	lines, err := mb.ReadFileToStrList(path)
	if err != nil {
		return nil, err
	}

	fmt.Fprintln(os.Stdout, "Restore data from "+path)
	bar := pb.Simple.Start(len(lines))
	bar.SetMaxWidth(80)
	var checksum string
	var paths = Paths{}
	lines = mb.ChopAll(lines)
	for _, v := range lines {
		bar.Increment()
		if isChecksumLine(v) {
			checksum = v
			paths = Paths{}
			continue
		}
		if v == "" {
			df[checksum] = paths
			checksum = ""
			paths = Paths{}
			continue
		}
		if v != "" && !isChecksumLine(v) {
			paths = append(paths, v)
		}
	}
	bar.Finish()

	return df, nil
}

func isChecksumLine(line string) bool {
	return strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") && len(line) == 34
}

func search(path string, opts options) (int, error) {
	fmt.Fprintln(os.Stdout, "Get all file path at "+path)
	_, files, err := mb.Walk(path)
	if err != nil {
		return ExitFailuer, nil
	}

	if len(files) == 0 {
		fmt.Fprintf(os.Stdout, path+" has no file")
		return ExitSuccess, nil
	}
	return dumpToFile(getSameFiles(files), decideOutputFileName(opts.Output))
}

func dumpToFile(df map[string]Paths, output string) (int, error) {
	f, err := os.Create(output)
	if err != nil {
		return ExitFailuer, err
	}
	defer f.Close()

	data := ""
	bar := pb.Simple.Start(len(df))
	bar.SetMaxWidth(80)
	for k, v := range df {
		data += "[" + k + "]\n"
		for _, path := range v {
			data += path + "\n"
		}
		data += "\n"
		bar.Increment()
	}
	bar.Finish()

	b := []byte(data)
	_, err = f.Write(b)
	if err != nil {
		return ExitFailuer, err
	}

	fmt.Fprintln(os.Stdout, "See duplicated file list: "+output)
	fmt.Fprintln(os.Stdout, "If you delete files, execute the following command.")
	fmt.Fprintln(os.Stdout, "$ sddf "+output)
	return ExitSuccess, nil
}

func getSameFiles(files []string) map[string]Paths {
	df := calcChecksum(files)
	for k, v := range df {
		if len(v) <= 1 {
			delete(df, k)
		}
	}
	return df
}

func calcChecksum(files []string) map[string]Paths {
	totalFileNum := len(files)
	df := map[string]Paths{}

	fmt.Fprintln(os.Stdout, "Check same file or not")
	bar := pb.Simple.Start(totalFileNum)
	bar.SetMaxWidth(80)
	for _, v := range files {
		absPath, err := filepath.Abs(v)
		if err != nil {
			fmt.Fprintln(os.Stderr, cmdName+": "+err.Error())
			continue
		}

		checksum, err := mb.CalcChecksum(md5.New(), v)
		if err != nil {
			fmt.Fprintln(os.Stderr, cmdName+": "+err.Error())
			continue
		}

		paths, ok := df[checksum]
		if ok {
			paths = append(paths, absPath)
			df[checksum] = paths
		} else {
			df[checksum] = []string{absPath}
		}
		bar.Increment()
	}
	bar.Finish()
	return df
}

func decideOutputFileName(output string) string {
	if strings.HasSuffix(output, ext) {
		return output
	}
	return output + ext // Do not exclude original file-extension
}

func parseArgs(opts *options) ([]string, error) {
	p := initParser(opts)

	args, err := p.Parse()
	if err != nil {
		return nil, err
	}

	if opts.Version {
		mb.ShowVersion(cmdName, version)
		osExit(ExitSuccess)
	}

	if !isValidArg(args, *opts) {
		showHelp(p)
		osExit(ExitFailuer)
	}

	return args, nil
}

func isValidArg(args []string, opts options) bool {
	return len(args) == 1
}

func showHelp(p *flags.Parser) {
	p.WriteHelp(os.Stdout)
}

func initParser(opts *options) *flags.Parser {
	parser := flags.NewParser(opts, flags.Default)
	parser.Name = cmdName
	parser.Usage = "[OPTIONS] PATH"

	return parser
}
