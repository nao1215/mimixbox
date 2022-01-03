//
// mimixbox/internal/applets/console-tools/gzip/gzip.go
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
package gzipCmd

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	mb "github.com/nao1215/mimixbox/internal/lib"

	"github.com/jessevdk/go-flags"
)

const cmdName string = "gzip"

const version = "1.0.0"

var osExit = os.Exit

// Exit code
const (
	ExitSuccess int = iota // 0
	ExitFailure
)

type options struct {
	Decomp  bool `short:"d" long:"decompress" description:"Decompress gzip file"`
	Force   bool `short:"f" long:"force" description:"Overwrite file if same name file already exists"`
	Version bool `short:"v" long:"version" description:"Show gzip command version"`
}

func Run() (int, error) {
	var opts options
	var err error
	var args []string

	if args, err = parseArgs(&opts); err != nil {
		return ExitFailure, nil
	}
	return __gzip(args, opts)
}

func __gzip(args []string, opts options) (int, error) {
	status := ExitSuccess
	for _, v := range args {
		path := os.ExpandEnv(v)

		if opts.Decomp && !strings.HasSuffix(path, ".gz") {
			path = path + ".gz"
		}

		if !mb.Exists(path) {
			fmt.Fprintln(os.Stderr, cmdName+": No such file or directory")
			status = ExitFailure
			continue
		}

		if mb.IsDir(path) {
			fmt.Fprintln(os.Stderr, cmdName+": "+v+" is a directory -- ignored")
			status = ExitFailure
			continue
		}

		if opts.Decomp {
			err := decompress(path, opts)
			if err != nil {
				fmt.Fprintln(os.Stderr, cmdName+": "+err.Error())
				status = ExitFailure
			}
		} else {
			err := compress(path, opts)
			if err != nil {
				fmt.Fprintln(os.Stderr, cmdName+": "+err.Error())
				status = ExitFailure
			}
		}
	}
	return status, nil
}

func compress(path string, opts options) error {
	gzFile := fileNameWithGz(path)
	question := cmdName + ": " + filepath.Base(gzFile) + " already exists; do you wish to overwrite ?"
	if mb.Exists(gzFile) && !opts.Force {
		if !mb.Question(question) {
			fmt.Fprintln(os.Stdout, "       not overwritten")
			return nil
		}
	}

	dest, err := os.Create(gzFile)
	if err != nil {
		return err
	}
	defer dest.Close()

	gzipWriter, err := gzip.NewWriterLevel(dest, gzip.BestCompression)
	if err != nil {
		return err
	}
	defer gzipWriter.Close()

	src, err := os.Open(path)
	if err != nil {
		return err
	}
	defer src.Close()

	if _, err := io.Copy(gzipWriter, src); err != nil {
		return err
	}

	if err := mb.RemoveFile(path, false); err != nil {
		return err
	}
	return nil
}

func decompress(path string, opts options) error {
	decompFile := fileNameWithoutGz(path)
	question := cmdName + ": " + filepath.Base(decompFile) + " already exists; do you wish to overwrite ?"
	if mb.Exists(decompFile) && !opts.Force {
		if !mb.Question(question) {
			fmt.Fprintln(os.Stdout, "       not overwritten")
			return nil
		}
	}

	dest, err := os.Create(decompFile)
	if err != nil {
		return err
	}
	defer dest.Close()

	src, err := os.Open(path)
	if err != nil {
		return err
	}
	defer src.Close()

	gzipReader, err := gzip.NewReader(src)
	if err != nil {
		return err
	}
	defer gzipReader.Close()

	if _, err := io.Copy(dest, gzipReader); err != nil {
		return err
	}

	if err := mb.RemoveFile(path, false); err != nil {
		return err
	}
	return nil
}

func fileNameWithGz(path string) string {
	return path + ".gz"
}

func fileNameWithoutGz(path string) string {
	return strings.TrimSuffix(path, ".gz")
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
	if !isValidArgNr(args) {
		fmt.Fprintln(os.Stderr, cmdName+": compressed data not written to a terminal")
		osExit(ExitFailure)
	}
	return args, nil
}

func initParser(opts *options) *flags.Parser {
	parser := flags.NewParser(opts, flags.Default)
	parser.Name = cmdName
	parser.Usage = "[OPTIONS] FILEs"

	return parser
}

func isValidArgNr(args []string) bool {
	return len(args) >= 1
}
