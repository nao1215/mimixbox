//
// mimixbox/internal/applets/fileutils/cp/cp.go
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
package cp

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	mb "github.com/nao1215/mimixbox/internal/lib"

	"github.com/jessevdk/go-flags"
)

const cmdName string = "cp"

const version = "1.0.0"

var osExit = os.Exit

// Exit code
const (
	ExitSuccess int = iota // 0
	ExitFailuer
)

type options struct {
	Force       bool `short:"f" long:"force" description:"If file exists, forcibly overwrite it"`
	Interactive bool `short:"i" long:"interactive" description:"Ask every time if you want to remove"`
	Recursive   bool `short:"r" long:"recursive" description:"Recursively copy directories"`
	Version     bool `short:"v" long:"version" description:"Show cp command version"`
}

func Run() (int, error) {
	var opts options
	var args []string
	var err error

	if args, err = parseArgs(&opts); err != nil {
		return ExitFailuer, nil
	}

	err = cp(args, opts)
	if err != nil {
		return ExitFailuer, err
	}

	return ExitSuccess, nil
}

func cp(files []string, opts options) error {
	dest := files[len(files)-1]

	for _, src := range files[:len(files)-1] {
		if !mb.Exists(src) {
			return errors.New(src + " does not exist")
		}

		if mb.IsSamePath(src, dest) {
			return errors.New(src + " and " + dest + " is same.")
		}

		if mb.IsFile(src) {
			if err := cpFile(src, dest, opts); err != nil {
				return err
			}
		} else {
			if err := cpDir(src, dest, opts); err != nil {
				return err
			}
		}
	}
	return nil
}

func cpFile(src string, dest string, opts options) error {
	if mb.IsFile(dest) && mb.IsSameFileName(src, dest) && opts.Interactive {
		if !mb.Question("Overwrite " + dest) {
			return nil // Skip this file
		}
	}
	return mb.Copy(src, dest)
}

func cpDir(src string, dest string, opts options) error {
	if !opts.Recursive {
		return errors.New("--recursive is not specified: omitting directory: " + src)
	}

	srcDirs, srcFiles, err := mb.Walk(src)
	if err != nil {
		return err
	}

	for _, dir := range srcDirs {
		// dirs has a tree structure path from the source directory.
		// Change the top directory name of this path to the destination name.
		// Bad implementation.
		dir = strings.Replace(dir, mb.TopDirName(dir), filepath.Base(dest), 1)
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}
	}

	for _, src := range srcFiles {
		destFile := strings.Replace(src, mb.TopDirName(src), filepath.Base(dest), 1)
		err := mb.Copy(src, destFile)
		if err != nil {
			return err
		}
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
	parser.Usage = "[OPTIONS] SOURCE DESTINATION"

	return parser
}

func isValidArgNr(args []string) bool {
	return len(args) >= 2
}

func showVersion() {
	description := cmdName + " version " + version + " (under Apache License verison 2.0)\n"
	fmt.Print(description)
}

func showHelp(p *flags.Parser) {
	p.WriteHelp(os.Stdout)
}
