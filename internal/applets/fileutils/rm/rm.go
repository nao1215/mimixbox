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
	mb "mimixbox/internal/lib"
	"os"
	"sort"

	"github.com/jessevdk/go-flags"
)

const cmdName string = "rm"

const version = "1.0.0"

var osExit = os.Exit

// Exit code
const (
	ExitSuccess int = iota // 0
	ExitFailuer
)

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
	var status int

	if args, err = parseArgs(&opts); err != nil {
		return ExitFailuer, nil
	}

	// Coremb will continue to delete files as much as possible.
	// MimixBox stops processing if an error occurs even once.
	for _, path := range args {
		if status, err := rm(path, opts); err != nil {
			return status, err
		}
	}

	return status, nil
}

func rm(path string, opts options) (int, error) {
	if status, err := validBeforeRemove(path, opts); status != ExitSuccess {
		return status, err
	}

	if mb.IsFile(path) {
		if opts.Interactive && !mb.Question("Remove "+path+"?") {
			return ExitSuccess, nil // Skip this file
		}
		if err := os.Remove(path); err != nil {
			return ExitFailuer, err
		}
		return ExitSuccess, nil
	}

	if err := removeDir(path, opts.Interactive); err != nil {
		return ExitFailuer, err
	}

	return ExitSuccess, nil
}

func removeDir(dir string, interactive bool) error {
	if !interactive {
		if err := os.RemoveAll(dir); err != nil {
			return err
		}
		return nil
	}
	if err := interactiveRemoveDir(dir); err != nil {
		return err
	}
	return nil
}

func interactiveRemoveDir(dir string) error {
	dirs, files, err := mb.Walk(dir)
	if err != nil {
		return err
	}

	// Start with the deepest directory or file
	sort.Sort(sort.Reverse(sort.StringSlice(dirs)))
	sort.Sort(sort.Reverse(sort.StringSlice(files)))

	for _, file := range files {
		if !mb.Question("Remove " + file + "?") {
			continue
		}
		err := os.Remove(file)
		if err != nil {
			return err
		}
	}
	for _, dir := range dirs {
		if !mb.Question("Remove " + dir + "?") {
			continue
		}
		err := os.Remove(dir)
		if err != nil {
			return err
		}
	}
	return nil
}

func validBeforeRemove(path string, opts options) (int, error) {
	if mb.IsRootDir(path) && !opts.NoPreserve {
		return ExitFailuer, errors.New("do not remove the root directory")
	}

	if !mb.Exists(path) {
		if !opts.Force {
			return ExitFailuer, errors.New("can't remove " + path + ": No such file or directory exists")
		}
		return ExitFailuer, nil
	}

	if mb.IsDir(path) && !opts.Recursive {
		return ExitFailuer, errors.New("can't remove " + path + ": It's directory")
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
		showVersion()
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

func showVersion() {
	description := cmdName + " version " + version + " (under Apache License verison 2.0)\n"
	fmt.Print(description)
}

func showHelp(p *flags.Parser) {
	p.WriteHelp(os.Stdout)
}
