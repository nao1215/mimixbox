//
// mimixbox/internal/applets/mbutils/path/path.go
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
package path

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jessevdk/go-flags"
	mb "github.com/nao1215/mimixbox/internal/lib"
)

const cmdName string = "path"

const version = "1.0.2"

var osExit = os.Exit
var errNotGetAbsPath = errors.New("can't get absolute path")

// Exit code
const (
	ExitSuccess int = iota // 0
	ExitFailuer
)

type options struct {
	Abs       bool `short:"a" long:"absolute" description:"Print absolute path"`
	Base      bool `short:"b" long:"basename" description:"Print basename (filename)"`
	Canonical bool `short:"c" long:"canonical" description:"Print canonical path (default)"`
	Dir       bool `short:"d" long:"dirname" description:"Print path without filename"`
	Ext       bool `short:"e" long:"extension" description:"Print file extention"`
	Version   bool `short:"v" long:"version" description:"Show path command version"`
}

func Run() (int, error) {
	var opts options
	var args []string
	var err error

	if args, err = parseArgs(&opts); err != nil {
		return ExitFailuer, nil
	}
	path := args[0]

	if opts.Abs {
		abs, err := filepath.Abs(path)
		if err != nil {
			return ExitFailuer, errNotGetAbsPath
		}
		fmt.Fprintf(os.Stdout, "%s\n", abs)
	}

	if opts.Base {
		fmt.Fprintf(os.Stdout, "%s\n", filepath.Base(path))
	}

	if opts.Canonical || isAllOptionsOff(opts) {
		fmt.Fprintf(os.Stdout, "%s\n", filepath.Clean(path))
	}

	if opts.Dir {
		fmt.Fprintf(os.Stdout, "%s\n", filepath.Dir(path))
	}

	if opts.Ext {
		fmt.Fprintf(os.Stdout, "%s\n", filepath.Ext(path))
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

func isAllOptionsOff(opts options) bool {
	return !opts.Abs && !opts.Base && !opts.Canonical && !opts.Dir && !opts.Ext
}
