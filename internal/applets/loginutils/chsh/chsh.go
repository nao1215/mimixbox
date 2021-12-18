//
// mimixbox/internal/applets/loginutils/chsh/chsh.go
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
package chsh

import (
	"os"

	mb "github.com/nao1215/mimixbox/internal/lib"

	"github.com/jessevdk/go-flags"
)

// TODO: This file stopped implementation prematurely.
// PAM (pluggable authentication modules) can only be used in a Linux environment
// MimixBox needs to be compatible with Linux environments without PAM, and will
// be compatible with Mac in the future.
// First of all, in order to verify Basic authentication, we stopped the implementation halfway.
// In mb library, Basic authentication using PAM is commented out.

const cmdName string = "chsh"

const version = "1.0.0"

var osExit = os.Exit

// Exit code
const (
	ExitSuccess int = iota // 0
	ExitFailuer
)

type options struct {
	Version bool `short:"v" long:"version" description:"Show chsh command version"`
}

func Run() (int, error) {
	var opts options
	var err error
	var args []string

	if args, err = parseArgs(&opts); err != nil {
		return ExitFailuer, nil
	}
	return chsh(args, opts)
}

func chsh(args []string, opts options) (int, error) {
	status := ExitSuccess

	if len(args) == 0 {
		interactiveChangeShell(opts)
	}
	//isRoot := mb.IsRootUser()
	return status, nil
}

func interactiveChangeShell(opts options) (int, error) {
	/*
		user, err := user.Current()
		if err != nil {
			return ExitFailuer, err
		}
			if err = mb.AuthByPasswordWithPam(user.Username); err != nil {
				return ExitFailuer, err
			}
	*/
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
	return args, nil
}

func initParser(opts *options) *flags.Parser {
	parser := flags.NewParser(opts, flags.Default)
	parser.Name = cmdName
	parser.Usage = "[OPTIONS] [LOGIN_USERNAME]"

	return parser
}
