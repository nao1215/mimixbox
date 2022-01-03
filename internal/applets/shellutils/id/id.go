//
// mimixbox/internal/applets/shellutils/id/id.go
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
package id

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"strings"

	"github.com/jessevdk/go-flags"
	mb "github.com/nao1215/mimixbox/internal/lib"
)

const cmdName string = "id"
const version = "1.0.1"

var osExit = os.Exit

type options struct {
	Group    bool `short:"g" long:"group" description:"Print only the effective group ID"`
	AllGroup bool `short:"G" long:"groups" description:"Print all group IDs"`
	Name     bool `short:"n" long:"name" description:"Print the name instead of a number (for -ugG)"`
	User     bool `short:"u" long:"user" description:"Print only the effective user ID"`
	Version  bool `short:"v" long:"version" description:"Show id command version"`
}

// Exit code
const (
	ExitSuccess int = iota // 0
	ExitFailuer
)

func Run() (int, error) {
	var opts options
	var err error

	if _, err = parseArgs(&opts); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return ExitFailuer, nil
	}
	return id(opts)
}

func id(opts options) (int, error) {
	user, err := user.Current()
	if err != nil {
		return ExitFailuer, err
	}

	groups, err := mb.Groups(user.Username)
	if err != nil {
		return ExitFailuer, err
	}

	switch {
	case opts.Group:
		return dumpGid(*user, opts.Name)
	case opts.AllGroup:
		mb.DumpGroups(groups, opts.Name)
		return ExitSuccess, nil
	case opts.User:
		return dumpUid(*user, opts.Name)
	default:
		if err := dumpAllId(*user, groups); err != nil {
			return ExitFailuer, err
		}
	}

	return ExitSuccess, nil
}

func dumpUid(u user.User, showName bool) (int, error) {
	var err error

	if showName {
		_, err = fmt.Fprintln(os.Stdout, u.Username)
	} else {
		_, err = fmt.Fprintln(os.Stdout, u.Uid)
	}

	if err != nil {
		return ExitFailuer, err
	}
	return ExitSuccess, err
}

func dumpGid(u user.User, showName bool) (int, error) {
	if showName {
		g, err := user.LookupGroup(u.Username)
		if err != nil {
			return ExitFailuer, err
		}
		fmt.Fprintln(os.Stdout, g.Name)
	} else {
		fmt.Fprintln(os.Stdout, u.Gid)
	}
	return ExitSuccess, nil
}

func dumpAllId(u user.User, groups []user.Group) error {
	var resultLine string = ""

	g, err := user.LookupGroupId(u.Gid)
	if err != nil {
		return err
	}
	resultLine = "uid=" + u.Uid + "(" + u.Username + ") "
	resultLine = resultLine + "gid=" + u.Gid + "(" + g.Name + ") "
	resultLine = resultLine + "groups="

	for _, v := range groups {
		resultLine = resultLine + v.Gid + "(" + v.Name + "),"
	}
	fmt.Fprintln(os.Stdout, strings.TrimSuffix(resultLine, ","))
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

	if !validSpecifiedSameTime(*opts) {
		return nil, errors.New("-g, -u, -G option cannot be specified at the same time")
	}

	if !validNameOpt(*opts) {
		return nil, errors.New("specify the -n option at the same time as -g, -u, -G")
	}

	return args, nil
}

func validSpecifiedSameTime(opts options) bool {
	return countShowOnlyOneItemOpts(opts) <= 1
}

func validNameOpt(opts options) bool {
	if opts.Name {
		return countShowOnlyOneItemOpts(opts) == 1
	}
	return true
}

func countShowOnlyOneItemOpts(opts options) int {
	var count int = 0
	if opts.AllGroup {
		count++
	}
	if opts.Group {
		count++
	}
	if opts.User {
		count++
	}
	return count
}

func initParser(opts *options) *flags.Parser {
	parser := flags.NewParser(opts, flags.Default)
	parser.Name = cmdName
	parser.Usage = "[OPTIONS]"

	return parser
}
