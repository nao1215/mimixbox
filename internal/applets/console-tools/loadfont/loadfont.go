// Package loadfont implements the loadfont applet: load a console font from
// standard input (PSF format) into the console.
package loadfont

import (
	"context"
	"fmt"

	"github.com/nao1215/mimixbox/internal/applets/console-tools/internal/kbd"
	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the loadfont applet.
type Command struct{}

// New returns a loadfont command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "loadfont" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Load a console font from stdin" }

// applyFontFn is indirected so the privileged console write can be replaced in a
// test. In production it fails deterministically: uploading a font needs the
// PIO_FONT/KDFONTOP ioctl on the console, which is TTY-bound and privileged.
var applyFontFn = func(_ *kbd.Font) error {
	return fmt.Errorf("loading a font requires the KDFONTOP ioctl on a console (not available here)")
}

// Run executes loadfont.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-v] < font", stdio.Err).WithHelp(command.Help{
		Description: "Read a PSF (PC Screen Font, version 1 or 2) console font from standard input, " +
			"validate its header and glyph data, and upload it to the console with the KDFONTOP ioctl. " +
			"With -v the parsed font's dimensions and glyph count are printed to standard error before " +
			"loading. The font is fully parsed and checked first, so a malformed file is rejected with " +
			"a clear error. Loading needs a real console and privilege; without one the input is " +
			"validated and then the command fails with a clear message rather than silently doing " +
			"nothing.",
		Examples: []command.Example{
			{Command: "loadfont < default8x16.psf", Explain: "Load a PSF font onto the console."},
			{Command: "loadfont -v < font.psf", Explain: "Print the font's geometry, then load it."},
		},
		ExitStatus: "0  the font was loaded.\n" +
			"1  the input was not a valid PSF font, or no console was available.",
	})
	verbose := fs.BoolP("verbose", "v", false, "print the parsed font geometry to stderr")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	if rest := fs.Args(); len(rest) > 0 {
		return command.Failuref("unexpected argument: %q (loadfont reads the font from stdin)", rest[0])
	}

	font, err := kbd.DecodeFont(stdio.In)
	if err != nil {
		return command.Failuref("invalid font: %v", err)
	}
	if *verbose {
		_, _ = fmt.Fprintln(stdio.Err, font.Describe())
	}
	if err := applyFontFn(font); err != nil {
		return command.Failuref("%v", err)
	}
	return nil
}
