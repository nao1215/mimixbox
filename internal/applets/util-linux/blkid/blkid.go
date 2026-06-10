// Package blkid implements the blkid applet: identify the filesystem type on a
// device or image file by probing for known superblock signatures.
package blkid

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the blkid applet.
type Command struct{}

// New returns a blkid command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "blkid" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Identify the filesystem type of a device or image" }

// signature is a filesystem magic at a fixed offset.
type signature struct {
	typ    string
	offset int64
	magic  []byte
}

// signatures are checked in order; the first match wins.
var signatures = []signature{
	{"xfs", 0, []byte("XFSB")},
	{"squashfs", 0, []byte("hsqs")},
	{"ntfs", 3, []byte("NTFS    ")},
	{"ext2", 0x438, []byte{0x53, 0xEF}},
	{"btrfs", 0x10040, []byte("_BHRfS_M")},
	{"swap", 0xFF6, []byte("SWAPSPACE2")},
	{"swap", 0xFF6, []byte("SWAP-SPACE")},
}

// Run executes blkid.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "FILE...", stdio.Err).WithHelp(command.Help{
		Description: "Print the filesystem TYPE of each FILE (a device or image), detected from its " +
			"superblock signature. Recognizes ext, xfs, btrfs, ntfs, squashfs, and swap. ext2/3/4 " +
			"are reported as ext2 (feature-flag version detection is not implemented).",
		Examples: []command.Example{
			{Command: "blkid disk.img", Explain: "Identify the filesystem in disk.img."},
		},
		ExitStatus: "0  a type was identified.\n2  nothing could be identified.\n1  a file could not be read.",
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	files := fs.Args()
	if len(files) == 0 {
		_, _ = fmt.Fprintln(stdio.Err, "blkid: a FILE is required")
		return command.SilentFailure()
	}

	identified := false
	readErr := false
	for _, name := range files {
		typ, err := probe(name)
		if err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "blkid: %s\n", command.FileError(name, err))
			readErr = true
			continue
		}
		if typ != "" {
			_, _ = fmt.Fprintf(stdio.Out, "%s: TYPE=%q\n", name, typ)
			identified = true
		}
	}

	if readErr {
		return command.SilentFailure()
	}
	if !identified {
		return &command.ExitError{Code: 2} // nothing identified, like blkid
	}
	return nil
}

// probe reads name and returns the first matching filesystem type, or "".
func probe(name string) (string, error) {
	f, err := os.Open(name) //nolint:gosec // user-named device/image
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()

	for _, s := range signatures {
		buf := make([]byte, len(s.magic))
		if _, err := f.ReadAt(buf, s.offset); err != nil {
			continue // too short to hold this signature
		}
		if bytes.Equal(buf, s.magic) {
			return s.typ, nil
		}
	}
	return "", nil
}
