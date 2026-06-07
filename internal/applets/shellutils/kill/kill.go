// Package kill implements the kill applet: terminate processes or send them a
// signal, following GNU/POSIX kill semantics.
package kill

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the kill applet.
type Command struct{}

// New returns a kill command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "kill" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Kill process or send signal to process" }

// defaultSignal is sent when no signal is specified, matching GNU/POSIX kill.
const defaultSignal = 15 // SIGTERM

// signal describes a single entry of the signal table.
type signal struct {
	number string
	name   string
	desc   string
}

// signals is the signal table the applet knows about. The list and the
// descriptions are preserved from the original implementation.
//
// [Reference]
// https://www-uxsup.csx.cam.ac.uk/courses/moved.Building/signals.pdf
var signals = []signal{
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

// Run executes kill.
//
// Forms:
//
//	kill PID...            send SIGTERM to each PID
//	kill -s SIGNAL PID...  send SIGNAL to each PID
//	kill -SIGNAL PID...    send SIGNAL to each PID (e.g. -9, -KILL, -SIGKILL)
//	kill -l                list signal names
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-s SIGNAL | -SIGNAL] PID... | -l", stdio.Err)
	sigName := fs.StringP("signal", "s", "", "send the given SIGNAL instead of SIGTERM")
	list := fs.BoolP("list", "l", false, "list signal names")

	// Separate POSIX "-SIGNAL" style operands (e.g. -9, -KILL, -SIGKILL) that
	// pflag cannot parse as flags. Anything that is a known signal spec after
	// the leading dash is pulled out before parsing; the rest is left to pflag.
	directSig := ""
	rest := make([]string, 0, len(args))
	for _, a := range args {
		if len(a) > 1 && a[0] == '-' && a != "--" && a[1] != '-' {
			if spec := strings.TrimLeft(a, "-"); isSignalSpec(spec) {
				directSig = spec
				continue
			}
		}
		rest = append(rest, a)
	}

	proceed, err := fs.Parse(stdio, rest)
	if err != nil || !proceed {
		return err
	}

	if *list {
		writeSignalList(stdio.Out)
		return nil
	}

	// Resolve the signal to send.
	sigSpec := *sigName
	if directSig != "" {
		sigSpec = directSig
	}
	sig := defaultSignal
	if sigSpec != "" {
		n, rerr := resolveSignal(sigSpec)
		if rerr != nil {
			_, _ = fmt.Fprintf(stdio.Err, "kill: %s: invalid signal specification\n", sigSpec)
			return command.SilentFailure()
		}
		sig = n
	}

	pids := fs.Args()
	if len(pids) == 0 {
		fs.WriteUsage(stdio.Err)
		return command.SilentFailure()
	}

	return sendSignals(stdio, pids, sig)
}

// sendSignals delivers sig to every operand, reporting per-process failures on
// stderr and continuing. The returned error only sets the exit code.
func sendSignals(stdio command.IO, pids []string, sig int) error {
	var failed bool
	for _, v := range pids {
		pid, err := strconv.Atoi(v)
		if err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "kill: %s: arguments must be process or job IDs\n", v)
			failed = true
			continue
		}
		p, err := os.FindProcess(pid)
		if err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "kill: %s: %v\n", v, err)
			failed = true
			continue
		}
		if err := p.Signal(syscall.Signal(sig)); err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "kill: %s: %v\n", v, err)
			failed = true
		}
	}
	if failed {
		return command.SilentFailure()
	}
	return nil
}

// isSignalSpec reports whether spec is a recognized signal number or name
// (with or without the SIG prefix).
func isSignalSpec(spec string) bool {
	_, err := resolveSignal(spec)
	return err == nil
}

// resolveSignal converts a signal specification to its number. It accepts a
// decimal number ("9"), a full name ("SIGKILL"), or a short name ("KILL"),
// case-insensitively for names. Unknown specs return an error.
func resolveSignal(spec string) (int, error) {
	if spec == "" {
		return 0, fmt.Errorf("empty signal specification")
	}
	// Numeric form: 0 is the POSIX null signal (existence check); any other
	// number must match a known signal in the table.
	if n, err := strconv.Atoi(spec); err == nil {
		if n == 0 {
			return 0, nil
		}
		num := strconv.Itoa(n)
		for _, s := range signals {
			if s.number == num {
				return n, nil
			}
		}
		return 0, fmt.Errorf("unknown signal number %q", spec)
	}
	// Name form: normalize to the SIG-prefixed, upper-case name.
	name := strings.ToUpper(spec)
	if !strings.HasPrefix(name, "SIG") {
		name = "SIG" + name
	}
	for _, s := range signals {
		if s.name == name {
			n, _ := strconv.Atoi(s.number)
			return n, nil
		}
	}
	return 0, fmt.Errorf("unknown signal name %q", spec)
}

// writeSignalList writes the table of known signals to w, one per line.
func writeSignalList(w io.Writer) {
	for _, s := range signals {
		_, _ = fmt.Fprintf(w, "%2s  %10s  %s\n", s.number, s.name, s.desc)
	}
}
