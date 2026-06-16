// Package script implements the script and scriptreplay applets: record a
// command's output (with a timing file) to a typescript, and replay it with the
// original pacing.
//
// The transcript and timing serialization shared by both entrypoints (the
// recorder, the "delay bytes" timing format, and the replay parsing) lives in
// transcript.go; this file keeps the script (recorder) and scriptreplay
// (replayer) command entrypoints.
package script

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the script or scriptreplay applet.
type Command struct{ replay bool }

// NewScript returns a script command.
func NewScript() *Command { return &Command{} }

// NewScriptreplay returns a scriptreplay command.
func NewScriptreplay() *Command { return &Command{replay: true} }

// Name returns the command name.
func (c *Command) Name() string {
	if c.replay {
		return "scriptreplay"
	}
	return "script"
}

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string {
	if c.replay {
		return "Replay a typescript using its timing file"
	}
	return "Record a command's output to a typescript"
}

// Run dispatches to the recorder or the player.
func (c *Command) Run(ctx context.Context, stdio command.IO, args []string) error {
	if c.replay {
		return c.runReplay(ctx, stdio, args)
	}
	return c.runScript(ctx, stdio, args)
}

// runScript records a command session.
func (c *Command) runScript(ctx context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-c COMMAND] [-T TIMINGFILE] [TYPESCRIPT]", stdio.Err).WithHelp(command.Help{
		Description: "Run COMMAND (with -c) and write everything it prints to TYPESCRIPT (default " +
			"'typescript'), wrapped in 'Script started/done' lines. With -T, also write a timing " +
			"file of 'delay bytes' records that scriptreplay can use.",
		Examples: []command.Example{
			{Command: "script -c 'make' -T timing build.log", Explain: "Record make's output and timing."},
		},
		ExitStatus: "The exit status of COMMAND.",
	})
	cmdStr := fs.StringP("command", "c", "", "run COMMAND rather than an interactive shell")
	timingFile := fs.StringP("timing", "T", "", "write timing data to this file")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	if *cmdStr == "" {
		_, _ = fmt.Fprintln(stdio.Err, "script: -c COMMAND is required in this implementation")
		return command.SilentFailure()
	}

	typescript := "typescript"
	if rest := fs.Args(); len(rest) > 0 {
		typescript = rest[0]
	}

	rec := &recorder{mirror: stdio.Out}
	cmd := exec.CommandContext(ctx, "sh", "-c", *cmdStr) //nolint:gosec // running the user's command is the point
	cmd.Stdin = stdio.In
	cmd.Stdout = rec
	cmd.Stderr = rec
	runErr := cmd.Run()
	exitCode := 0
	if ee, ok := runErr.(*exec.ExitError); ok {
		exitCode = ee.ExitCode()
	}

	started := clock().Format("2006-01-02 15:04:05-07:00")
	var b strings.Builder
	fmt.Fprintf(&b, "Script started on %s [COMMAND=%q]\n", started, *cmdStr)
	b.Write(rec.buf.Bytes())
	fmt.Fprintf(&b, "\nScript done on %s [COMMAND_EXIT_CODE=%q]\n", started, strconv.Itoa(exitCode))
	if err := os.WriteFile(typescript, []byte(b.String()), 0o644); err != nil { //nolint:gosec // user-named file
		_, _ = fmt.Fprintf(stdio.Err, "script: %v\n", err)
		return command.SilentFailure()
	}

	if *timingFile != "" {
		if err := os.WriteFile(*timingFile, []byte(formatTiming(rec.timing)), 0o644); err != nil { //nolint:gosec // user-named file
			_, _ = fmt.Fprintf(stdio.Err, "script: %v\n", err)
			return command.SilentFailure()
		}
	}

	if exitCode != 0 {
		return &command.ExitError{Code: exitCode}
	}
	return nil
}

// runReplay plays back a recorded session.
func (c *Command) runReplay(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "TIMINGFILE [TYPESCRIPT]", stdio.Err).WithHelp(command.Help{
		Description: "Replay the captured output in TYPESCRIPT (default 'typescript') to standard " +
			"output, pausing between chunks by the delays recorded in TIMINGFILE.",
		Examples: []command.Example{
			{Command: "scriptreplay timing build.log", Explain: "Replay the recorded session."},
		},
		ExitStatus: "0  the session replayed successfully.\n1  the timing or typescript file could not be read.",
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	rest := fs.Args()
	if len(rest) == 0 {
		_, _ = fmt.Fprintln(stdio.Err, "scriptreplay: a timing file is required")
		return command.SilentFailure()
	}
	timingPath := rest[0]
	typescript := "typescript"
	if len(rest) > 1 {
		typescript = rest[1]
	}

	timing, err := readTiming(timingPath)
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "scriptreplay: %s\n", command.FileError(timingPath, err))
		return command.SilentFailure()
	}
	data, err := os.ReadFile(typescript) //nolint:gosec // user-named file
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "scriptreplay: %s\n", command.FileError(typescript, err))
		return command.SilentFailure()
	}

	replay(stdio.Out, transcriptBody(data), timing)
	return nil
}
