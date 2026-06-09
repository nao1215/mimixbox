// Package script implements the script and scriptreplay applets: record a
// command's output (with a timing file) to a typescript, and replay it with the
// original pacing.
package script

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
)

// clock is indirected so recorded timing is deterministic under test.
var clock = time.Now

// sleep is indirected so replay does not actually wait under test.
var sleep = time.Sleep

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

type entry struct {
	delay float64
	bytes int
}

// recorder tees writes into a buffer while recording per-write timing.
type recorder struct {
	buf    bytes.Buffer
	timing []entry
	last   time.Time
	mirror io.Writer
}

func (r *recorder) Write(p []byte) (int, error) {
	t := clock()
	delay := 0.0
	if !r.last.IsZero() {
		delay = t.Sub(r.last).Seconds()
	}
	r.last = t
	r.timing = append(r.timing, entry{delay: delay, bytes: len(p)})
	if r.mirror != nil {
		_, _ = r.mirror.Write(p)
	}
	return r.buf.Write(p)
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
		var tb strings.Builder
		for _, e := range rec.timing {
			fmt.Fprintf(&tb, "%.6f %d\n", e.delay, e.bytes)
		}
		if err := os.WriteFile(*timingFile, []byte(tb.String()), 0o644); err != nil { //nolint:gosec // user-named file
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

	// Skip the "Script started" header line; the timing covers the bytes after it.
	body := data
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		body = data[i+1:]
	}

	pos := 0
	for _, e := range timing {
		sleep(time.Duration(e.delay * float64(time.Second)))
		end := pos + e.bytes
		if end > len(body) {
			end = len(body)
		}
		if pos < end {
			_, _ = stdio.Out.Write(body[pos:end])
		}
		pos = end
	}
	return nil
}

// readTiming parses a "delay bytes" timing file.
func readTiming(path string) ([]entry, error) {
	f, err := os.Open(path) //nolint:gosec // user-named file
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	var out []entry
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) != 2 {
			continue
		}
		delay, _ := strconv.ParseFloat(fields[0], 64)
		n, _ := strconv.Atoi(fields[1])
		out = append(out, entry{delay: delay, bytes: n})
	}
	return out, sc.Err()
}
