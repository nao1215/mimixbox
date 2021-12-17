//
// mimixbox/internal/applets/shellutils/pwd/pwd.go
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
package pwd

import (
	"fmt"
	"os"
	"path/filepath"

	mb "github.com/nao1215/mimixbox/internal/lib"

	"github.com/jessevdk/go-flags"
)

const cmdName string = "pwd"

const version = "1.0.0"

var osExit = os.Exit

// Exit code
const (
	ExitSuccess int = iota // 0
	ExitFailuer
)

type options struct {
	Logical  bool `short:"L"  description:"print the value of $PWD if it names the current working directory (default)"`
	Physical bool `short:"P"  description:"print the physical directory, without any symbolic links"`
	Version  bool `short:"v" long:"version" description:"Show pwd command version"`
}

func Run() (int, error) {
	var opts options
	var err error

	if _, err = parseArgs(&opts); err != nil {
		return ExitFailuer, nil
	}
	return pwd(opts)
}

func pwd(opts options) (int, error) {
	if !opts.Logical && !opts.Physical {
		fmt.Fprintln(os.Stdout, os.Getenv("PWD"))
	} else if opts.Logical && opts.Physical {
		fmt.Fprintln(os.Stdout, os.Getenv("PWD"))
	} else if opts.Logical {
		fmt.Fprintln(os.Stdout, os.Getenv("PWD"))
	} else if opts.Physical {
		path, err := filepath.EvalSymlinks(os.Getenv("PWD"))
		if err != nil {
			return ExitFailuer, err
		}
		fmt.Fprintln(os.Stdout, path)
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

	return args, nil
}

func initParser(opts *options) *flags.Parser {
	parser := flags.NewParser(opts, flags.Default)
	parser.Name = cmdName
	parser.Usage = "[OPTIONS]"

	return parser
}
