// Package lsattr implements the lsattr applet: list ext2/ext4 file attributes.
package lsattr

import (
	"context"
	"fmt"
	"os"

	"github.com/nao1215/mimixbox/internal/command"
	"golang.org/x/sys/unix"
)

// Command is the lsattr applet.
type Command struct{}

// New returns a lsattr command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "lsattr" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "List ext2/ext4 file attributes" }

// attrFlags maps the inode flag bits to their lsattr letters, in display order.
// The bit values are stable across the ext2/ext3/ext4 family.
var attrFlags = []struct {
	bit uint
	ch  byte
}{
	{0x00000001, 's'}, // secure deletion
	{0x00000002, 'u'}, // undelete
	{0x00000008, 'S'}, // synchronous updates
	{0x00010000, 'D'}, // synchronous directory updates
	{0x00000010, 'i'}, // immutable
	{0x00000020, 'a'}, // append only
	{0x00000040, 'd'}, // no dump
	{0x00000080, 'A'}, // no atime updates
	{0x00000004, 'c'}, // compressed
	{0x00000800, 'E'}, // encrypted
	{0x00004000, 'j'}, // data journaling
	{0x00001000, 'I'}, // hash-indexed directory
	{0x00008000, 't'}, // no tail-merging
	{0x00020000, 'T'}, // top of directory hierarchy
	{0x00080000, 'e'}, // extents
	{0x00800000, 'C'}, // no copy-on-write
}

// getFlags is indirected so the decoding can be tested without a real ioctl.
var getFlags = func(path string) (int, error) {
	f, err := os.Open(path) //nolint:gosec // user-named file
	if err != nil {
		return 0, err
	}
	defer func() { _ = f.Close() }()
	return unix.IoctlGetInt(int(f.Fd()), unix.FS_IOC_GETFLAGS)
}

// Run executes lsattr.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[FILE...]", stdio.Err).WithHelp(command.Help{
		Description: "List the ext2/ext4 inode attributes of each FILE as a string of attribute letters " +
			"(e.g. i immutable, a append-only, e extents), or '-' where the attribute is unset. With " +
			"no FILE, the entries of the current directory are listed.",
		Examples: []command.Example{
			{Command: "lsattr file.txt", Explain: "Show the attributes of file.txt."},
		},
		ExitStatus: "0  all files were read.\n1  a file's attributes could not be read.",
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	files := fs.Args()
	if len(files) == 0 {
		files = dirEntries(".")
	}

	failed := false
	for _, path := range files {
		flags, err := getFlags(path)
		if err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "lsattr: %s: %v\n", path, err)
			failed = true
			continue
		}
		_, _ = fmt.Fprintf(stdio.Out, "%s %s\n", decode(flags), path)
	}
	if failed {
		return command.SilentFailure()
	}
	return nil
}

// decode renders the attribute flags as the lsattr letter string.
func decode(flags int) string {
	b := make([]byte, len(attrFlags))
	for i, f := range attrFlags {
		if uint(flags)&f.bit != 0 {
			b[i] = f.ch
		} else {
			b[i] = '-'
		}
	}
	return string(b)
}

// dirEntries returns the names of the entries in dir (non-recursive).
func dirEntries(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names
}
