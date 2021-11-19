//
// mimixbox/internal/applets/shellutils/basename/basename.go
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
package basename

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jessevdk/go-flags"
)

const cmdName string = "basename"

const version = "1.0.0"

var osExit = os.Exit

// Exit code
const (
	ExitSuccess int = iota // 0
	ExitFailuer
)

type options struct {
	Multiple bool   `short:"a" long:"multiple" description:"Process for multiple PATHs"`
	Suffix   string `short:"s" long:"suffix" default:"" description:"Delete suffix"`
	Zero     bool   `short:"z" long:"zero" description:"Basename do not include line feed"`
	Version  bool   `short:"v" long:"version" description:"Show basename command version"`
}

func Run() (int, error) {
	var opts options
	var args []string
	var err error

	if args, err = parseArgs(&opts); err != nil {
		return ExitFailuer, nil
	}

	// Different from Coreutils, MimixBox does not allow suffix
	// specification without -s option.
	if !opts.Multiple && len(args) >= 2 {
		return ExitFailuer, errors.New("multiple PATHs specified (use --multiple)")
	}

	for _, path := range args {
		basename := filepath.Base(path)
		if opts.Suffix != "" && strings.HasSuffix(basename, opts.Suffix) {
			basename = strings.TrimRight(basename, opts.Suffix)
		}

		if opts.Zero {
			fmt.Printf("%s", basename)
		} else {
			fmt.Println(basename)
		}
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