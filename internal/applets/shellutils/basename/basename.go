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
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jessevdk/go-flags"
	mb "github.com/nao1215/mimixbox/internal/lib"
)

const cmdName string = "basename"

const version = "1.0.2"

var osExit = os.Exit

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
		return mb.ExitFailure, nil
	}

	for _, path := range args {
		basename := filepath.Base(path)
		if opts.Suffix != "" && strings.HasSuffix(basename, opts.Suffix) {
			basename = strings.TrimSuffix(basename, opts.Suffix)
		}

		// The result when the user specifies ""(empty string) is different from Coreutils.
		// So change the result from "." to "" to match the result with Coreutils.
		if path == "" && basename == "." {
			basename = ""
		}

		if opts.Zero {
			fmt.Fprintf(os.Stdout, "%s", basename)
		} else {
			fmt.Fprintln(os.Stdout, basename)
		}

		if !opts.Multiple {
			break
		}
	}
	return mb.ExitSuccess, nil
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
		fmt.Fprintln(os.Stderr, "basename: no operand")
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
