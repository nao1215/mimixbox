// Package sum implements the sum applet: print a BSD-style 16-bit checksum and
// the 1024-byte block count for each file (or standard input).
package sum

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the sum applet.
type Command struct{}

// New returns a sum command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "sum" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Checksum and count the blocks in a file (BSD)" }

// Run executes sum.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[FILE]...", stdio.Err).WithHelp(command.Help{
		Description: "Print a checksum and the number of 1024-byte blocks for each FILE, using the " +
			"BSD algorithm. With no FILE, or when FILE is -, read standard input.",
		Examples: []command.Example{
			{Command: "sum file.txt", Explain: "Print: <checksum> <blocks> file.txt"},
		},
		ExitStatus: "0  success.\n1  a file could not be read.",
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	files := fs.Args()
	if len(files) == 0 {
		s, blocks, rerr := bsdSum(stdio.In)
		if rerr != nil {
			_, _ = fmt.Fprintf(stdio.Err, "sum: %v\n", rerr)
			return command.SilentFailure()
		}
		_, _ = fmt.Fprintf(stdio.Out, "%05d %5d\n", s, blocks)
		return nil
	}

	var failed bool
	for _, name := range files {
		s, blocks, rerr := sumFile(stdio, name)
		if rerr != nil {
			_, _ = fmt.Fprintf(stdio.Err, "sum: %s\n", command.FileError(name, rerr))
			failed = true
			continue
		}
		_, _ = fmt.Fprintf(stdio.Out, "%05d %5d %s\n", s, blocks, name)
	}
	if failed {
		return command.SilentFailure()
	}
	return nil
}

func sumFile(stdio command.IO, name string) (uint16, int64, error) {
	if name == "-" {
		return bsdSum(stdio.In)
	}
	f, err := os.Open(name) //nolint:gosec // user-named file
	if err != nil {
		return 0, 0, err
	}
	defer func() { _ = f.Close() }()
	return bsdSum(f)
}

// bsdSum computes the BSD 16-bit rotate checksum and the number of 1024-byte
// blocks (rounded up) of r.
func bsdSum(r io.Reader) (uint16, int64, error) {
	var sum uint32
	var total int64
	buf := make([]byte, 32*1024)
	for {
		n, err := r.Read(buf)
		for _, b := range buf[:n] {
			sum = (sum >> 1) | ((sum & 1) << 15)
			sum = (sum + uint32(b)) & 0xffff
		}
		total += int64(n)
		if err != nil {
			if err == io.EOF {
				break
			}
			return 0, 0, err
		}
	}
	blocks := (total + 1023) / 1024
	return uint16(sum), blocks, nil
}
