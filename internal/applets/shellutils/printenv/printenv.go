//
// mimixbox/internal/applets/shellutils/printenv/printenv.go
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
package printenv

import (
	"fmt"
	"os"

	mb "github.com/nao1215/mimixbox/internal/lib"

	"github.com/jessevdk/go-flags"
)

const cmdName string = "printenv"

const version = "1.0.0"

var osExit = os.Exit

type options struct {
	Null    bool `short:"0" long:"null" description:"End each output line with NULL temination, not newline"`
	Version bool `short:"v" long:"version" description:"Show printenv command version"`
}

func Run() (int, error) {
	var opts options
	var err error
	var args []string

	if args, err = parseArgs(&opts); err != nil {
		return mb.ExitFailure, nil
	}
	return printenv(args, opts)
}

func printenv(args []string, opts options) (int, error) {
	if len(args) == 0 {
		return printAllEnvironmentVar(opts)
	}
	for _, v := range args {
		if opts.Null {
			fmt.Fprintf(os.Stdout, "%s", os.Getenv(v))
		} else {
			fmt.Fprintln(os.Stdout, os.Getenv(v))
		}
	}
	return mb.ExitSuccess, nil
}

func printAllEnvironmentVar(opts options) (int, error) {
	for _, e := range os.Environ() {
		if opts.Null {
			fmt.Fprintf(os.Stdout, "%s", e)
		} else {
			fmt.Fprintln(os.Stdout, e)
		}
	}
	return mb.ExitSuccess, nil
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
	return args, nil
}

func initParser(opts *options) *flags.Parser {
	parser := flags.NewParser(opts, flags.Default)
	parser.Name = cmdName
	parser.Usage = "[OPTIONS] ENV"

	return parser
}
