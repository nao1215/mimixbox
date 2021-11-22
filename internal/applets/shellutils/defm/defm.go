//
// mimixbox/internal/applets/shellutils/defm/defm.go
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
package defm // Desktop Entry File Manager

import (
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"
)

const cmdName string = "defm"

const version = "1.0.0"

var osExit = os.Exit

// Exit code
const (
	ExitSuccess int = iota // 0
	ExitFailuer
)

type options struct {
	Version bool `short:"v" long:"version" description:"Show defm command version"`
}

func Run() (int, error) {
	var opts options
	//var args []string
	//var err error

	if _, err := parseArgs(&opts); err != nil {
		return ExitFailuer, nil
	}

	//setLocaleCtype()

	if err := NewAllView().Run(); err != nil {
		return ExitFailuer, err
	}

	return ExitSuccess, nil
}

func setLocaleCtype() {
	if err := os.Setenv("LC_CTYPE", "en_US.UTF-8"); err != nil {
		fmt.Println("Can't change locale. The text may not be displayed correctly")
	}
	fmt.Println(os.Getenv("LC_CTYPE"))
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

	return args, nil
}

func initParser(opts *options) *flags.Parser {
	parser := flags.NewParser(opts, flags.Default)
	parser.Name = cmdName
	parser.Usage = "[OPTIONS]"

	return parser
}

func showVersion() {
	description := cmdName + " version " + version + " (under Apache License verison 2.0)\n"
	fmt.Print(description)
}
