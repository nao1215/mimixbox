//
// mimixbox/internal/applets/shellutils/sleep/sleep.go
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
package sleep

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	mb "github.com/nao1215/mimixbox/internal/lib"
)

const cmdName string = "sleep"
const version = "1.0.1"

var osExit = os.Exit

type duration struct {
	val  float64
	unit time.Duration
}

func Run() (int, error) {
	var args []string
	var waitTime []duration
	var err error

	args = parseArgs(os.Args)

	if waitTime, err = getWaitTime(args); err != nil {
		return mb.ExitFailure, err
	}

	for _, d := range waitTime {
		time.Sleep(time.Duration(d.val) * d.unit)
	}

	return mb.ExitSuccess, nil
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

	if !isValidArgNr(args) {
		showHelp()
		osExit(mb.ExitSuccess)
	}

	return args[1:]
}

func getWaitTime(input []string) ([]duration, error) {
	var waitTime []duration
	var err error

	for _, v := range input {
		var d duration
		if !hasSuffix(v) {
			if d.val, err = strconv.ParseFloat(v, 64); err != nil {
				return nil, errors.New("Input format error :" + v)
			}
			d.unit = time.Second
		} else {
			strList := strings.Split(v, "")
			lastChar := strList[len(strList)-1]
			onlyNumStr := strings.TrimSuffix(v, lastChar)
			if d.val, err = strconv.ParseFloat(onlyNumStr, 64); err != nil {
				return nil, errors.New("Input format error :" + v)
			}
			d.unit = convToTimeDuration(lastChar)
		}
		waitTime = append(waitTime, d)
	}
	return waitTime, nil
}

func convToTimeDuration(s string) time.Duration {
	switch s {
	case "s":
		return time.Second
	case "m":
		return time.Minute
	case "h":
		return time.Hour
	case "d":
		return (24 * time.Hour)
	default:
		return time.Second
	}
}

func hasSuffix(input string) bool {
	for _, v := range []string{"s", "m", "h", "d"} {
		if strings.HasSuffix(input, v) {
			return true
		}
	}
	return false
}

func isValidArgNr(args []string) bool {
	// 0:sleep, 1:numbers, 2:...
	return len(args) >= 2
}

func showHelp() {
	fmt.Fprintln(os.Stdout, "Usage:")
	fmt.Fprintln(os.Stdout, "  sleep [OPTIONS] NUMBER[SUFFIX]")
	fmt.Fprintln(os.Stdout, "  SUFFIX is s(seconds, default), m(minutes), h(hours), d(days)")
	fmt.Fprintln(os.Stdout, "")
	fmt.Fprintln(os.Stdout, "Application Options:")
	fmt.Fprintln(os.Stdout, "  -v, --version       Show sleep command version")
	fmt.Fprintln(os.Stdout, "")
	fmt.Fprintln(os.Stdout, "Help Options:")
	fmt.Fprintln(os.Stdout, "  -h, --help          Show this help message")
}
