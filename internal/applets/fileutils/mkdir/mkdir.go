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
	"os"

	"github.com/jessevdk/go-flags"
	mb "github.com/nao1215/mimixbox/internal/lib"
)

const cmdName string = "mkdir"

const version = "1.0.2"

var osExit = os.Exit

// Exit code
const (
	ExitSuccess int = iota // 0
	ExitFailuer
)

type options struct {
	Parent  bool `short:"p" long:"parents" description:"No error if existing, make parent directories as needed"`
	Version bool `short:"v" long:"version" description:"Show mkdir command version"`
}

func Run() (int, error) {
	var opts options
	var args []string
	var err error

	if args, err = parseArgs(&opts); err != nil {
		return ExitFailuer, nil
	}
	path := os.ExpandEnv(args[0])

	if opts.Parent {
		err = os.MkdirAll(path, 0755)
	} else {
		err = os.Mkdir(path, 0755)
	}

	if err != nil {
		return ExitFailuer, err
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
		mb.ShowVersion(cmdName, version)
		osExit(ExitSuccess)
	}

	if !isValidArgNr(args) {
		showHelp(p)
		osExit(ExitFailuer)
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
	return len(args) == 1
}

func showHelp(p *flags.Parser) {
	p.WriteHelp(os.Stdout)
}
