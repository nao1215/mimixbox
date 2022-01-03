//
// mimixbox/internal/applets/fileutils/chgrp/chgrp.go
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
package chgrp

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	mb "github.com/nao1215/mimixbox/internal/lib"

	"github.com/jessevdk/go-flags"
)

const cmdName string = "chgrp"

const version = "1.0.0"

var osExit = os.Exit

type groupInfo struct {
	group string
	files []string
}

type options struct {
	Recursive bool `short:"R" long:"recursive" description:"Change file group IDs recursively"`
	Version   bool `short:"v" long:"version" description:"Show chgrp command version"`
}

func Run() (int, error) {
	var opts options
	var args []string
	var err error

	if args, err = parseArgs(&opts); err != nil {
		return mb.ExitFailure, nil
	}

	groupInfo := groupInfo{args[0], args[1:]}
	return chgrp(groupInfo, opts)
}

func chgrp(gInfo groupInfo, opts options) (int, error) {
	gid, err := mb.LookupGid(gInfo.group)
	if err != nil {
		return mb.ExitFailure, err
	}

	status := mb.ExitSuccess
	for _, path := range gInfo.files {
		path = os.ExpandEnv(path)
		if opts.Recursive {
			if err := changeGroupRecursive(path, gid); err != nil {
				status = mb.ExitFailure
				fmt.Fprintln(os.Stderr, cmdName+": "+path+": "+err.Error())
				continue
			}
		} else {
			if err := changeGroup(path, gid); err != nil {
				status = mb.ExitFailure
				fmt.Fprintln(os.Stderr, cmdName+": "+path+": "+err.Error())
				continue
			}
		}

	}
	return status, nil
}

func changeGroupRecursive(path string, gid int) error {
	err := filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if err := changeGroup(p, gid); err != nil {
			return err
		}
		return nil
	})
	return err
}

func changeGroup(path string, gid int) error {
	var st syscall.Stat_t
	if err := syscall.Stat(path, &st); err != nil {
		return err
	}
	return os.Chown(path, int(st.Uid), gid)
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
		if len(args) == 0 {
			fmt.Fprintln(os.Stderr, cmdName+": no operand")
		} else if len(args) == 1 {
			fmt.Fprintln(os.Stderr, cmdName+": no operand after "+args[0])
		}
		osExit(mb.ExitFailure)
	}
	return args, nil
}

func initParser(opts *options) *flags.Parser {
	parser := flags.NewParser(opts, flags.Default)
	parser.Name = cmdName
	parser.Usage = "[OPTIONS] GROUP FILES"

	return parser
}

func isValidArgNr(args []string) bool {
	return len(args) >= 2
}
