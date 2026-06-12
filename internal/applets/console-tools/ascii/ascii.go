// Package ascii implements the ascii applet: print the table of ASCII codes and
// their characters or control-code names.
package ascii

import (
	"context"
	"fmt"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the ascii applet.
type Command struct{}

// New returns an ascii command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "ascii" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Print the ASCII code table" }

// controlNames holds the mnemonic for each control code 0-31, plus DEL at 127.
var controlNames = []string{
	"NUL", "SOH", "STX", "ETX", "EOT", "ENQ", "ACK", "BEL",
	"BS", "HT", "LF", "VT", "FF", "CR", "SO", "SI",
	"DLE", "DC1", "DC2", "DC3", "DC4", "NAK", "SYN", "ETB",
	"CAN", "EM", "SUB", "ESC", "FS", "GS", "RS", "US",
}

// Run executes ascii.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "", stdio.Err).WithHelp(command.Help{
		Description: "Print the 128-entry ASCII table, one code per line as decimal, hexadecimal, and " +
			"either the printable character or the control-code mnemonic (e.g. NUL, LF, ESC, DEL).",
		Examples: []command.Example{
			{Command: "ascii", Explain: "Print the full ASCII table."},
		},
		ExitStatus: "0  always.",
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	for code := 0; code < 128; code++ {
		_, _ = fmt.Fprintf(stdio.Out, "%3d  0x%02X  %s\n", code, code, repr(code))
	}
	return nil
}

// repr returns the printable character or control-code name for code.
func repr(code int) string {
	switch {
	case code < len(controlNames):
		return controlNames[code]
	case code == 32:
		return "SP"
	case code == 127:
		return "DEL"
	default:
		return string(rune(code))
	}
}
