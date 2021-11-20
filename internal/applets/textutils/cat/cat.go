//
// mimixbox/internal/applets/textutils/cat/cat.go
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
package cat

import (
	"fmt"
	mb "mimixbox/internal/lib"
	"os"

	"github.com/jessevdk/go-flags"
)

const cmdName string = "cat"

const version = "1.0.4"

var osExit = os.Exit

// Exit code
const (
	ExitSuccess int = iota // 0
	ExitFailuer
)

type options struct {
	Number  bool `short:"n" long:"number" description:"Print with line number"`
	Version bool `short:"v" long:"version" description:"Show cat command version"`
}

func Run() (int, error) {
	var opts options
	var args []string
	var err error

	if args, err = parseArgs(&opts); err != nil {
		return ExitFailuer, nil
	}

	if len(args) == 0 || mb.Contains(args, "-") {
		mb.Parrot(opts.Number)
		return ExitSuccess, nil
	}

	strLisr, err := mb.Concatenate(args, false)
	if err != nil {
		return ExitFailuer, nil
	}

	if opts.Number {
		mb.PrintStrListWithNumberLine(strLisr, true)
	} else {
		for _, str := range strLisr {
			fmt.Print(str)
		}
	}

	return ExitSuccess, nil
}

func parseArgs(opts *options) ([]string, error) {
	p := initParser(opts)

	args, err := p.Parse()
	if err != nil {
		return nil, err
	}

	if opts.Version {
		showVersion()
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

func showVersion() {
	description := cmdName + " version " + version + " (under Apache License verison 2.0)\n"
	fmt.Print(description)
}
