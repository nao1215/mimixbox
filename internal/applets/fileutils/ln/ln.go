//
// mimixbox/internal/applets/fileutils/ln/ln.go
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
package ln

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/jessevdk/go-flags"
	mb "github.com/nao1215/mimixbox/internal/lib"
)

const cmdName string = "ln"

const version = "1.0.1"

var osExit = os.Exit

type options struct {
	Force    bool `short:"f" long:"force" description:"If file exists, forcibly overwrite it"`
	Symbolic bool `short:"s" long:"symbolic" description:"Create symbolic link"`
	Version  bool `short:"v" long:"version" description:"Show cp command version"`
}

func Run() (int, error) {
	var opts options
	var args []string
	var err error

	if args, err = parseArgs(&opts); err != nil {
		return mb.ExitFailure, nil
	}

	err = ln(args, opts)
	if err != nil {
		return mb.ExitFailure, err
	}

	return mb.ExitSuccess, nil
}

func ln(args []string, opts options) error {
	if isHardLinkForDir(args, opts.Symbolic) {
		return errors.New("hard links to directories are not allowed")
	}

	src := os.ExpandEnv(args[0])
	dest := decideDestination(args)
	if err := deleteFileIfNeeded(dest, opts.Force); err != nil {
		return err
	}

	if opts.Symbolic {
		if err := os.Symlink(src, dest); err != nil {
			return err
		}
	} else {
		if err := os.Link(src, dest); err != nil {
			return err
		}
	}
	return nil
}

func deleteFileIfNeeded(path string, force bool) error {
	if mb.Exists(path) && !force {
		return errors.New(path + " already exists")
	}

	if mb.Exists(path) && !mb.IsFile(path) {
		return errors.New("hard links to directories are not allowed")
	}

	if force {
		if err := os.Remove(path); err != nil {
			return err
		}
	}
	return nil
}

func decideDestination(args []string) string {
	if len(args) == 1 {
		return filepath.Base(args[0])
	}
	return os.ExpandEnv(args[1])
}

func isHardLinkForDir(args []string, symbolic bool) bool {
	return !symbolic && mb.IsDir(os.ExpandEnv(args[0]))
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

	if !isValidArgNr(args) {
		showHelp(p)
		osExit(mb.ExitFailure)
	}
	return args, nil
}

func initParser(opts *options) *flags.Parser {
	parser := flags.NewParser(opts, flags.Default)
	parser.Name = cmdName
	parser.Usage = "[OPTIONS] SOURCE DESTINATION"

	return parser
}

func isValidArgNr(args []string) bool {
	return len(args) == 1 || len(args) == 2
}

func showHelp(p *flags.Parser) {
	p.WriteHelp(os.Stdout)
}
