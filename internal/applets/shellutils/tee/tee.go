// Package tee implements the tee applet: read from standard input and write to
// standard output and to each named file, fanning the stream out with the
// common GNU options.
package tee

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"syscall"

	"github.com/nao1215/mimixbox/internal/command"
)

// isBrokenPipe reports whether err is (or wraps) an EPIPE write error, the
// condition the *-nopipe output-error modes silently tolerate.
func isBrokenPipe(err error) bool {
	return errors.Is(err, syscall.EPIPE)
}

// outputErrorMode selects how tee reacts to write errors, mirroring GNU tee's
// --output-error[=MODE].
type outputErrorMode int

const (
	// outputErrorWarnNopipe diagnoses write errors but ignores broken-pipe
	// errors; tee keeps going and exits nonzero at the end. This is the GNU
	// default when --output-error is not given.
	outputErrorWarnNopipe outputErrorMode = iota
	// outputErrorWarn diagnoses every write error (including broken pipe) and
	// keeps going, exiting nonzero at the end.
	outputErrorWarn
	// outputErrorExit exits on the first write error that is not a broken pipe.
	outputErrorExit
	// outputErrorExitNopipe exits on the first write error, treating a broken
	// pipe as a normal (silent, successful) end of output.
	outputErrorExitNopipe
)

// parseOutputErrorMode maps a --output-error MODE string onto an
// outputErrorMode. The empty string selects GNU's "--output-error" default of
// "warn".
func parseOutputErrorMode(mode string) (outputErrorMode, error) {
	switch mode {
	case "", "warn":
		return outputErrorWarn, nil
	case "warn-nopipe":
		return outputErrorWarnNopipe, nil
	case "exit":
		return outputErrorExit, nil
	case "exit-nopipe":
		return outputErrorExitNopipe, nil
	default:
		return 0, fmt.Errorf("invalid argument %q for --output-error", mode)
	}
}

// ignoresPipe reports whether the mode treats a broken-pipe write error as a
// normal, silent end of output rather than a diagnosable error.
func (m outputErrorMode) ignoresPipe() bool {
	return m == outputErrorWarnNopipe || m == outputErrorExitNopipe
}

// exitsOnError reports whether the mode stops at the first reportable write
// error instead of continuing to the remaining writers.
func (m outputErrorMode) exitsOnError() bool {
	return m == outputErrorExit || m == outputErrorExitNopipe
}

// Command is the tee applet.
type Command struct{}

// New returns a tee command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "tee" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string {
	return "Read from standard input and write to standard output and files"
}

// Run executes tee.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... [FILE]...", stdio.Err).WithHelp(command.Help{
		Description: "Copy standard input to standard output and to each FILE. By default each FILE is " +
			"overwritten; with -a the input is appended instead.",
		Examples: []command.Example{
			{Command: "tee out.txt", Explain: "Write standard input to out.txt and to standard output."},
			{Command: "tee -a log.txt", Explain: "Append standard input to log.txt instead of overwriting it."},
		},
		ExitStatus: "0  success.\n1  an error occurred (e.g. a file could not be written).",
	})
	appendMode := fs.BoolP("append", "a", false, "append to the given FILEs, do not overwrite")
	// -i/--ignore-interrupts is accepted for GNU compatibility; in this
	// in-memory model there is no SIGINT to ignore, so it is a no-op.
	_ = fs.BoolP("ignore-interrupts", "i", false, "ignore interrupt signals")
	// --output-error[=MODE] takes an optional value. When given without "=MODE"
	// GNU selects "warn"; when omitted entirely the default is "warn-nopipe".
	fs.String("output-error", "",
		"set behavior on write error: warn, warn-nopipe, exit, exit-nopipe (default warn when given without a MODE)")
	fs.Lookup("output-error").NoOptDefVal = "warn"

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	modeStr, _ := fs.GetString("output-error")
	mode, err := parseOutputErrorMode(modeStr)
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "tee: %v\n", err)
		_, _ = fmt.Fprintln(stdio.Err, "Try 'tee --help' for more information.")
		return command.SilentFailure()
	}

	return tee(stdio, fs.Args(), *appendMode, mode)
}

// teeWriter pairs a destination with the label used to diagnose its errors and
// tracks whether it has already failed so a broken writer is skipped for the
// rest of the stream.
type teeWriter struct {
	w      io.Writer
	label  string // empty for standard output
	closer io.Closer
	dead   bool
}

// tee copies standard input to standard output and to each named file. How a
// write error is handled depends on mode: by default (warn-nopipe) errors are
// reported on stderr, broken pipes are ignored, and tee keeps writing to the
// remaining destinations, exiting nonzero at the end; the exit modes stop at the
// first reportable error instead.
func tee(stdio command.IO, files []string, appendMode bool, mode outputErrorMode) error {
	flags := os.O_CREATE | os.O_WRONLY
	if appendMode {
		flags |= os.O_APPEND
	} else {
		flags |= os.O_TRUNC
	}

	dsts := []*teeWriter{{w: stdio.Out}}
	var opened []*os.File
	var failed bool
	for _, name := range files {
		f, err := os.OpenFile(name, flags, 0o644) //nolint:gosec // operating on a user-named file is the whole point
		if err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "tee: %s\n", command.FileError(name, err))
			failed = true
			if mode.exitsOnError() {
				return command.SilentFailure()
			}
			continue
		}
		dsts = append(dsts, &teeWriter{w: f, label: name, closer: f})
		opened = append(opened, f)
	}

	copyFailed, stopped := copyStream(stdio, dsts, mode)
	failed = failed || copyFailed

	for _, f := range opened {
		if cerr := f.Close(); cerr != nil {
			if reportWriteError(stdio, f.Name(), cerr, mode) {
				failed = true
			}
		}
	}

	if failed || stopped {
		return command.SilentFailure()
	}
	return nil
}

// copyStream reads standard input in chunks and writes each chunk to every live
// destination, applying the output-error policy. It returns whether any
// reportable error occurred and whether an exit-mode caused an early stop.
func copyStream(stdio command.IO, dsts []*teeWriter, mode outputErrorMode) (failed, stopped bool) {
	buf := make([]byte, 32*1024)
	for {
		n, rerr := stdio.In.Read(buf)
		if n > 0 {
			chunk := buf[:n]
			for _, d := range dsts {
				if d.dead {
					continue
				}
				if _, werr := d.w.Write(chunk); werr != nil {
					if reportWriteError(stdio, d.label, werr, mode) {
						failed = true
					}
					d.dead = true
					if mode.exitsOnError() {
						return failed, true
					}
				}
			}
		}
		if rerr != nil {
			if !errors.Is(rerr, io.EOF) {
				_, _ = fmt.Fprintf(stdio.Err, "tee: %v\n", rerr)
				failed = true
			}
			return failed, stopped
		}
	}
}

// reportWriteError diagnoses a write error according to mode and reports whether
// it counts toward a nonzero exit. A broken-pipe error is silently swallowed in
// the *-nopipe modes; for the standard-output destination (empty label) the bare
// error is printed, otherwise the GNU file-error form is used.
func reportWriteError(stdio command.IO, label string, err error, mode outputErrorMode) bool {
	if mode.ignoresPipe() && isBrokenPipe(err) {
		return false
	}
	if label == "" {
		_, _ = fmt.Fprintf(stdio.Err, "tee: %v\n", err)
	} else {
		_, _ = fmt.Fprintf(stdio.Err, "tee: %s\n", command.FileError(label, err))
	}
	return true
}
