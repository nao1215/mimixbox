// Package crc32 implements the crc32 applet: print the CRC-32 (IEEE) checksum of
// each file (or standard input) as an 8-digit hexadecimal value.
package crc32

import (
	"context"
	"fmt"
	"hash/crc32"
	"io"
	"os"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the crc32 applet.
type Command struct{}

// New returns a crc32 command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "crc32" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Print the CRC-32 checksum of each file" }

// Run executes crc32.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[FILE]...", stdio.Err).WithHelp(command.Help{
		Description: "Print the CRC-32 (IEEE) checksum of each FILE as eight hexadecimal digits, " +
			"followed by the file name. With no FILE, or when FILE is -, read standard input.",
		Examples: []command.Example{
			{Command: "crc32 file.bin", Explain: "Print: <crc32 hex>  file.bin"},
		},
		ExitStatus: "0  success.\n1  a file could not be read.",
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	files := fs.Args()
	if len(files) == 0 {
		files = []string{"-"}
	}

	var failed bool
	for _, name := range files {
		sum, rerr := crcOf(stdio, name)
		if rerr != nil {
			_, _ = fmt.Fprintf(stdio.Err, "crc32: %s\n", command.FileError(name, rerr))
			failed = true
			continue
		}
		_, _ = fmt.Fprintf(stdio.Out, "%08x  %s\n", sum, name)
	}
	if failed {
		return command.SilentFailure()
	}
	return nil
}

func crcOf(stdio command.IO, name string) (uint32, error) {
	r := stdio.In
	if name != "-" {
		f, err := os.Open(name) //nolint:gosec // user-named file
		if err != nil {
			return 0, err
		}
		defer func() { _ = f.Close() }()
		r = f
	}
	h := crc32.NewIEEE()
	if _, err := io.Copy(h, r); err != nil {
		return 0, err
	}
	return h.Sum32(), nil
}
