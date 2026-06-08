// Package watch implements the watch applet: run a command periodically and
// display its output, refreshing the screen on each run.
package watch

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the watch applet.
type Command struct{}

// New returns a watch command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "watch" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Execute a program periodically, showing output fullscreen" }

// clearScreen is the escape sequence that homes the cursor and clears the
// screen between refreshes.
const clearScreen = "\033[H\033[2J"

// Run executes watch. It renders once immediately and then again every
// interval until the context is cancelled (for example by Ctrl-C).
func (c *Command) Run(ctx context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... COMMAND [ARG]...", stdio.Err)
	// Stop parsing options at the command name so its flags pass through.
	fs.SetInterspersed(false)
	interval := fs.Float64P("interval", "n", 2.0, "seconds to wait between updates")
	noTitle := fs.BoolP("no-title", "t", false, "turn off the header showing the interval, command and time")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	rest := fs.Args()
	if len(rest) == 0 {
		return command.Failuref("missing command operand")
	}
	if *interval <= 0 {
		return command.Failuref("interval must be positive")
	}

	render := func() error { return c.renderOnce(stdio, *interval, *noTitle, rest) }
	if err := render(); err != nil {
		return err
	}

	ticker := time.NewTicker(time.Duration(*interval * float64(time.Second)))
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := render(); err != nil {
				return err
			}
		}
	}
}

// renderOnce clears the screen, writes the optional header and then the output
// of one run of the command.
func (c *Command) renderOnce(stdio command.IO, interval float64, noTitle bool, argv []string) error {
	var b bytes.Buffer
	b.WriteString(clearScreen)
	if !noTitle {
		b.WriteString(header(interval, argv))
	}
	b.WriteString(runCommand(argv))
	if _, err := io.Copy(stdio.Out, &b); err != nil {
		return command.Failure(err)
	}
	return nil
}

// header renders the title line: the interval and the command being watched.
func header(interval float64, argv []string) string {
	return fmt.Sprintf("Every %ss: %s\n\n", strconv.FormatFloat(interval, 'g', -1, 64), join(argv))
}

// runCommand executes argv once and returns its combined output, or the error
// text when it cannot run, so something is always shown on screen.
func runCommand(argv []string) string {
	cmd := exec.Command(argv[0], argv[1:]...) //nolint:gosec // running a user-named command is the point
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out) + fmt.Sprintf("watch: %v\n", err)
	}
	return string(out)
}

// join concatenates argv with single spaces.
func join(argv []string) string {
	s := ""
	for i, a := range argv {
		if i > 0 {
			s += " "
		}
		s += a
	}
	return s
}
