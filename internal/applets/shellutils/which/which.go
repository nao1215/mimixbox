//
// mimixbox/internal/applets/shellutils/which/which.go
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
package which

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/jessevdk/go-flags"
	mb "github.com/nao1215/mimixbox/internal/lib"
)

const cmdName string = "which"

const version = "1.0.1"

var osExit = os.Exit

type options struct {
	Version bool `short:"v" long:"version" description:"Show which command version"`
}

// Exit code
const (
	ExitSuccess int = iota // 0
	ExitFailuer
)

func Run() (int, error) {
	var opts options
	var args []string
	var err error

	if args, err = parseArgs(&opts); err != nil {
		return ExitFailuer, nil
	}

	for _, path := range args {
		p, err := exec.LookPath(path)
		if err != nil {
			e, ok := err.(*exec.Error)
			if ok && e.Err == exec.ErrNotFound {
				continue // Ignore NotFound error
			}
			return ExitFailuer, err
		}
		fmt.Fprintln(os.Stdout, p)
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

	if !isValidArgNr(args) {
		showHelp(p)
		osExit(ExitFailuer)
	}
	return args, nil
}

func initParser(opts *options) *flags.Parser {
	parser := flags.NewParser(opts, flags.Default)
	parser.Name = cmdName
	parser.Usage = "[OPTIONS] COMMAN_NAME"

	return parser
}

func isValidArgNr(args []string) bool {
	return len(args) == 1
}

func showHelp(p *flags.Parser) {
	p.WriteHelp(os.Stdout)
}
