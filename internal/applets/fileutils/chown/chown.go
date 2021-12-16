//
// mimixbox/internal/applets/fileutils/chown/chown.go
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
package chown

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	mb "github.com/nao1215/mimixbox/internal/lib"

	"github.com/jessevdk/go-flags"
)

const cmdName string = "chown"

const version = "1.0.0"

var osExit = os.Exit

// Exit code
const (
	ExitSuccess int = iota // 0
	ExitFailuer
)

type idInfo struct {
	owner string
	group string
}

type options struct {
	Recursive bool `short:"R" long:"recursive" description:"Change file owner and/or group IDs recursively"`
	Version   bool `short:"v" long:"version" description:"Show chown command version"`
}

func Run() (int, error) {
	var opts options
	var args []string
	var err error

	if args, err = parseArgs(&opts); err != nil {
		return ExitFailuer, nil
	}

	return chown(ids(args[0]), args[1:], opts)
}

// ids is "owner:group" or "owner" or something.
func ids(ids string) idInfo {
	idInfo := idInfo{"", ""}
	if strings.Contains(ids, ":") {
		strList := strings.Split(ids, ":")
		idInfo.owner = strList[0]
		idInfo.group = strList[1]
	} else {
		idInfo.owner = ids
	}
	return idInfo
}

func chown(ids idInfo, files []string, opts options) (int, error) {
	owner, err := mb.LookupUid(ids.owner)
	if err != nil {
		return ExitFailuer, err
	}

	var gid int = -1
	if ids.group != "" {
		gid, err = mb.LookupGid(ids.group)
		if err != nil {
			return ExitFailuer, err
		}
	}

	status := ExitSuccess
	for _, path := range files {
		path = os.ExpandEnv(path)

		if ids.group == "" {
			var st syscall.Stat_t
			if err := syscall.Stat(path, &st); err != nil {
				status = ExitFailuer
				fmt.Fprintln(os.Stderr, cmdName+": "+path+": "+err.Error())
				continue
			}
			gid = int(st.Gid)
		}

		if opts.Recursive {
			if err := changeOwnerRecursive(path, owner, gid); err != nil {
				status = ExitFailuer
				fmt.Fprintln(os.Stderr, cmdName+": "+path+": "+err.Error())
				continue
			}
		} else {
			if err := os.Chown(path, owner, gid); err != nil {
				status = ExitFailuer
				fmt.Fprintln(os.Stderr, cmdName+": "+path+": "+err.Error())
				continue
			}
		}
	}
	return status, nil
}

func changeOwnerRecursive(path string, uid int, gid int) error {
	err := filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if err := os.Chown(p, uid, gid); err != nil {
			return err
		}
		return nil
	})
	return err
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
		if len(args) == 0 {
			fmt.Fprintln(os.Stderr, cmdName+": no operand")
		} else if len(args) == 1 {
			fmt.Fprintln(os.Stderr, cmdName+": no operand after "+args[0])
		}
		osExit(ExitFailuer)
	}
	return args, nil
}

func initParser(opts *options) *flags.Parser {
	parser := flags.NewParser(opts, flags.Default)
	parser.Name = cmdName
	parser.Usage = "[OPTIONS] [OWNER][:GROUP] FILES"

	return parser
}

func isValidArgNr(args []string) bool {
	return len(args) >= 2
}
