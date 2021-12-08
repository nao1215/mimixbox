//
// mimixbox/internal/applets/fileutils/rmdir/rmdir.go
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
package rmdir

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	mb "github.com/nao1215/mimixbox/internal/lib"

	"github.com/jessevdk/go-flags"
)

const cmdName string = "rmdir"

const version = "1.0.1"

var osExit = os.Exit

// Exit code
const (
	ExitSuccess int = iota // 0
	ExitFailuer
)

type options struct {
	Ignore  bool `short:"i" long:"ignore-fail-on-non-empty" description:"Ignore the error if the directory is not empty"`
	Parents bool `short:"p" long:"parents" description:"Remove DIRECTORY and its parents"`
	Version bool `short:"v" long:"version" description:"Show rmdir command version"`
}

func Run() (int, error) {
	var opts options
	var args []string
	var err error
	var status int

	if args, err = parseArgs(&opts); err != nil {
		return ExitFailuer, nil
	}

	// Coremb will continue to delete files as much as possible.
	// MimixBox stops processing if an error occurs even once.
	for _, path := range args {
		if status, err := rmdir(path, opts); err != nil {
			return status, err
		}
	}
	return status, nil
}

func rmdir(path string, opts options) (int, error) {
	p := os.ExpandEnv(path)
	if err := validBeforeRemove(p); err != nil {
		return ExitFailuer, err
	}

	var target string
	if opts.Parents {
		target = ancestorDir(p)
	} else {
		target = p
	}

	_, files, err := mb.Walk(target, false)
	if err != nil {
		return ExitFailuer, err
	}
	if len(files) != 0 {
		if opts.Ignore {
			return ExitSuccess, nil
		}
		return ExitFailuer, errors.New("Can't remove " + path + ": It's not empty directory")
	}

	// The contents of the directory are empty. Delete directories at once
	if err := os.RemoveAll(target); err != nil {
		return ExitFailuer, err
	}
	return ExitSuccess, nil
}

func ancestorDir(path string) string {
	dirs := strings.Split(path, string(filepath.Separator))
	return dirs[0]
}

func validBeforeRemove(path string) error {
	if !mb.Exists(path) {
		return errors.New("can't remove " + path + ": No such file or directory exists")
	}

	if mb.IsFile(path) {
		return errors.New("can't remove " + path + ": It's not directory")
	}

	return nil
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
	return len(args) >= 1
}

func showHelp(p *flags.Parser) {
	p.WriteHelp(os.Stdout)
}
