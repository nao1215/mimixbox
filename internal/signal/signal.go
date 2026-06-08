// Package signal is the single source of truth for the signal name/number
// table shared by the kill, killall and timeout applets. Keeping one table and
// one set of resolvers here is what stops those applets from drifting apart.
package signal

import (
	"fmt"
	"strconv"
	"strings"
)

// Signal is one entry of the signal table.
type Signal struct {
	Number int
	Name   string // SIG-prefixed canonical name, e.g. "SIGTERM"
	Desc   string
}

// table lists the standard signals 1-31. Real-time signals (33-64) have no
// predefined meaning and are intentionally omitted.
//
// [Reference]
// https://www-uxsup.csx.cam.ac.uk/courses/moved.Building/signals.pdf
var table = []Signal{
	{1, "SIGHUP", "Hangup detected on controlling terminal or death of controlling process"},
	{2, "SIGINT", "The process was interrupted (When user hits Cnrl+C)"},
	{3, "SIGQUIT", "Quit program"},
	{4, "SIGILL", "Illegal instruction"},
	{5, "SIGTRAP", "Trace trap for debugging"},
	{6, "SIGABRT", "Emegency stop.  Abort program (formerly SIGIOT)"},
	{7, "SIGBUS", "Bus error. e.g. alignment errors in memory access"},
	{8, "SIGFPE", "A floating point exception happened in the program."},
	{9, "SIGKILL", "Kill program"},
	{10, "SIGUSR1", "Left for the programmers to do whatever they want"},
	{11, "SIGSEGV", "Segmentation violation"},
	{12, "SIGUSR2", "Left for the programmers to do whatever they want"},
	{13, "SIGPIPE", "Write on a pipe with no reader"},
	{14, "SIGALRM", "Real-time timer (request a wake up call) expired"},
	{15, "SIGTERM", "Software termination"},
	{16, "SIGSTKFLT", "Unused (Stack fault in the FPU)"},
	{17, "SIGCHLD", "Stop or exit child process"},
	{18, "SIGCONT", "Restart from stop"},
	{19, "SIGSTOP", "Stop process"},
	{20, "SIGTSTP", "Stop process from terminal (When user hits Cnrl+Z)"},
	{21, "SIGTTIN", "Signal to a backgrounded process when it tries to read input from its terminal"},
	{22, "SIGTTOU", "Signal to a backgrounded process when it tries to write output to its terminal"},
	{23, "SIGURG", "Network connection when urgent out of band data is sent to it"},
	{24, "SIGXCPU", "Exceeded CPU limit"},
	{25, "SIGXFSZ", "Exceeded file size limit"},
	{26, "SIGVTALRM", "Virtual alram cloc"},
	{27, "SIGPROF", "Profiling timer's timeout"},
	{28, "SIGWINCH", "Window resize signal"},
	{29, "SIGIO", "Input / output is possible"},
	{30, "SIGPWR", "Power failure"},
	{31, "SIGSYS", "Unused (Illegal argument to routine)"},
}

// List returns the signal table. The returned slice is a fresh copy, so callers
// cannot mutate the canonical data.
func List() []Signal {
	out := make([]Signal, len(table))
	copy(out, table)
	return out
}

// Name returns the SIG-prefixed canonical name for a signal number.
func Name(num int) (string, bool) {
	for _, s := range table {
		if s.Number == num {
			return s.Name, true
		}
	}
	return "", false
}

// normalizeName upper-cases spec and ensures the SIG prefix, so that "kill",
// "KILL" and "SIGKILL" all map to the canonical "SIGKILL".
func normalizeName(spec string) string {
	name := strings.ToUpper(spec)
	if !strings.HasPrefix(name, "SIG") {
		name = "SIG" + name
	}
	return name
}

// numberByName looks up a signal number by a name spec (case-insensitive, with
// or without the SIG prefix).
func numberByName(spec string) (int, bool) {
	name := normalizeName(spec)
	for _, s := range table {
		if s.Name == name {
			return s.Number, true
		}
	}
	return 0, false
}

// Number resolves a signal specification to its number, validating numbers
// against the table. It accepts a decimal number ("9"), a full name
// ("SIGKILL") or a short name ("KILL", case-insensitive). The number 0 is the
// POSIX null signal (used for existence checks). Unknown specs return an error.
// This strict form is what lets kill tell a real "-SIGNAL" operand from an
// ordinary dash argument.
func Number(spec string) (int, error) {
	if spec == "" {
		return 0, fmt.Errorf("empty signal specification")
	}
	if n, err := strconv.Atoi(spec); err == nil {
		if n == 0 {
			return 0, nil
		}
		if _, ok := Name(n); ok {
			return n, nil
		}
		return 0, fmt.Errorf("unknown signal number %q", spec)
	}
	if n, ok := numberByName(spec); ok {
		return n, nil
	}
	return 0, fmt.Errorf("unknown signal name %q", spec)
}

// NumberLax is like Number but forwards any decimal number as-is instead of
// validating it against the table, matching GNU timeout/killall which pass an
// arbitrary signal number straight to the kernel. Names still must be known.
func NumberLax(spec string) (int, error) {
	if n, err := strconv.Atoi(spec); err == nil {
		return n, nil
	}
	if n, ok := numberByName(spec); ok {
		return n, nil
	}
	return 0, fmt.Errorf("unknown signal %q", spec)
}
