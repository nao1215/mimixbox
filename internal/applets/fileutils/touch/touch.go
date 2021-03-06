//
// mimixbox/internal/applets/fileutils/touch/touch.go
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
package touch

import (
	"fmt"
	"os"
	"time"

	mb "github.com/nao1215/mimixbox/internal/lib"

	"github.com/jessevdk/go-flags"
)

const cmdName string = "touch"

const version = "1.0.2"

var osExit = os.Exit

type options struct {
	NoCreate bool `short:"c" long:"no-create" description:"Not create file"`
	Version  bool `short:"v" long:"version" description:"Show touch command version"`
}

func Run() (int, error) {
	var opts options
	var args []string
	var err error

	if args, err = parseArgs(&opts); err != nil {
		return mb.ExitFailure, nil
	}

	status := mb.ExitSuccess
	for _, file := range args {
		if err = touch(file, opts); err != nil {
			fmt.Fprintln(os.Stderr, "touch: "+err.Error())
			status = mb.ExitFailure
			continue
		}
	}
	return status, nil
}

// atime = access time
// ctime = change time
// mtime = modify time
func touch(file string, opts options) error {
	path := os.ExpandEnv(file)
	if !mb.Exists(path) && !opts.NoCreate {
		f, err := os.Create(path)
		if err != nil {
			return err
		}
		defer f.Close()
	} else {
		currentTime := time.Now().Local()
		err := os.Chtimes(path, currentTime, currentTime)
		if err != nil {
			return err
		}
	}
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
		osExit(mb.ExitSuccess)
	}

	if !isValidArgNr(args) {
		showHelp(p)
		osExit(mb.ExitFailure)
	}
	return args, nil
}

func initParser(opts *options) *flags.Parser {
	parser := flags.NewParser(opts, flags.Default)
	parser.Name = cmdName
	parser.Usage = "[OPTIONS] FILE_PATH"

	return parser
}

func isValidArgNr(args []string) bool {
	return len(args) >= 1
}

func showHelp(p *flags.Parser) {
	p.WriteHelp(os.Stdout)
}
