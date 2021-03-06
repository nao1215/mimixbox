//
// mimixbox/internal/applets/fileutils/mkdir/mkdir.go
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
package mkdir

import (
	"errors"
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"
	mb "github.com/nao1215/mimixbox/internal/lib"
)

const cmdName string = "mkdir"

const version = "1.0.3"

var osExit = os.Exit

var ErrNoOperand = errors.New("no operand")

type options struct {
	Parent  bool `short:"p" long:"parents" description:"No error if existing, make parent directories as needed"`
	Version bool `short:"v" long:"version" description:"Show mkdir command version"`
}

func Run() (int, error) {
	var opts options
	var args []string
	var err error

	if args, err = parseArgs(&opts); err != nil {
		if err == ErrNoOperand {
			return mb.ExitFailure, err
		}
		return mb.ExitFailure, nil
	}

	status := mb.ExitSuccess
	for _, path := range args {
		target := os.ExpandEnv(path)
		if opts.Parent {
			err = os.MkdirAll(target, 0755)
		} else {
			err = os.Mkdir(target, 0755)
		}

		if err != nil {
			status = mb.ExitFailure
			fmt.Fprintln(os.Stderr, err.Error())
		}
	}
	return status, nil
}

func parseArgs(opts *options) ([]string, error) {
	p := initParser(opts)

	args, err := p.Parse()
	if err != nil {
		return nil, err
	}

	if opts.Version {
		mb.ShowVersion(cmdName, version)
		osExit(mb.ExitSuccess)
	}

	if !isValidArgNr(args) {
		return nil, ErrNoOperand
	}
	return args, nil
}

func initParser(opts *options) *flags.Parser {
	parser := flags.NewParser(opts, flags.Default)
	parser.Name = cmdName
	parser.Usage = "[OPTIONS] PATH"

	return parser
}

func isValidArgNr(args []string) bool {
	return len(args) >= 1
}
