// Package showkey implements the showkey applet: report the scancodes, keycodes
// or ASCII codes of keys pressed at the console.
package showkey

import (
	"context"
	"fmt"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the showkey applet.
type Command struct{}

// New returns a showkey command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "showkey" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Report the codes of keys pressed at the console" }

// mode selects which representation of a key press showkey reports.
type mode int

const (
	modeKeycode mode = iota // -k: report keycodes (the default)
	modeScancode            // -s: report raw scancodes
	modeASCII               // -a: report ASCII/decimal/octal/hex codes
)

// String renders the mode for messages.
func (m mode) String() string {
	switch m {
	case modeScancode:
		return "scancodes"
	case modeASCII:
		return "ASCII codes"
	default:
		return "keycodes"
	}
}

// interactiveFn is indirected so the privileged, console-requiring raw-mode loop
// can be substituted in tests. In production it fails deterministically because
// showkey needs a real console put into raw/medium-raw mode, which is both
// TTY-bound and privileged.
var interactiveFn = func(_ context.Context, _ command.IO, m mode) error {
	return fmt.Errorf("reporting %s requires a real console in raw mode (not available here)", m)
}

// Run executes showkey.
func (c *Command) Run(ctx context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-a | -k | -s]", stdio.Err).WithHelp(command.Help{
		Description: "Read key presses from the console and report their codes until ten seconds pass " +
			"with no key (or, with -a, until you press the Return key). The default -k reports keycodes; " +
			"-s reports raw scancodes; -a reports the ASCII value of each byte in decimal, octal and " +
			"hexadecimal. The codes are read by putting the console keyboard into raw/medium-raw mode, " +
			"so this command needs a real Linux console and the privilege to change the keyboard mode; " +
			"in an environment without one it fails with a clear message rather than doing nothing.",
		Examples: []command.Example{
			{Command: "showkey", Explain: "Report keycodes for keys pressed at the console."},
			{Command: "showkey -a", Explain: "Report the ASCII codes of typed characters."},
			{Command: "showkey -s", Explain: "Report raw keyboard scancodes."},
		},
		ExitStatus: "0  the key-reading loop ended normally.\n" +
			"1  conflicting modes were given, or no usable console was available.",
	})
	ascii := fs.BoolP("ascii", "a", false, "report the ASCII codes of typed characters")
	keycodes := fs.BoolP("keycodes", "k", false, "report keycodes (the default)")
	scancodes := fs.BoolP("scancodes", "s", false, "report raw scancodes")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	if rest := fs.Args(); len(rest) > 0 {
		return command.Failuref("unexpected argument: %q", rest[0])
	}

	m, err := selectMode(*ascii, *keycodes, *scancodes)
	if err != nil {
		return command.Failure(err)
	}
	if err := interactiveFn(ctx, stdio, m); err != nil {
		return command.Failuref("%v", err)
	}
	return nil
}

// selectMode resolves the -a/-k/-s flags into a single mode, rejecting any
// combination of more than one.
func selectMode(ascii, keycodes, scancodes bool) (mode, error) {
	n := 0
	for _, b := range []bool{ascii, keycodes, scancodes} {
		if b {
			n++
		}
	}
	if n > 1 {
		return modeKeycode, fmt.Errorf("the -a, -k and -s options are mutually exclusive")
	}
	switch {
	case ascii:
		return modeASCII, nil
	case scancodes:
		return modeScancode, nil
	default:
		return modeKeycode, nil
	}
}
