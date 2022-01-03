//
// mimixbox/internal/applets/textutils/head/head.go
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
package head

import (
	"fmt"
	"os"
	"strings"

	"github.com/jessevdk/go-flags"
	mb "github.com/nao1215/mimixbox/internal/lib"
)

const cmdName string = "head"

const version = "1.0.4"

var osExit = os.Exit

// Exit code
const (
	ExitSuccess int = iota // 0
	ExitFailure
)

type options struct {
	Lines   int  `short:"n" long:"lines" default:"10" description:"Print the first NUM lines instead of the first 10"`
	Version bool `short:"v" long:"version" description:"Show head command version"`
}

func Run() (int, error) {
	var opts options
	var args []string
	var err error

	if args, err = parseArgs(&opts); err != nil {
		return ExitFailure, nil
	}

	if len(args) == 0 || mb.Contains(args, "-") {
		mb.Parrot(false)
		return ExitSuccess, nil
	}

	err = head(args, opts)
	if err != nil {
		return ExitFailure, err
	}

	return ExitSuccess, nil
}

func head(args []string, opts options) error {
	for _, v := range args {
		var output []string
		var err error
		target := os.ExpandEnv(v)
		if len(args) >= 2 {
			printNameBanner(target)
		}
		if !mb.Exists(target) {
			output = strings.Split(args[0], "\n")
		} else if mb.IsDir(target) {
			output = append(output, target+" is directory")
		} else if mb.IsFile(target) {
			output, err = mb.ReadFileToStrList(target)
			if err != nil {
				return err
			}
			output = mb.ChopAll(output)
		}

		if opts.Lines <= len(output) {
			output = output[:opts.Lines]
		}
		dump(output)
	}
	return nil
}

func dump(s []string) {
	for _, v := range s {
		fmt.Fprintln(os.Stdout, v)
	}
}

func printNameBanner(path string) {
	fmt.Fprintf(os.Stdout, "==> %s <==\n", path)
}

func parseArgs(opts *options) ([]string, error) {
	p := initParser(opts)
	args, err := p.Parse()
	if err != nil {
		return nil, err
	}

	if mb.HasPipeData() {
		stdin, err := mb.FromPIPE()
		if err != nil {
			return nil, err
		}
		return []string{stdin}, nil
	}

	if opts.Version {
		mb.ShowVersion(cmdName, version)
		osExit(ExitSuccess)
	}

	if opts.Lines < 0 {
		opts.Lines = 10
	}

	return args, nil
}

func initParser(opts *options) *flags.Parser {
	parser := flags.NewParser(opts, flags.Default)
	parser.Name = cmdName
	parser.Usage = "[OPTIONS] [FILE_PATH]"

	return parser
}
