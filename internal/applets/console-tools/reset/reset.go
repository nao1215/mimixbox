// Package reset implements the reset applet: reset the terminal to its
// initial state by writing the standard reset escape sequence.
package reset

import (
	"context"
	"io"

	"github.com/nao1215/mimixbox/internal/command"
)

// resetSequence is the escape sequence emitted to reset the terminal.
// See 'man 4 console_codes' for details:
//
//	"ESC c"        -- Reset (RIS)
//	"ESC ( B"      -- Select G0 Character Set (B = US)
//	"ESC [ m"      -- Reset all display attributes
//	"ESC [ J"      -- Erase to the end of screen
//	"ESC [ ? 25 h" -- Make cursor visible
const resetSequence = "\033c\033(B\033[m\033[J\033[?25h"

// Command is the reset applet.
type Command struct{}

// New returns a reset command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "reset" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Reset terminal" }

// Run executes reset: it writes the terminal-reset escape sequence to
// stdio.Out.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]", stdio.Err).WithHelp(command.Help{
		Description: "Reset the terminal to its initial state by writing the standard reset escape " +
			"sequence: it restores the default character set and display attributes, clears the " +
			"screen, and makes the cursor visible again. Use it to recover a terminal left in a " +
			"confused state, for example after a program emits raw binary data.",
		Examples: []command.Example{
			{Command: "reset", Explain: "Reset the terminal to a sane, usable state."},
		},
		ExitStatus: "0  the reset sequence was written successfully.",
	})

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	_, _ = io.WriteString(stdio.Out, resetSequence)
	return nil
}
