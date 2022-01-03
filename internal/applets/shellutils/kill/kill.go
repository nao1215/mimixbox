//
// mimixbox/internal/applets/shellutils/kill/kill.go
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
package kill

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"

	mb "github.com/nao1215/mimixbox/internal/lib"
)

const cmdName string = "kill"
const version = "1.0.0"

var osExit = os.Exit

type options struct {
	nameFlg      bool
	signalName   string
	numberFlg    bool
	signalNumber string
	listFlg      bool
	list         string
	directFlg    bool
	direct       string
}

func Run() (int, error) {
	process, opts := parseArgs(os.Args)

	valid(process, opts)
	if opts.listFlg {
		if opts.list == "" {
			mb.PrintSignalList()
		} else {
			mb.PrintSignal(opts.list)
		}
		osExit(mb.ExitSuccess)
	}

	return kill(process, opts)
}

func kill(process []string, opts options) (int, error) {
	status := mb.ExitSuccess
	for _, v := range process {
		pid, err := strconv.Atoi(v)
		if err != nil {
			fmt.Fprintln(os.Stderr, "kill: "+err.Error()+": "+v)
			status = mb.ExitFailure
			continue
		}

		p, err := os.FindProcess(pid)
		if err != nil {
			fmt.Fprintln(os.Stderr, "kill: "+err.Error()+": "+v)
			status = mb.ExitFailure
			continue
		}

		signal := decideSignal(opts)
		err = p.Signal(syscall.Signal(signal))
		if err != nil {
			fmt.Fprintln(os.Stderr, "kill: "+err.Error())
			status = mb.ExitFailure
			continue
		}
	}
	return status, nil
}

func decideSignal(opts options) int32 {
	var signal int32 = 0
	// TODO: Workaround. I have not investigated the response
	// when multiple signals are specified.
	if opts.directFlg {
		str := strings.TrimLeft(opts.direct, "-")
		if len(str) <= 2 {
			signal = mb.SignalAtoi(str)
		} else {
			signal = mb.ConvSignalNameToNum(str)
		}
	} else if opts.nameFlg {
		signal = mb.ConvSignalNameToNum(opts.signalName)
	} else if opts.numberFlg {
		signal = mb.SignalAtoi(opts.signalNumber)
	} else {
		signal = 9
	}
	return signal
}

func valid(process []string, opts options) {
	if len(process) == 0 && !opts.listFlg {
		showHelp()
		osExit(mb.ExitFailure)
	}
	if opts.nameFlg && !mb.IsSignalName(opts.signalName) {
		fmt.Fprintln(os.Stderr, "kill: -s: invalid signal specification:"+opts.signalName)
		osExit(mb.ExitFailure)
	}
	if opts.numberFlg && !mb.IsSignalName(opts.signalNumber) {
		fmt.Fprintln(os.Stderr, "kill: -n: invalid signal specification:"+opts.signalNumber)
		osExit(mb.ExitFailure)
	}

	trim := strings.TrimLeft(opts.direct, "-")
	if opts.directFlg && !mb.IsSignalName(trim) && !mb.IsSignalNumber(trim) {
		fmt.Fprintln(os.Stderr, "kill: "+opts.direct+": invalid signal specification")
		osExit(mb.ExitFailure)
	}
}

func parseArgs(args []string) ([]string, options) {
	if mb.HasVersionOpt(args) {
		mb.ShowVersion(cmdName, version)
		osExit(mb.ExitSuccess)
	}

	if mb.HasHelpOpt(args) {
		showHelp()
		osExit(mb.ExitSuccess)
	}

	var opts options = options{false, "", false, "", false, "", false, ""}
	args = args[1:]
	for i, v := range args {
		if v == "-s" {
			opts.nameFlg = true
			if len(args) > i+1 {
				opts.signalName = args[i+1]
			} else {
				fmt.Fprintln(os.Stderr, "kill: -s: option requires an argument")
				osExit(mb.ExitFailure)
			}
			continue
		} else if v == "-n" {
			opts.numberFlg = true
			if len(args) > i+1 {
				opts.signalNumber = args[i+1]
			} else {
				fmt.Fprintln(os.Stderr, "kill: -n: option requires an argument")
				osExit(mb.ExitFailure)
			}
			continue
		} else if v == "-l" {
			opts.listFlg = true
			if len(args) > i+1 {
				opts.list = args[i+1]
			}
		} else if strings.Contains(v, "-") {
			opts.directFlg = true
			opts.direct = v
		}
	}
	return decideProcess(args[1:], opts), opts
}

func decideProcess(args []string, opts options) []string {
	new := mb.Remove(args, "-s")
	new = mb.Remove(new, "-n")
	new = mb.Remove(new, "-l")
	new = mb.Remove(new, opts.list)
	new = mb.Remove(new, opts.signalName)
	new = mb.Remove(new, opts.signalNumber)
	new = mb.Remove(new, opts.direct)
	return new
}

func showHelp() {
	fmt.Fprintln(os.Stdout, "Usage:")
	fmt.Fprintln(os.Stdout, "  kill [-s sigspec | -n signum | -sigspec] PID  or")
	fmt.Fprintln(os.Stdout, "  kill -l [PID]")
	fmt.Fprintln(os.Stdout, "")
	fmt.Fprintln(os.Stdout, "Application Options:")
	fmt.Fprintln(os.Stdout, "  -v, --version       Show kill command version")
	fmt.Fprintln(os.Stdout, "")
	fmt.Fprintln(os.Stdout, "Help Options:")
	fmt.Fprintln(os.Stdout, "  -h, --help          Show this help message")
}
