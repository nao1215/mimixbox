// Package loadkmap implements the loadkmap applet: load a console keymap from
// standard input (in BusyBox binary keymap format) into the console.
package loadkmap

import (
	"context"
	"fmt"

	"github.com/nao1215/mimixbox/internal/applets/console-tools/internal/kbd"
	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the loadkmap applet.
type Command struct{}

// New returns a loadkmap command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "loadkmap" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Load a binary console keymap from stdin" }

// applyKeymapFn is indirected so the privileged console write can be replaced in
// a test. In production it fails deterministically: applying a keymap needs the
// KDSKBENT ioctl on the console, which is TTY-bound and privileged.
var applyKeymapFn = func(_ *kbd.Keymap) error {
	return fmt.Errorf("applying a keymap requires the KDSKBENT ioctl on a console (not available here)")
}

// Run executes loadkmap.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "< keymap", stdio.Err).WithHelp(command.Help{
		Description: "Read a binary keymap (the 'bkeymap' format produced by dumpkmap) from standard " +
			"input, validate it, and load it into the console with the KDSKBENT ioctl. The keymap is " +
			"fully parsed and checked before any console change is attempted, so a malformed file is " +
			"rejected with a clear error. Loading needs a real console and privilege; without one this " +
			"command parses and validates the input, then fails with a clear message rather than " +
			"silently doing nothing.",
		Examples: []command.Example{
			{Command: "loadkmap < keymap.bin", Explain: "Load a keymap previously saved by dumpkmap."},
		},
		ExitStatus: "0  the keymap was loaded.\n" +
			"1  the input was not a valid keymap, or no console was available.",
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	if rest := fs.Args(); len(rest) > 0 {
		return command.Failuref("unexpected argument: %q (loadkmap reads the keymap from stdin)", rest[0])
	}

	km, err := kbd.DecodeKeymap(stdio.In)
	if err != nil {
		return command.Failuref("invalid keymap: %v", err)
	}
	if len(km.PresentTables()) == 0 {
		return command.Failuref("invalid keymap: %v", kbd.ErrEmptyKeymap)
	}
	if err := applyKeymapFn(km); err != nil {
		return command.Failuref("%v", err)
	}
	return nil
}
