//
// mimixbox/internal/applets/textutils/expand/expand.go
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
package expand

import (
	"fmt"
	"os"
	"strings"

	"github.com/jessevdk/go-flags"
	mb "github.com/nao1215/mimixbox/internal/lib"
)

const cmdName string = "expand"
const version = "1.0.3"

var osExit = os.Exit

type options struct {
	Tab     int  `short:"t" long:"tab" default:"8" description:"Convert TAB to N space (default:N=8)"`
	Version bool `short:"v" long:"version" description:"Show expand command version"`
}

// Exit code
const (
	ExitSuccess int = iota // 0
	ExitFailure
)

func Run() (int, error) {
	var opts options
	var err error
	var args []string

	if args, err = parseArgs(&opts); err != nil {
		return ExitSuccess, nil
	}

	if mb.HasPipeData() {
		mb.Dump(mb.AddLineFeed(strings.Split(args[0], "\n")), false)
		return ExitSuccess, nil
	}

	if len(args) == 0 || mb.Contains(args, "-") {
		mb.Parrot(false)
		return ExitSuccess, nil
	}

	return expand(args, opts)
}

func expand(args []string, opts options) (int, error) {
	status := ExitSuccess
	for _, file := range args {
		target := os.ExpandEnv(file)
		if !mb.IsFile(target) {
			fmt.Fprintln(os.Stderr, target+": No such file. Skip it")
			status = ExitFailure
			continue
		}
		lines, err := mb.ReadFileToStrList(target)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			status = ExitFailure
			continue
		}

		mb.Dump(mb.ReplaceAll(lines, "\t", strings.Repeat(" ", opts.Tab)), false)
	}
	return status, nil
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

	if opts.Tab <= 0 {
		opts.Tab = 8
	}

	return args, nil
}

func initParser(opts *options) *flags.Parser {
	parser := flags.NewParser(opts, flags.Default)
	parser.Name = cmdName
	parser.Usage = "[OPTIONS] FILE_NAME"

	return parser
}
