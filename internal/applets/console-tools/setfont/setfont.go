// Package setfont implements the setfont applet: load a console font named on
// the command line (PSF format) into the console.
package setfont

import (
	"context"
	"fmt"
	"os"

	"github.com/nao1215/mimixbox/internal/applets/console-tools/internal/kbd"
	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the setfont applet.
type Command struct{}

// New returns a setfont command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "setfont" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Load a console font from a file" }

// applyFontFn is indirected so the privileged console write can be replaced in a
// test. In production it fails deterministically: uploading a font needs the
// KDFONTOP ioctl on the console, which is TTY-bound and privileged.
var applyFontFn = func(_ *kbd.Font) error {
	return fmt.Errorf("loading a font requires the KDFONTOP ioctl on a console (not available here)")
}

// Run executes setfont.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-v] FONT", stdio.Err).WithHelp(command.Help{
		Description: "Read the PSF (PC Screen Font, version 1 or 2) console font named by FONT, validate " +
			"its header and glyph data, and upload it to the console with the KDFONTOP ioctl. With -v " +
			"the parsed font's dimensions and glyph count are printed before loading. The font is fully " +
			"parsed and checked first, so a missing or malformed file is rejected with a clear error. " +
			"Loading needs a real console and privilege; without one the file is validated and then the " +
			"command fails with a clear message rather than silently doing nothing. Unlike the kbd " +
			"setfont, the unicode map and ACM options are not yet implemented.",
		Examples: []command.Example{
			{Command: "setfont /usr/share/consolefonts/default8x16.psf", Explain: "Load a font file onto the console."},
			{Command: "setfont -v font.psf", Explain: "Print the font's geometry, then load it."},
		},
		ExitStatus: "0  the font was loaded.\n" +
			"1  no FONT was given, the file was missing or not a valid PSF font, or no console was available.",
	})
	verbose := fs.BoolP("verbose", "v", false, "print the parsed font geometry before loading")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	rest := fs.Args()
	if len(rest) == 0 {
		return command.Failuref("a font file is required")
	}
	if len(rest) > 1 {
		return command.Failuref("unexpected argument: %q", rest[1])
	}

	f, err := os.Open(rest[0]) //nolint:gosec // user-named font file
	if err != nil {
		return command.Failuref("cannot open font %q: %v", rest[0], err)
	}
	defer func() { _ = f.Close() }()

	font, err := kbd.DecodeFont(f)
	if err != nil {
		return command.Failuref("invalid font %q: %v", rest[0], err)
	}
	if *verbose {
		_, _ = fmt.Fprintln(stdio.Out, font.Describe())
	}
	if err := applyFontFn(font); err != nil {
		return command.Failuref("%v", err)
	}
	return nil
}
