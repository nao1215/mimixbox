//
// mimixbox/internal/applets/shellutils/seq/seq.go
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
package seq

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	mb "github.com/nao1215/mimixbox/internal/lib"
)

const cmdName string = "seq"
const version = "1.0.1"

var osExit = os.Exit

type seqElem struct {
	num int
	err error
}

type seqInfo struct {
	first     seqElem
	increment seqElem
	last      seqElem
}

// Not support arguments containing float numbers
func Run() (int, error) {
	args := parseArgs(os.Args)
	return seq(args)
}

func seq(args []string) (int, error) {
	si, err := parseSeqInfo(args)
	if err != nil {
		return mb.ExitFailure, err
	}

	for i := si.first.num; i <= si.last.num; i = increment(i, si.increment) {
		fmt.Fprintln(os.Stdout, i)
	}
	return mb.ExitSuccess, nil
}

func increment(now int, se seqElem) int {
	return now + se.num
}

func parseSeqInfo(args []string) (seqInfo, error) {
	si := seqInfo{first: seqElem{1, nil},
		increment: seqElem{1, nil},
		last:      seqElem{0, nil}}

	if len(args) == 0 || len(args) > 3 {
		showHelp()
		osExit(mb.ExitSuccess)
	} else if len(args) == 1 {
		setSeqElem(args[0], &si.last)
	} else if len(args) == 2 {
		setSeqElem(args[0], &si.first)
		setSeqElem(args[1], &si.last)
	} else {
		setSeqElem(args[0], &si.first)
		setSeqElem(args[1], &si.increment)
		setSeqElem(args[2], &si.last)
	}

	if err := validSeqInfo(si); err != nil {
		return si, err
	}

	return si, nil
}

func validSeqInfo(si seqInfo) error {
	if si.first.err != nil {
		return si.first.err
	}
	if si.increment.err != nil {
		return si.first.err
	}
	if si.increment.err == nil && si.increment.num == 0 {
		return errors.New("invalid zero increment value")
	}
	if si.last.err != nil {
		return si.first.err
	}
	return nil
}

func setSeqElem(val string, se *seqElem) {
	var err error

	se.num, err = strconv.Atoi(val)
	if err != nil {
		se.err = errors.New("invalid argument: " + val)
		return
	}
	se.err = nil
}

func parseArgs(args []string) []string {
	if mb.HasVersionOpt(args) {
		mb.ShowVersion(cmdName, version)
		osExit(mb.ExitSuccess)
	}

	if mb.HasHelpOpt(args) {
		showHelp()
		osExit(mb.ExitSuccess)
	}

	return args[1:]
}

func showHelp() {
	fmt.Fprintln(os.Stdout, "Usage:")
	fmt.Fprintln(os.Stdout, "  seq [OPTIONS] LAST                  or")
	fmt.Fprintln(os.Stdout, "  seq [OPTIONS] FIRST LAST            or")
	fmt.Fprintln(os.Stdout, "  seq [OPTIONS] FIRST INCREMENT LAS")
	fmt.Fprintln(os.Stdout, "")
	fmt.Fprintln(os.Stdout, "Application Options:")
	fmt.Fprintln(os.Stdout, "  -v, --version       Show seq command version")
	fmt.Fprintln(os.Stdout, "")
	fmt.Fprintln(os.Stdout, "Help Options:")
	fmt.Fprintln(os.Stdout, "  -h, --help          Show this help message")
}
