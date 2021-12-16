//
// mimixbox/internal/lib/signal.go
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
package mb

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type signal struct {
	number string
	name   string
	desc   string
}

// [Reference]
// https://www-uxsup.csx.cam.ac.uk/courses/moved.Building/signals.pdf
var Signals []signal = []signal{
	{"1", "SIGHUP", "Hangup detected on controlling terminal or death of controlling process"},
	{"2", "SIGINT", "The process was interrupted (When user hits Cnrl+C)"},
	{"3", "SIGQUIT", "Quit program"},
	{"4", "SIGILL", "Illegal instruction"},
	{"5", "SIGTRAP", "Trace trap for debugging"},
	{"6", "SIGABRT", "Emegency stop.  Abort program (formerly SIGIOT)"},
	{"7", "SIGBUS", "Bus error. e.g. alignment errors in memory access"},
	{"8", "SIGFPE", "A floating point exception happened in the program."},
	{"9", "SIGKILL", "Kill program"},
	{"10", "SIGUSR1", "Left for the programmers to do whatever they want"},
	{"11", "SIGSEGV", "Segmentation violation"},
	{"12", "SIGUSR2", "Left for the programmers to do whatever they want"},
	{"13", "SIGPIPE", "Write on a pipe with no reader"},
	{"14", "SIGALRM", "Real-time timer (request a wake up call) expired"},
	{"15", "SIGTERM", "Software termination"},
	{"16", "SIGSTKFLT", "Unused (Stack fault in the FPU)"},
	{"17", "SIGCHLD", "Stop or exit child process"},
	{"18", "SIGCONT", "Restart from stop"},
	{"19", "SIGSTOP", "Stop process"},
	{"20", "SIGTSTP", "Stop process from terminal (When user hits Cnrl+Z)"},
	{"21", "SIGTTIN", "Signal to a backgrounded process when it tries to read input from its terminal"},
	{"22", "SIGTTOU", "Signal to a backgrounded process when it tries to write output to its terminal"},
	{"23", "SIGURG", "Network connection when urgent out of band data is sent to it"},
	{"24", "SIGXCPU", "Exceeded CPU limit"},
	{"25", "SIGXFSZ", "Exceeded file size limit"},
	{"26", "SIGVTALRM", "Virtual alram cloc"},
	{"27", "SIGPROF", "Profiling timer's timeout"},
	{"28", "SIGWINCH", "Window resize signal"},
	{"29", "SIGIO", "Input / output is possible"},
	{"30", "SIGPWR", "Power failure"},
	{"31", "SIGSYS", "Unused (Illegal argument to routine)"},
	// signal number 33-64 is real-time signal.
	// It has no predefined meaning and can be used for application-defined purposes.
}

func IsSignalNumber(num string) bool {
	for _, v := range Signals {
		if num == v.number {
			return true
		}
	}
	return false
}

func IsSignalName(name string) bool {
	for _, v := range Signals {
		if name == v.name {
			return true
		}
		shortName := strings.TrimPrefix(v.name, "SIG")
		if name == shortName {
			return true
		}
	}
	return false
}

func SignalAtoi(num string) int32 {
	n, err := strconv.Atoi(num)
	if err != nil {
		return int32(-1)
	}
	return int32(n)
}

func ConvSignalNameToNum(name string) int32 {
	if !strings.HasPrefix(name, "SIG") {
		name = "SIG" + name
	}
	for _, v := range Signals {
		if name == v.name {
			return SignalAtoi(v.number)
		}
	}
	return int32(-1)
}

func PrintSignalList() {
	for _, v := range Signals {
		fmt.Fprintf(os.Stdout, "%2s  %10s  %s\n", v.number, v.name, v.desc)
	}
}

func PrintSignal(numOrName string) {
	for _, v := range Signals {
		if numOrName == v.name {
			fmt.Fprintln(os.Stdout, v.number)
		}
		shortName := strings.TrimPrefix(v.name, "SIG")
		if numOrName == shortName {
			fmt.Fprintln(os.Stdout, v.number)
		}
		if numOrName == v.number {
			fmt.Fprintln(os.Stdout, strings.TrimPrefix(v.name, "SIG"))
		}
	}
}
