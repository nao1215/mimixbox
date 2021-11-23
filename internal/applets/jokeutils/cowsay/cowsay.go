//
// mimixbox/internal/applets/shellutils/cowsay/cowsay.go
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
package cowsay

import (
	"fmt"
	"os"
	"strings"

	mb "github.com/nao1215/mimixbox/internal/lib"
)

const cmdName string = "cowsay"

const version = "0.9.1"

var osExit = os.Exit

const cow = `   \ 
    \   ^__^
     \  (oo)\_______
        (__)\       )\/\
            ||----w |
            ||     ||`

// Exit code
const (
	ExitSuccess int = iota // 0
	ExitFailuer
)

func Run() (int, error) {
	var messages string
	args, err := parseArgs(os.Args)
	if err != nil {
		return ExitFailuer, nil
	}

	if mb.HasPipeData() {
		messages = strings.TrimRight(strings.Join(args, ""), "\n")
	} else if len(args) == 0 {
		messages = userInput()
	} else {
		messages = strings.Join(args, "")
	}
	cowsay(messages)

	return ExitSuccess, nil
}

func cowsay(msg string) {
	fmt.Println("------------------------------------------------------------")
	fmt.Printf("%s\n", mb.WrapString(msg, 60))
	fmt.Println("------------------------------------------------------------")
	fmt.Println(cow)
}

func userInput() string {
	var inputs string
	for {
		input, next := mb.Input()
		if !next {
			break
		}
		inputs += input
	}
	return inputs
}

func parseArgs(args []string) ([]string, error) {

	if mb.HasVersionOpt(args) {
		showVersion()
		osExit(ExitSuccess)
	}

	if mb.HasHelpOpt(args) {
		showHelp()
		osExit(ExitSuccess)
	}

	if mb.HasPipeData() {
		stdin, err := mb.FromPIPE()
		if err != nil {
			return nil, err
		}
		return []string{stdin}, nil
	}

	return args[1:], nil
}

func showVersion() {
	description := cmdName + " version " + version + " (under Apache License verison 2.0)\n"
	fmt.Print(description)
}

func showHelp() {
	fmt.Println("Usage:")
	fmt.Println("  cowsay [OPTIONS] message")
	fmt.Println("")
	fmt.Println("Application Options:")
	fmt.Println("  -v, --version       Show cowsay command version")
	fmt.Println("")
	fmt.Println("Help Options:")
	fmt.Println("  -h, --help          Show this help message")
}
