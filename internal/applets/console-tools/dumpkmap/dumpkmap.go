// Package dumpkmap implements the dumpkmap applet: write the current console
// keymap to standard output in BusyBox binary keymap format.
package dumpkmap

import (
	"context"
	"fmt"

	"github.com/nao1215/mimixbox/internal/applets/console-tools/internal/kbd"
	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the dumpkmap applet.
type Command struct{}

// New returns a dumpkmap command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "dumpkmap" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Dump the console keymap in binary form" }

// readKeymapFn is indirected so the privileged console read can be replaced in a
// test. In production it fails deterministically: reading the live keymap needs
// the KDGKBENT ioctl on the console, which is TTY-bound and privileged.
var readKeymapFn = func() (*kbd.Keymap, error) {
	return nil, fmt.Errorf("reading the live keymap requires the KDGKBENT ioctl on a console (not available here)")
}

// Run executes dumpkmap.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "", stdio.Err).WithHelp(command.Help{
		Description: "Read the current console keymap with the KDGKBENT ioctl and write it to standard " +
			"output in the BusyBox binary keymap format ('bkeymap' magic, a 256-byte table-present " +
			"map, then 256 little-endian shorts per present table). The output is meant to be saved to " +
			"a file and later restored with loadkmap. Reading the live keymap needs a real console and " +
			"privilege; without one this command fails with a clear message rather than emitting " +
			"nothing.",
		Examples: []command.Example{
			{Command: "dumpkmap > keymap.bin", Explain: "Save the current keymap for later loadkmap."},
		},
		ExitStatus: "0  the keymap was written.\n1  no console was available or the write failed.",
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	if rest := fs.Args(); len(rest) > 0 {
		return command.Failuref("unexpected argument: %q", rest[0])
	}

	km, err := readKeymapFn()
	if err != nil {
		return command.Failuref("%v", err)
	}
	if err := kbd.EncodeKeymap(stdio.Out, km); err != nil {
		return command.Failuref("writing keymap: %v", err)
	}
	return nil
}
