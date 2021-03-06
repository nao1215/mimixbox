//
// mimixbox/internal/applets/shellutils/echo/echo.go
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
package echo

import (
	"fmt"
	"os"
	"strings"

	"github.com/jessevdk/go-flags"
	mb "github.com/nao1215/mimixbox/internal/lib"
)

const cmdName string = "echo"
const version = "1.0.2"

var osExit = os.Exit

type options struct {
	Version bool `short:"v" long:"version" description:"Show echo command version"`
}

func Run() (int, error) {
	var output []string
	var opts options
	var args []string
	var err error

	if args, err = parseArgs(&opts); err != nil {
		return mb.ExitFailure, nil
	}

	if len(args) == 0 {
		output = make([]string, 0)
	} else {
		output = os.Args[1:]
	}

	fmt.Fprintln(os.Stdout, strings.Join(output, " "))
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
	parser.Usage = "[OPTIONS] STRING"

	return parser
}
