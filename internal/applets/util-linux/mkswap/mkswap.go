// Package mkswap implements the mkswap applet: set up a Linux swap area on a
// file or device.
package mkswap

import (
	"context"
	"encoding/binary"
	"fmt"
	"os"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the mkswap applet.
type Command struct{}

// New returns a mkswap command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "mkswap" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Set up a Linux swap area" }

// pageSize is the swap page size; tests may override it.
var pageSize = os.Getpagesize()

const (
	headerOffset = 1024 // swap_header_info begins here
	labelOffset  = 1052 // volume_name[16] within the header
	signature    = "SWAPSPACE2"
)

// Run executes mkswap.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-L LABEL] FILE", stdio.Err).WithHelp(command.Help{
		Description: "Write a version-1 Linux swap signature and header to FILE (a swap file or block " +
			"device), making it usable by swapon. -L sets a volume label. The file must already be at " +
			"least two pages in size; mkswap does not grow it.",
		Examples: []command.Example{
			{Command: "mkswap /swapfile", Explain: "Format an existing file as swap."},
			{Command: "mkswap -L swap /swapfile", Explain: "Format it with a label."},
		},
		ExitStatus: "0  the swap area was written.\n1  the file was missing or too small.",
	})
	label := fs.StringP("label", "L", "", "set the swap-area label")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	rest := fs.Args()
	if len(rest) == 0 {
		_, _ = fmt.Fprintln(stdio.Err, "mkswap: a file is required")
		return command.SilentFailure()
	}
	path := rest[0]

	f, err := os.OpenFile(path, os.O_RDWR, 0o600) //nolint:gosec // user-named swap file
	if err != nil {
		return command.Failuref("cannot open %s: %v", path, err)
	}
	defer func() { _ = f.Close() }()

	info, err := f.Stat()
	if err != nil {
		return command.Failuref("cannot stat %s: %v", path, err)
	}
	npages := int(info.Size()) / pageSize
	if npages < 2 {
		return command.Failuref("%s: swap area needs at least %d bytes", path, 2*pageSize)
	}
	lastPage := npages - 1

	if err := writeSwap(f, lastPage, *label); err != nil {
		return command.Failuref("%v", err)
	}

	usable := int64(lastPage) * int64(pageSize)
	_, _ = fmt.Fprintf(stdio.Out, "Setting up swapspace version 1, size = %d KiB (%d bytes)\n",
		usable/1024, usable)
	if *label != "" {
		_, _ = fmt.Fprintf(stdio.Out, "LABEL=%s\n", *label)
	}
	return nil
}

// writeSwap writes the version-1 header, optional label, and signature.
func writeSwap(f *os.File, lastPage int, label string) error {
	header := make([]byte, 12)
	binary.LittleEndian.PutUint32(header[0:], 1)                  // version
	binary.LittleEndian.PutUint32(header[4:], uint32(lastPage))   // last usable page
	binary.LittleEndian.PutUint32(header[8:], 0)                  // nr_badpages
	if _, err := f.WriteAt(header, headerOffset); err != nil {
		return err
	}
	if label != "" {
		buf := make([]byte, 16)
		copy(buf, label)
		if _, err := f.WriteAt(buf, labelOffset); err != nil {
			return err
		}
	}
	if _, err := f.WriteAt([]byte(signature), int64(pageSize-len(signature))); err != nil {
		return err
	}
	return nil
}
