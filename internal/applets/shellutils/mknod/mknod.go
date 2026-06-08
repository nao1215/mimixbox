// Package mknod implements the mknod applet: create a FIFO, character special,
// or block special file. Creating device nodes (types b, c and u) needs
// privileges; a FIFO (type p) can be created by any user.
package mknod

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/nao1215/mimixbox/internal/command"
	"golang.org/x/sys/unix"
)

// Command is the mknod applet.
type Command struct{}

// New returns an mknod command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "mknod" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Make block or character special files" }

// Run executes mknod.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... NAME TYPE [MAJOR MINOR]", stdio.Err)
	modeStr := fs.StringP("mode", "m", "666", "set file permission bits to MODE, not a=rw - umask")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	operands := fs.Args()
	if len(operands) < 2 {
		_, _ = fmt.Fprintln(stdio.Err, "mknod: missing operand")
		_, _ = fmt.Fprintln(stdio.Err, "Try 'mknod --help' for more information.")
		return command.SilentFailure()
	}

	mode, err := parseMode(*modeStr)
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "mknod: invalid mode '%s'\n", *modeStr)
		return command.SilentFailure()
	}

	name, typ := operands[0], operands[1]
	if err := makeNode(name, typ, mode, operands[2:]); err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "mknod: %v\n", err)
		return command.SilentFailure()
	}
	return nil
}

// parseMode interprets an octal permission string such as "666" or "0644".
func parseMode(s string) (uint32, error) {
	v, err := strconv.ParseUint(s, 8, 32)
	if err != nil {
		return 0, err
	}
	return uint32(v), nil
}

// makeNode creates the special file described by typ. For p (FIFO) no device
// numbers are allowed; for b, c and u the MAJOR and MINOR operands are
// required. After creation the exact mode is applied so the result is not
// reduced by the process umask.
func makeNode(name, typ string, mode uint32, devArgs []string) error {
	var fileType uint32
	needDev := true
	switch typ {
	case "p":
		fileType = unix.S_IFIFO
		needDev = false
	case "b":
		fileType = unix.S_IFBLK
	case "c", "u":
		fileType = unix.S_IFCHR
	default:
		return fmt.Errorf("invalid device type '%s'", typ)
	}

	dev := 0
	if needDev {
		if len(devArgs) != 2 {
			return fmt.Errorf("type '%s' requires MAJOR and MINOR device numbers", typ)
		}
		major, err := parseDevNum(devArgs[0])
		if err != nil {
			return fmt.Errorf("invalid major device number '%s'", devArgs[0])
		}
		minor, err := parseDevNum(devArgs[1])
		if err != nil {
			return fmt.Errorf("invalid minor device number '%s'", devArgs[1])
		}
		dev = int(unix.Mkdev(major, minor))
	} else if len(devArgs) != 0 {
		return fmt.Errorf("type '%s' must not have device numbers", typ)
	}

	if err := unix.Mknod(name, fileType|mode, dev); err != nil {
		return fmt.Errorf("cannot create '%s': %w", name, err)
	}
	// Mknod honors the umask, so set the requested mode explicitly.
	if err := os.Chmod(name, os.FileMode(mode)); err != nil {
		return fmt.Errorf("cannot set mode of '%s': %w", name, err)
	}
	return nil
}

// parseDevNum parses a device number in decimal, hex (0x...) or octal (0...).
func parseDevNum(s string) (uint32, error) {
	v, err := strconv.ParseUint(s, 0, 32)
	if err != nil {
		return 0, err
	}
	return uint32(v), nil
}
