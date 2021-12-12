//
// mimixbox/internal/applets/textutils/unix2dos/unix2dos.go
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
package unix2dos

import (
	"fmt"
	"os"
	"strings"

	mb "github.com/nao1215/mimixbox/internal/lib"

	"github.com/jessevdk/go-flags"
)

const cmdName string = "unix2dos"
const version = "1.0.5"

var osExit = os.Exit

type options struct {
	Version bool `short:"v" long:"version" description:"Show unix2dos command version"`
}

// Exit code
const (
	ExitSuccess int = iota // 0
	ExitFailuer
	ExitNoSuchFile
)

func Run() (int, error) {
	var opts options
	var args []string
	var err error

	if args, err = parseArgs(&opts); err != nil {
		return ExitFailuer, nil
	}

	if mb.HasPipeData() {
		fmt.Fprint(os.Stdout, toCRLF(strings.Split(args[0], "")))
		return ExitSuccess, nil
	}

	if len(args) == 0 || mb.Contains(args, "-") {
		mb.Parrot(false)
		return ExitSuccess, nil
	}

	return unix2dos(args)
}

func unix2dos(files []string) (int, error) {
	status := ExitSuccess
	for _, file := range files {
		target := os.ExpandEnv(file)
		if !mb.IsFile(target) {
			fmt.Fprintln(os.Stderr, cmdName+": skip "+target+": not regular file")
			status = ExitNoSuchFile
			continue
		}

		lines, err := mb.ReadFileToStrList(target)
		if err != nil {
			fmt.Fprintln(os.Stderr, cmdName+": "+target+": Can't read file and convert LF to CRLF")
			status = ExitFailuer
			continue
		}
		fmt.Fprintln(os.Stdout, cmdName+": converting file "+target+" to DOS format...")
		lines = toCRLF(lines)
		if err := mb.ListToFile(target, lines); err != nil {
			fmt.Fprintln(os.Stderr, err)
			status = ExitFailuer
			continue
		}
	}
	return status, nil
}

func toCRLF(unixStr []string) []string {
	var replaceStr []string
	for _, v := range unixStr {
		if strings.HasSuffix(v, "\r\n") {
			replaceStr = append(replaceStr, v)
		} else {
			replaceStr = append(replaceStr, strings.Replace(v, "\n", "\r\n", -1))
		}
	}
	return replaceStr
}

func parseArgs(opts *options) ([]string, error) {
	p := initParser(opts)

	args, err := p.Parse()
	if err != nil {
		return nil, err
	}

	if mb.HasPipeData() {
		stdin, err := mb.FromPIPE()
		if err != nil {
			return nil, err
		}
		return []string{stdin}, nil
	}

	if opts.Version {
		mb.ShowVersion(cmdName, version)
		osExit(ExitSuccess)
	}

	return args, nil
}

func initParser(opts *options) *flags.Parser {
	parser := flags.NewParser(opts, flags.Default)
	parser.Name = cmdName
	parser.Usage = "[OPTIONS] FILE_PATH"

	return parser
}
