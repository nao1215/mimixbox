//
// mimixbox/internal/applets/fileutils/mkfifo/mkfifo.go
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
package mkfifo

import (
	"fmt"
	"os"
	"syscall"

	mb "github.com/nao1215/mimixbox/internal/lib"

	"github.com/jessevdk/go-flags"
)

const cmdName string = "mkfifo"

const version = "1.0.2"

var osExit = os.Exit

// Exit code
const (
	ExitSuccess int = iota // 0
	ExitFailuer
)

type options struct {
	Version bool `short:"v" long:"version" description:"Show mkfifo command version"`
}

func Run() (int, error) {
	var opts options
	var args []string
	var err error
	status := ExitSuccess

	if args, err = parseArgs(&opts); err != nil {
		return ExitFailuer, nil
	}

	for _, path := range args {
		p := os.ExpandEnv(path)
		if mb.Exists(p) {
			status = ExitFailuer
			fmt.Fprintln(os.Stderr, cmdName+": can't make "+p+": already exist")
			continue
		}
		if err := syscall.Mkfifo(p, 0644); err != nil {
			status = ExitFailuer
			fmt.Fprintln(os.Stderr, cmdName+": "+p+": "+err.Error())
			continue
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
	parser.Usage = "[OPTIONS] FILE_PATH"

	return parser
}

func isValidArgNr(args []string) bool {
	return len(args) >= 1
}

func showHelp(p *flags.Parser) {
	p.WriteHelp(os.Stdout)
}
