// Package blockdev implements the blockdev applet: report block-device
// properties (size, sector size, block size, read-only flag) via ioctls.
package blockdev

import (
	"context"
	"fmt"
	"os"
	"unsafe"

	"github.com/nao1215/mimixbox/internal/command"
	"golang.org/x/sys/unix"
)

// Command is the blockdev applet.
type Command struct{}

// New returns a blockdev command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "blockdev" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Report block device properties" }

// blockQuery is indirected so the flag dispatch can be tested without a device.
var blockQuery = func(path, name string) (uint64, error) {
	f, err := os.OpenFile(path, os.O_RDONLY, 0) //nolint:gosec // user-named device
	if err != nil {
		return 0, err
	}
	defer func() { _ = f.Close() }()
	fd := int(f.Fd())

	switch name {
	case "getsize64":
		var sz uint64
		if _, _, errno := unix.Syscall(unix.SYS_IOCTL, uintptr(fd), uintptr(unix.BLKGETSIZE64), uintptr(unsafe.Pointer(&sz))); errno != 0 {
			return 0, errno
		}
		return sz, nil
	case "getss":
		v, err := unix.IoctlGetInt(fd, unix.BLKSSZGET)
		return uint64(v), err
	case "getbsz":
		v, err := unix.IoctlGetInt(fd, unix.BLKBSZGET)
		return uint64(v), err
	case "getro":
		v, err := unix.IoctlGetInt(fd, unix.BLKROGET)
		return uint64(v), err
	default:
		return 0, fmt.Errorf("unknown query %q", name)
	}
}

// queries lists the supported flags in display order.
var queries = []struct {
	flag, name, help string
}{
	{"getsize64", "getsize64", "print the device size in bytes"},
	{"getsz", "getsize64", "print the device size in bytes (alias of --getsize64)"},
	{"getss", "getss", "print the logical sector size in bytes"},
	{"getbsz", "getbsz", "print the block size in bytes"},
	{"getro", "getro", "print 1 if read-only, 0 otherwise"},
}

// Run executes blockdev.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[--getsize64] [--getss] [--getbsz] [--getro] DEVICE", stdio.Err).WithHelp(command.Help{
		Description: "Report properties of a block DEVICE. Each query flag prints one value: the " +
			"size in bytes, the logical sector size, the block size, or the read-only flag. Reading " +
			"a device usually requires privilege.",
		Examples: []command.Example{
			{Command: "blockdev --getsize64 /dev/sda", Explain: "Print the device size in bytes."},
		},
		ExitStatus: "0  success.\n1  no query was given, or the device could not be read.",
	})
	flags := map[string]*bool{}
	for _, q := range queries {
		flags[q.flag] = fs.Bool(q.flag, false, q.help)
	}

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	rest := fs.Args()
	if len(rest) == 0 {
		_, _ = fmt.Fprintln(stdio.Err, "blockdev: a device is required")
		return command.SilentFailure()
	}
	device := rest[0]

	any := false
	for _, q := range queries {
		if !*flags[q.flag] {
			continue
		}
		any = true
		v, err := blockQuery(device, q.name)
		if err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "blockdev: cannot read %s: %v\n", device, err)
			return command.SilentFailure()
		}
		_, _ = fmt.Fprintln(stdio.Out, v)
	}
	if !any {
		_, _ = fmt.Fprintln(stdio.Err, "blockdev: a query flag is required")
		return command.SilentFailure()
	}
	return nil
}
