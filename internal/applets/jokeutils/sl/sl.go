//
// mimixbox/internal/applets/shellutils/sl/cowsay.go
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
package sl

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jessevdk/go-flags"
	mb "github.com/nao1215/mimixbox/internal/lib"
	"golang.org/x/term"
)

const cmdName string = "sl"

const version = "0.9.0"

var osExit = os.Exit

type options struct {
	Version bool `short:"v" long:"version" description:"Show sl command version"`
}

var slAA = [][][]string{
	{
		{
			"                 (   )",
			"             (@@@@)",
			"          (    )",
			"",
			"        (@@@)",
		},
		{
			"                 (@@@)",
			"             (    )",
			"          (@@@@)",
			"",
			"        (   )",
		},
		{
			"                 (   )",
			"              @@@",
			"          (    )",
			"          @@",
			"        (   )",
		},
	},
	{
		{
			"      ====        ________                ___________ ",
			"  _D _|  |_______/        \\__I_I_____===__|_________| ",
			"   |(_)---  |   H\\________/ |   |        =|___ ___|      _________________         ",
			"   /     |  |   H  |  |     |   |         ||_| |_||     _|                \\_____A  ",
			"  |      |  |   H  |__--------------------| [___] |   =|                        |  ",
			"  | ________|___H__/__|_____/[ ][ ]\\_______|       |   -|      ʕ ◔ ϖ ◔ ʔ      |  ",
			"  |/ |   |-----------I_____I [ ][ ] [ ]  D   |=======|____|_______________________|_ ",
		},
	},
	{
		{
			"__/ =| o |=-O=====O=====O=====O \\ ____Y___________|__|__________________________|_ ",
			" |/-=|___|=    ||    ||    ||    |_____/~\\___/          |_D__D__D_|  |_D__D__D_|   ",
			"  \\_/      \\__/  \\__/  \\__/  \\__/      \\_/               \\_/   \\_/    \\_/   \\_/    ",
		},
		{
			"__/ =| o |=-~~\\  /~~\\  /~~\\  /~~\\ ____Y___________|__|__________________________|_ ",
			" |/-=|___|=O=====O=====O=====O   |_____/~\\___/          |_D__D__D_|  |_D__D__D_|   ",
			"  \\_/      \\__/  \\__/  \\__/  \\__/      \\_/               \\_/   \\_/    \\_/   \\_/    ",
		},
		{
			"__/ =| o |=-~~\\  /~~\\  /~~\\  /~~\\ ____Y___________|__|__________________________|_ ",
			" |/-=|___|=    ||    ||    ||    |_____/~\\___/          |_D__D__D_|  |_D__D__D_|   ",
			"  \\_/      \\O=====O=====O=====O_/      \\_/               \\_/   \\_/    \\_/   \\_/    ",
		},
		{
			"__/ =| o |=-~~\\  /~~\\  /~~\\  /~~\\ ____Y___________|__|__________________________|_ ",
			" |/-=|___|=    ||    ||    ||    |_____/~\\___/          |_D__D__D_|  |_D__D__D_|   ",
			"  \\_/      \\_O=====O=====O=====O/      \\_/               \\_/   \\_/    \\_/   \\_/    ",
		},
		{
			"__/ =| o |=-~~\\  /~~\\  /~~\\  /~~\\ ____Y___________|__|__________________________|_ ",
			" |/-=|___|=   O=====O=====O=====O|_____/~\\___/          |_D__D__D_|  |_D__D__D_|   ",
			"  \\_/      \\__/  \\__/  \\__/  \\__/      \\_/               \\_/   \\_/    \\_/   \\_/    ",
		},
		{
			"__/ =| o |=-~O=====O=====O=====O\\ ____Y___________|__|__________________________|_ ",
			" |/-=|___|=    ||    ||    ||    |_____/~\\___/          |_D__D__D_|  |_D__D__D_|   ",
			"  \\_/      \\__/  \\__/  \\__/  \\__/      \\_/               \\_/   \\_/    \\_/   \\_/    ",
		},
	},
}

func Run() (int, error) {
	var opts options

	_, err := parseArgs(&opts)
	if err != nil {
		return mb.ExitFailure, nil
	}
	return sl()
}

func sl() (int, error) {
	width, _, err := term.GetSize(0)
	if err != nil {
		return mb.ExitFailure, err
	}

	if width < 80 {
		return mb.ExitFailure, errors.New("terminal width is too small")
	}
	for _, ll := range slAA {
		for _, l := range ll {
			for n := range l {
				l[n] = strings.Repeat(" ", width/5) + l[n]
			}
		}
	}

	x := 0
	for {
		fmt.Print("\x0c")
		for _, pp := range slAA {
			for _, l := range pp[x%len(pp)] {
				if x < len(l) {
					fmt.Println(string(l[x:]))
				} else {
					fmt.Println()
				}

			}
		}
		time.Sleep(time.Second / 30)
		if x++; x > width {
			break
		}
	}
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
	parser.Usage = "[OPTIONS]"

	return parser
}
