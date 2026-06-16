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
	"github.com/nao1215/mimixbox/internal/signal"
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

// signals is the canonical signal table, shared with killall and timeout.
var signals = signal.List()

// Run executes kill.
//
// Forms:
//
//	kill PID...            send SIGTERM to each PID
//	kill -s SIGNAL PID...  send SIGNAL to each PID
//	kill -SIGNAL PID...    send SIGNAL to each PID (e.g. -9, -KILL, -SIGKILL)
//	kill -l                list signal names
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-s SIGNAL | -SIGNAL] PID... | -l", stdio.Err).WithHelp(command.Help{
		Description: "Send a signal to each process identified by PID. The default signal is " +
			"SIGTERM (15); a different signal may be given by name or number with -s or as " +
			"-SIGNAL. With -l, list the known signal names.",
		Examples: []command.Example{
			{Command: "kill 1234", Explain: "Send SIGTERM to process 1234."},
			{Command: "kill -9 1234", Explain: "Send SIGKILL to process 1234."},
			{Command: "kill -s HUP 1234", Explain: "Send SIGHUP to process 1234."},
			{Command: "kill -l", Explain: "List the available signal names."},
		},
		ExitStatus: "0  every signal was delivered.\n1  a PID was invalid or a signal could not be sent.",
	})
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
// case-insensitively for names. Unknown specs return an error. It delegates to
// the canonical strict resolver so kill, killall and timeout stay in sync.
func resolveSignal(spec string) (int, error) {
	return signal.Number(spec)
}

// writeSignalList writes the table of known signals to w, one per line.
func writeSignalList(w io.Writer) {
	for _, s := range signals {
		_, _ = fmt.Fprintf(w, "%2d  %10s  %s\n", s.Number, s.Name, s.Desc)
	}
}
