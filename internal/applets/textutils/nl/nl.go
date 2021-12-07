//
// mimixbox/internal/applets/textutils/nl/nl.go
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
package nl

import (
	"fmt"
	"os"

	mb "github.com/nao1215/mimixbox/internal/lib"

	"github.com/jessevdk/go-flags"
)

const cmdName string = "nl"

const version = "1.0.2"

var osExit = os.Exit

// Exit code
const (
	ExitSuccess int = iota // 0
	ExitFailuer
)

type options struct {
	Version bool `short:"v" long:"version" description:"Show nl command version"`
}

func Run() (int, error) {
	var opts options
	var args []string
	var err error

	if args, err = parseArgs(&opts); err != nil {
		return ExitFailuer, nil
	}

	if mb.HasPipeData() && len(os.Args) == 1 {
		mb.PrintStrListWithNumberLine(args, true)
		return ExitSuccess, nil
	}

	var pipeData []string
	if len(args) == 0 || mb.Contains(args, "-") {
		var nr int = 1
		for {
			input, next := mb.Input()
			if !next {
				break
			}
			pipeData = append(pipeData, input+"\n")
			if input != "" && len(args) == 0 {
				mb.PrintStrWithNumberLine(nr, "  %6d  %s", input+"\n")
				nr++
			} else if input == "" && len(args) == 0 {
				fmt.Fprintln(os.Stdout, "")
			}
		}
		if len(args) == 0 {
			return ExitSuccess, nil
		}
		// If this case, Heredocuments and files may be concatenated.
		args = mb.Remove(args, "-")
	}

	lines, err := mb.Concatenate(args)
	if err != nil {
		return ExitFailuer, err
	}

	if len(pipeData) >= 1 {
		lines = append(pipeData, lines...)
	}
	mb.PrintStrListWithNumberLine(lines, false)

	return ExitSuccess, nil
}

func parseArgs(opts *options) ([]string, error) {
	p := initParser(opts)

	args, err := p.Parse()
	if err != nil {
		return nil, err
	}

	if mb.HasPipeData() && len(args) == 0 {
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

	return args, nil
}

func initParser(opts *options) *flags.Parser {
	parser := flags.NewParser(opts, flags.Default)
	parser.Name = cmdName
	parser.Usage = "[OPTIONS] FILE_PATH"

	return parser
}
