//
// mimixbox/internal/applets/textutils/tac/tac.go
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
package tac

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	mb "github.com/nao1215/mimixbox/internal/lib"

	"github.com/jessevdk/go-flags"
)

const cmdName string = "tac"

const version = "1.0.4"

var osExit = os.Exit

// Exit code
const (
	ExitSuccess int = iota // 0
	ExitFailuer
)

type options struct {
	Version bool `short:"v" long:"version" description:"Show cat command version"`
}

func Run() (int, error) {
	var opts options
	var args []string
	var err error

	if args, err = parseArgs(&opts); err != nil {
		return ExitFailuer, nil
	}

	if mb.HasPipeData() {
		printFromTail(strings.Split(args[0], "\n"))
		return ExitSuccess, nil
	}

	if len(args) == 0 || mb.Contains(args, "-") {
		tacUserInput()
		return ExitSuccess, nil
	}

	for _, file := range args {
		err := tac(file)
		if err != nil {
			return ExitFailuer, err
		}
	}

	return ExitSuccess, nil
}

func tac(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	lines := []string{}
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	printFromTail(lines)
	return nil
}

func printFromTail(lines []string) {
	for i := range lines {
		fmt.Fprintln(os.Stdout, lines[len(lines)-i-1])
	}
}

func tacUserInput() {
	var inputs []string
	for {
		input, next := mb.Input()
		if !next {
			break
		}
		inputs = append(inputs, input)
	}
	for i := range inputs {
		fmt.Fprintln(os.Stdout, inputs[len(inputs)-i-1])
	}
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
