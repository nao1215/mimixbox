//
// mimixbox/internal/applets/fileutils/rm/rm.go
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
package rm

import (
	"errors"
	"fmt"
	"os"
	"strings"

	mb "github.com/nao1215/mimixbox/internal/lib"

	"github.com/jessevdk/go-flags"
)

const cmdName string = "rm"

const version = "1.0.3"

var osExit = os.Exit

type options struct {
	Force       bool `short:"f" long:"force" description:"Ignore non-existent files, not prompt"`
	Interactive bool `short:"i" long:"interactive" description:"Ask every time if you want to remove"`
	NoPreserve  bool `short:"n" long:"no-preserve-root" description:"Allow deletion of root directory"`
	Recursive   bool `short:"r" long:"recursive" description:"Remove directories and their contents recursively"`
	Version     bool `short:"v" long:"version" description:"Show rm command version"`
}

func Run() (int, error) {
	var opts options
	var args []string
	var err error
	status := mb.ExitSuccess

	if args, err = parseArgs(&opts); err != nil {
		return mb.ExitFailure, nil
	}

	for _, path := range args {
		if s, err := rm(path, opts); err != nil {
			fmt.Fprintln(os.Stderr, cmdName+": "+err.Error())
			status = s
		}
	}

	return status, nil
}

func rm(path string, opts options) (int, error) {
	p := os.ExpandEnv(path)
	if status, err := validBeforeRemove(p, opts); status != mb.ExitSuccess {
		return status, err
	}

	if mb.IsFile(p) {
		if err := mb.RemoveFile(p, opts.Interactive); err != nil {
			return mb.ExitFailure, err
		}
	}

	if err := mb.RemoveDir(p, opts.Interactive); err != nil {
		return mb.ExitFailure, err
	}

	return mb.ExitSuccess, nil
}

func validBeforeRemove(path string, opts options) (int, error) {
	if mb.IsRootDir(path) && !opts.NoPreserve {
		return mb.ExitFailure, errors.New("do not remove the root directory")
	}

	if !mb.Exists(path) {
		if !opts.Force {
			return mb.ExitFailure, errors.New("can't remove " + path + ": No such file or directory exists")
		}
		return mb.ExitFailure, nil
	}

	if mb.IsDir(path) && !opts.Recursive {
		return mb.ExitFailure, errors.New("can't remove " + path + ": It's directory")
	}

	return mb.ExitSuccess, nil
}

func parseArgs(opts *options) ([]string, error) {
	p := initParser(opts)

	args, err := p.Parse()
	if err != nil {
		return nil, err
	}

	if mb.HasPipeData() && len(args) == 0 {
		stdin, err := mb.FromPIPE()
		if err != nil {
			return nil, err
		}
		lines := strings.Split(stdin, "\n")
		return mb.AddLineFeed(lines[:len(lines)-1]), nil
	}

	if opts.Version {
		mb.ShowVersion(cmdName, version)
		osExit(mb.ExitSuccess)
	}

	if !isValidArgNr(args) {
		showHelp(p)
		osExit(mb.ExitFailure)
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

func showHelp(p *flags.Parser) {
	p.WriteHelp(os.Stdout)
}
