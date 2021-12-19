//
// mimixbox/internal/applets/debianutils/valid-shell/valid-shell.go
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
package validShell

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/jessevdk/go-flags"
	mb "github.com/nao1215/mimixbox/internal/lib"
)

const cmdName string = "valid-shell"

const version = "1.0.0"

var osExit = os.Exit

type options struct {
	Show    bool `short:"s" long:"show" description:"Print contents of /etc/shells"`
	Fix     bool `short:"f" long:"fix" description:"Fix problems in /etc/shells"`
	Version bool `short:"v" long:"version" description:"Show valid-shell command version"`
}

// Exit code
const (
	ExitSuccess int = iota // 0
	ExitFailuer
)

func Run() (int, error) {
	var opts options
	var err error

	if _, err = parseArgs(&opts); err != nil {
		return ExitFailuer, nil
	}

	return validShell(opts)
}

func validShell(opts options) (int, error) {
	if opts.Show {
		return printShellsFile()
	} else if opts.Fix {
		return fix()
	}
	return valid()
}

func printShellsFile() (int, error) {
	lines, err := mb.ReadFileToStrList(mb.ShellsFilePath)
	if err != nil {
		return ExitFailuer, err
	}
	for _, v := range lines {
		fmt.Fprintf(os.Stdout, "%s", v)
	}
	return ExitSuccess, nil
}

func fix() (int, error) {
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
	for _, v := range lines {
		if strings.HasPrefix(v, "#") {
			fmt.Fprintln(f, v)
			continue
		}
		if isFalseCmd(v) {
			continue // NG: Bupass
		}
		if mb.Exists(v) {
			fmt.Fprintln(f, v)
		} // else is NG: Bypass
	}

	err = mb.Copy(mb.TmpShellsFile(), mb.ShellsFilePath)
	if err != nil {
		mb.RemoveFile(mb.TmpShellsFile(), false)
		return ExitFailuer, err
	}
	mb.RemoveFile(mb.TmpShellsFile(), false)

	return ExitSuccess, nil
}

func valid() (int, error) {
	lines, err := mb.ReadFileToStrList(mb.ShellsFilePath)
	if err != nil {
		return ExitFailuer, err
	}

	lines = mb.ChopAll(lines)
	for _, v := range lines {
		if strings.HasPrefix(v, "#") {
			continue
		}
		if isFalseCmd(v) {
			fmt.Fprintf(os.Stdout, "NG: %s (not preferable for security)\n", v)
			continue
		}
		if mb.Exists(v) {
			fmt.Fprintf(os.Stdout, "OK: %s\n", v)
		} else {
			fmt.Fprintf(os.Stdout, "NG: %s (not exist in the system)\n", v)
		}
	}
	return ExitSuccess, nil
}

func isFalseCmd(str string) bool {
	path, err := exec.LookPath("false")
	if err != nil {
		return false
	}
	return path == str
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

	return args, nil
}

func initParser(opts *options) *flags.Parser {
	parser := flags.NewParser(opts, flags.Default)
	parser.Name = cmdName
	parser.Usage = "[OPTIONS]"

	return parser
}
