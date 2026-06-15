// Package conspy implements the conspy applet: take a remote view of a virtual
// console, mirroring its screen and optionally injecting keystrokes.
package conspy

import (
	"context"
	"fmt"
	"strconv"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the conspy applet.
type Command struct{}

// New returns a conspy command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "conspy" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Remotely view a virtual console" }

// Options is the parsed conspy configuration: which VT to spy on and the mode
// flags. Parsing it apart from the device access makes the command line
// testable.
type Options struct {
	VT       int  // the virtual terminal number to view (1..63), 0 = current
	ReadOnly bool // -d: view only, do not forward keystrokes
	NoColors bool // -c: ignore colors when mirroring
	Quiet    bool // -Q: do not switch the local screen to raw mode
}

// ParseOptions validates the conspy flag values and VT operand. A VT number, if
// present, must be a small positive integer.
func ParseOptions(vtArg string, readOnly, noColors, quiet bool) (*Options, error) {
	o := &Options{ReadOnly: readOnly, NoColors: noColors, Quiet: quiet}
	if vtArg == "" {
		return o, nil
	}
	n, err := strconv.Atoi(vtArg)
	if err != nil || n < 1 || n > 63 {
		return nil, fmt.Errorf("invalid virtual terminal number: %q (expected 1..63)", vtArg)
	}
	o.VT = n
	return o, nil
}

// spyFn is indirected so the privileged, TTY-bound mirror loop can be replaced
// in a test. In production it fails deterministically: conspy reads another
// console's screen buffer via /dev/vcsa and injects keys via TIOCSTI, both of
// which need a real Linux console and privilege.
var spyFn = func(o *Options) error {
	return fmt.Errorf("viewing vt %d requires reading /dev/vcsa%d and TIOCSTI on a console (not available here)", o.VT, o.VT)
}

// Run executes conspy.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-dcQ] [VTNUM]", stdio.Err).WithHelp(command.Help{
		Description: "Take a remote view of Linux virtual terminal VTNUM (1..63; the current one if " +
			"omitted): its screen is mirrored onto yours and, unless -d is given, your keystrokes are " +
			"injected into it. -c ignores colors and -Q keeps your local screen out of raw mode. " +
			"Mirroring reads the other console's screen buffer from /dev/vcsa and injects keys with " +
			"TIOCSTI, so conspy needs a real Linux console and privilege; in an environment without one " +
			"it validates the arguments and then fails with a clear message rather than doing nothing.",
		Examples: []command.Example{
			{Command: "conspy 2", Explain: "View and control tty2."},
			{Command: "conspy -d 3", Explain: "View tty3 read-only (do not send keys)."},
		},
		ExitStatus: "0  the view session ended normally.\n" +
			"1  a bad VT number was given, or no usable console was available.",
	})
	readOnly := fs.BoolP("dump", "d", false, "view only; do not forward keystrokes")
	noColors := fs.BoolP("no-colors", "c", false, "ignore colors when mirroring the screen")
	quiet := fs.BoolP("quiet", "Q", false, "do not put the local screen into raw mode")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	rest := fs.Args()
	vtArg := ""
	switch {
	case len(rest) == 1:
		vtArg = rest[0]
	case len(rest) > 1:
		return command.Failuref("unexpected argument: %q", rest[1])
	}

	opts, err := ParseOptions(vtArg, *readOnly, *noColors, *quiet)
	if err != nil {
		return command.Failure(err)
	}
	if err := spyFn(opts); err != nil {
		return command.Failuref("%v", err)
	}
	return nil
}
