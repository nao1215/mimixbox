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
	"strconv"
	"strings"
	"time"

	"github.com/cheggaaa/pb/v3"
	"github.com/jessevdk/go-flags"
	mb "github.com/nao1215/mimixbox/internal/lib"
)

const cmdName string = "sddf"
const ext string = ".sddf"

type Paths []string

type fileInfo struct {
	path     string
	checksum string
}

const version = "1.0.3"

var osExit = os.Exit

// Exit code
const (
	ExitSuccess int = iota // 0
	ExitFailure
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
		return ExitFailure, nil
	}
	return sddfMainSeq(os.ExpandEnv(args[0]), opts)
}

func sddfMainSeq(path string, opts options) (int, error) {
	if !mb.Exists(path) {
		return ExitFailure, errors.New(path + " does not exists")
	}

	if mb.IsFile(path) {
		return restoreAndDelete(path)
	}
	return search(path, opts)
}

func restoreAndDelete(path string) (int, error) {
	if !strings.HasSuffix(path, ext) {
		return ExitFailure, errors.New(path + ": file format is not *.sddf")
	}

	df, err := restore(path)
	if err != nil {
		return ExitFailure, err
	}

	return deleteFiles(df)
}

func deleteFiles(df map[string]Paths) (int, error) {
	var status int = ExitSuccess
	var deleteFileList []string

	fmt.Fprintln(os.Stdout, "Decide delete target files")
	for _, v := range df {
		list, err := decideDeleteTarget(v)
		if err != nil {
			return ExitFailure, err
		}
		deleteFileList = append(deleteFileList, list...)
	}

	fmt.Fprintln(os.Stdout, "Start deleting files")
	var sumByteSize int64 = 0
	for _, v := range deleteFileList {
		size, err := mb.Size(v)
		if err != nil {
			status = ExitFailure
			fmt.Fprintln(os.Stdout, "Delete(Failure): "+v)
			continue
		}

		err = mb.RemoveFile(v, false)
		if err != nil {
			status = ExitFailure
			fmt.Fprintln(os.Stdout, "Delete(Failure): "+v)
		} else {
			fmt.Fprintln(os.Stdout, "Delete(Success): "+v+": "+strconv.FormatInt(size, 10)+"Byte")
		}
		sumByteSize += size
	}
	fmt.Fprintln(os.Stdout, "End deleting files. Size="+strconv.FormatInt(sumByteSize, 10)+"Byte")
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
	for k := range fileModTime {
		if k == leaveKey {
			continue
		}
		deleteTarget = append(deleteTarget, k)
	}
	return deleteTarget, nil
}

func processing(cancel chan struct{}) {
	i := 0
	for {
		select {
		case <-cancel:
			time.Sleep(1 * time.Second)
			return
		default:
			if i != 80 {
				fmt.Fprintf(os.Stdout, ".")
				time.Sleep(100 * time.Millisecond)
				i++
			} else {
				fmt.Fprintln(os.Stdout, "")
				i = 0
			}
		}
	}
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
	files := findFiles(path)

	if len(files) == 0 {
		fmt.Fprintln(os.Stdout, path+" has no file")
		return ExitSuccess, nil
	}
	return dumpToFile(getSameFiles(files), decideOutputFileName(opts.Output))
}

func findFiles(path string) []string {
	fmt.Fprintln(os.Stdout, "Get all file path at "+path)

	cancel := (make(chan struct{}))
	go processing(cancel)
	_, files, _ := mb.Walk(path, true)
	files = excludeNamedPipe(excludeImportantFiles(files))
	close(cancel)
	fmt.Fprintln(os.Stdout)

	return files
}

func excludeImportantFiles(files []string) []string {
	ng := []string{"/boot", "/dev", "/etc", "/lib", "/lib32", "/lib64", "/libx32", "/lost+found",
		"/proc", "/root", "/run", "/sys", "/bin", "/sbin"}

	newFileList := []string{}
	for _, path := range files {
		for i, v := range ng {
			if strings.HasPrefix(path, v) {
				break
			}
			if (i + 1) == len(ng) {
				newFileList = append(newFileList, path)
			}
		}
	}
	return newFileList
}

// Explain why we want to get rid of named PIPE.
// sddf command calculates the checksums of all files to verify file identity.
// The checksum calculation for the named PIPE will stop unless there is writing
// to the named PIPE. It's looks like deadlock. To avoid this problem, exclude
// named PIPE from target file list
func excludeNamedPipe(files []string) []string {
	newFileList := []string{}
	for _, path := range files {
		if mb.IsNamedPipe(path) {
			continue
		}
		newFileList = append(newFileList, path)
	}
	return newFileList
}

func dumpToFile(df map[string]Paths, output string) (int, error) {
	fmt.Fprintln(os.Stdout, "Write down duplicated file list to "+output)
	f, err := os.Create(output)
	if err != nil {
		return ExitFailure, err
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
		return ExitFailure, err
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
	fmt.Fprintln(os.Stdout, "Find the same file on a file content basis")
	totalFileNum := len(files)
	ch := make(chan fileInfo, totalFileNum)
	defer close(ch)

	if totalFileNum < 100 {
		go calcChecksumThread(files, ch)
	} else {
		for n := 0; n < 10; n++ {
			ratio := totalFileNum / 10
			first := n * ratio
			last := ((n + 1) * ratio)
			if n == 9 {
				last = totalFileNum
			}
			go calcChecksumThread(files[first:last], ch)
		}
	}

	df := map[string]Paths{}
	bar := pb.Simple.Start(totalFileNum)
	bar.SetMaxWidth(80)
	for n := 0; n < totalFileNum; n++ {
		fi := <-ch
		paths, ok := df[fi.checksum]
		if ok {
			paths = append(paths, fi.path)
			df[fi.checksum] = paths
		} else {
			df[fi.checksum] = []string{fi.path}
		}
		bar.Increment()
	}
	bar.Finish()
	return df
}

func calcChecksumThread(files []string, ch chan fileInfo) {
	var fi fileInfo = fileInfo{"", ""}
	for _, v := range files {
		fi = fileInfo{"", ""}
		checksum, err := mb.CalcChecksum(md5.New(), v)
		if err != nil {
			//fmt.Fprintln(os.Stderr, cmdName+": "+err.Error())
			ch <- fi
			continue
		}
		fi.path = v
		fi.checksum = checksum
		ch <- fi
	}
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
		osExit(ExitFailure)
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
