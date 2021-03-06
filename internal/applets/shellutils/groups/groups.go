//
// mimixbox/internal/applets/shellutils/groups/groups.go
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
package groups

import (
	"fmt"
	"os"
	"os/user"

	"github.com/jessevdk/go-flags"
	mb "github.com/nao1215/mimixbox/internal/lib"
)

const cmdName string = "groups"
const version = "1.0.2"

var osExit = os.Exit

type options struct {
	Version bool `short:"v" long:"version" description:"Show groups command version"`
}

func Run() (int, error) {
	var opts options
	var err error
	var args []string

	if args, err = parseArgs(&opts); err != nil {
		return mb.ExitSuccess, nil
	}
	return groups(args)
}

func groups(args []string) (int, error) {

	if len(args) == 0 {
		return showCurrentUserGroups()
	}

	var status int = mb.ExitSuccess
	for _, uname := range args {
		groups, err := mb.Groups(uname)
		if err != nil {
			fmt.Fprintln(os.Stderr, "can't get "+uname+" groups information")
			status = mb.ExitFailure
			continue
		}
		fmt.Fprint(os.Stdout, uname+" : ")
		mb.DumpGroups(groups, true)
	}
	return status, nil
}

func showCurrentUserGroups() (int, error) {
	u, err := user.Current()
	if err != nil {
		return mb.ExitFailure, err
	}

	groups, err := mb.Groups(u.Username)
	if err != nil {
		return mb.ExitFailure, err
	}
	mb.DumpGroups(groups, true)
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

	return args, nil
}

func initParser(opts *options) *flags.Parser {
	parser := flags.NewParser(opts, flags.Default)
	parser.Name = cmdName
	parser.Usage = "[OPTIONS] [USERNAME]"

	return parser
}
