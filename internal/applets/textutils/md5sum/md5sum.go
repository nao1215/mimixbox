//
// mimixbox/internal/applets/textutils/md5sum/md5sum.go
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
package md5sum

import (
	"crypto/md5"
	"os"
	"strings"

	mb "github.com/nao1215/mimixbox/internal/lib"

	"github.com/jessevdk/go-flags"
)

const cmdName string = "md5sum"

const version = "1.0.1"

var osExit = os.Exit

type options struct {
	Check   bool `short:"c" long:"check" description:"Check if the MD5 hash value matches the file"`
	Version bool `short:"v" long:"version" description:"Show md5sum command version"`
}

func Run() (int, error) {
	var opts options
	var args []string
	var err error = nil
	hash := md5.New()

	if args, err = parseArgs(&opts); err != nil {
		return mb.ExitFailure, nil
	}

	if mb.HasPipeData() && mb.HasNoOperand(os.Args, cmdName) {
		err = mb.ChecksumOutput(hash, strings.NewReader(args[0]), "-")
		if err != nil {
			return mb.ExitFailure, err
		}
		return mb.ExitSuccess, nil
	}

	if len(args) == 0 || mb.Contains(args, "-") {
		err = mb.ChecksumOutput(hash, os.Stdin, "-")
		if err != nil {
			return mb.ExitSuccess, nil
		}
		return mb.ExitSuccess, nil
	}

	if opts.Check {
		err = mb.CompareChecksum(hash, args)
		if err != nil {
			return mb.ExitFailure, err
		}
		return mb.ExitSuccess, nil
	}

	return mb.PrintChecksums(cmdName, hash, args)
}

func parseArgs(opts *options) ([]string, error) {
	p := initParser(opts)

	args, err := p.Parse()
	if err != nil {
		return nil, err
	}

	if mb.HasPipeData() && len(args) == 0 {
		stdin, err := mb.FromPIPE()
		if err != nil {
			return nil, err
		}
		return []string{stdin}, nil
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
	parser.Usage = "[OPTIONS] FILE_PATH"

	return parser
}
