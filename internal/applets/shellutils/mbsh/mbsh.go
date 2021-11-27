//
// mimixbox/internal/applets/shellutils/mbsh/sh.go
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
package mbsh

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/jessevdk/go-flags"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/mbsh/builtin"
	mb "github.com/nao1215/mimixbox/internal/lib"
)

const cmdName string = "mbsh"

var osExit = os.Exit

// ErrNoPath is returned when 'cd' was called without a second argument.
var ErrNoPath = errors.New("path required")

const version = "0.0.2"

const (
	ExitSuccess int = iota // 0
	ExitFailuer
)

type options struct {
	Version bool `short:"v" long:"version" description:"Show shell version"`
}

func Run() (int, error) {
	args, opts := parseArgs()

	fmt.Fprintf(os.Stdout, "Dummy(Not implement shell option): %s:%v\n", args, opts.Version)

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		// Read the keyboad input.
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}

		// Handle the execution of the input.
		if err := execInput(input); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}
}

func execInput(input string) error {
	input = strings.TrimSuffix(input, "\n")
	args := strings.Split(input, " ")

	// ユーザ入力がビルトインコマンドであれば、優先的に実行する。
	if builtin.IsBuiltinCmd(args[0]) {
		return builtin.Run(args[0], args[1:])
	}

	// Prepare the command to execute.
	cmd := exec.Command(args[0], args[1:]...)

	// Set the correct output device.
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	// Execute the command and return the error.
	return cmd.Run()
}

// オプション解析
func parseArgs() ([]string, options) {
	var opts options
	p := initParser(&opts)

	args, err := p.Parse()
	if err != nil {
		osExit(ExitFailuer)
	}

	if opts.Version {
		mb.ShowVersion(cmdName, version)
		osExit(ExitSuccess)
	}

	return args, opts
}

// オプション解析用のパーサを初期化する。
func initParser(opts *options) *flags.Parser {
	parser := flags.NewParser(opts, flags.Default)
	parser.Name = cmdName
	parser.Usage = "[OPTIONS]"

	return parser
}
