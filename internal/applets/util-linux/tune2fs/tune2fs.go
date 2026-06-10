// Package tune2fs implements the tune2fs applet: display ext2/ext3/ext4
// filesystem parameters from a superblock. Adjusting parameters is not done by
// this slice.
package tune2fs

import (
	"context"
	"encoding/binary"
	"fmt"
	"os"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the tune2fs applet.
type Command struct{}

// New returns a tune2fs command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "tune2fs" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Show ext2/ext3/ext4 filesystem parameters" }

const (
	superblockOffset = 1024   // the ext2 superblock starts here
	extMagic         = 0xEF53 // s_magic
)

// Run executes tune2fs.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "-l FILE", stdio.Err).WithHelp(command.Help{
		Description: "List the parameters of the ext2/ext3/ext4 filesystem in FILE (a device or image): " +
			"its volume name, UUID, and inode/block counts and block size, read from the superblock. " +
			"Only the read-only listing (-l) is implemented; changing parameters is not.",
		Examples: []command.Example{
			{Command: "tune2fs -l disk.img", Explain: "Show the filesystem parameters."},
		},
		ExitStatus: "0  the parameters were listed.\n1  the file was unreadable or not an ext filesystem.",
	})
	list := fs.BoolP("list", "l", false, "list the filesystem superblock contents")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	rest := fs.Args()
	if len(rest) == 0 {
		return command.Failuref("a filesystem (device or image) is required")
	}
	if !*list {
		_, _ = fmt.Fprintln(stdio.Err, "tune2fs: only the read-only listing (-l) is supported by this build")
		return command.SilentFailure()
	}

	sb, err := readSuperblock(rest[0])
	if err != nil {
		return command.Failuref("%s: %v", rest[0], err)
	}

	_, _ = fmt.Fprintf(stdio.Out, "Filesystem volume name:   %s\n", orNone(sb.volumeName))
	_, _ = fmt.Fprintf(stdio.Out, "Filesystem UUID:          %s\n", sb.uuid)
	_, _ = fmt.Fprintf(stdio.Out, "Inode count:              %d\n", sb.inodes)
	_, _ = fmt.Fprintf(stdio.Out, "Block count:              %d\n", sb.blocks)
	_, _ = fmt.Fprintf(stdio.Out, "Block size:               %d\n", sb.blockSize)
	return nil
}

type superblock struct {
	inodes, blocks uint32
	blockSize      int
	uuid           string
	volumeName     string
}

// readSuperblock reads and parses the ext2 superblock from path.
func readSuperblock(path string) (*superblock, error) {
	f, err := os.Open(path) //nolint:gosec // user-named device or image
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	buf := make([]byte, 264) // up to s_volume_name (offset 120) + 16
	if _, err := f.ReadAt(buf, superblockOffset); err != nil {
		return nil, fmt.Errorf("cannot read superblock: %w", err)
	}
	if binary.LittleEndian.Uint16(buf[56:]) != extMagic {
		return nil, fmt.Errorf("not an ext2/3/4 filesystem (bad magic)")
	}

	return &superblock{
		inodes:     binary.LittleEndian.Uint32(buf[0:]),
		blocks:     binary.LittleEndian.Uint32(buf[4:]),
		blockSize:  1024 << binary.LittleEndian.Uint32(buf[24:]),
		uuid:       formatUUID(buf[104:120]),
		volumeName: strings.TrimRight(string(buf[120:136]), "\x00"),
	}, nil
}

func formatUUID(b []byte) string {
	if len(b) < 16 {
		return ""
	}
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

func orNone(s string) string {
	if s == "" {
		return "<none>"
	}
	return s
}
