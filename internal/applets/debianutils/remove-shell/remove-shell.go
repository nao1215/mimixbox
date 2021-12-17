//
// mimixbox/internal/applets/debianutils/remove-shell/remove-shell.go
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
package removeShell

import (
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"
	mb "github.com/nao1215/mimixbox/internal/lib"
)

const cmdName string = "remove-shell"

const version = "1.0.0"

var osExit = os.Exit

type options struct {
	Version bool `short:"v" long:"version" description:"Show remove-shell command version"`
}

// Exit code
const (
	ExitSuccess int = iota // 0
	ExitFailuer
)

func Run() (int, error) {
	var opts options
	var args []string
	var err error

	if args, err = parseArgs(&opts); err != nil {
		return ExitFailuer, nil
	}
	return removeShell(args)
}

func removeShell(args []string) (int, error) {
	f, err := os.OpenFile(mb.TmpShellsFile(), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return ExitFailuer, err
	}
	defer f.Close()

	lines, err := mb.ReadFileToStrList(mb.ShellsFilePath)
	if err != nil {
		return ExitFailuer, err
	}

	lines = mb.ChopAll(lines)
	for _, shell := range args {
		lines = mb.Remove(lines, shell)
	}
	for _, v := range lines {
		fmt.Fprintln(f, v)
	}

	err = mb.Copy(mb.TmpShellsFile(), mb.ShellsFilePath)
	if err != nil {
		mb.RemoveFile(mb.TmpShellsFile(), false)
		return ExitFailuer, err
	}

	mb.RemoveFile(mb.TmpShellsFile(), false)
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
		fmt.Fprintln(os.Stderr, cmdName+": shellname [shellname ...]")
		osExit(ExitFailuer)
	}
	return args, nil
}

func initParser(opts *options) *flags.Parser {
	parser := flags.NewParser(opts, flags.Default)
	parser.Name = cmdName
	parser.Usage = "[OPTIONS] SHELL_NAME"

	return parser
}

func isValidArgNr(args []string) bool {
	return len(args) >= 1
}
