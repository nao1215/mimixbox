// Package fatattr implements the fatattr applet: display or change the
// attributes of files on a FAT filesystem.
package fatattr

import (
	"context"
	"fmt"
	"os"
	"strings"
	"unsafe"

	"github.com/nao1215/mimixbox/internal/command"
	"golang.org/x/sys/unix"
)

// Command is the fatattr applet.
type Command struct{}

// New returns a fatattr command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "fatattr" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Show or change FAT file attributes" }

// FAT_IOCTL_GET/SET_ATTRIBUTES, not exported by this x/sys version.
const (
	fatGetAttrs = 0x80047210
	fatSetAttrs = 0x40047211
)

// attrBits maps the fatattr letters to the FAT attribute bits, in display order.
var attrBits = []struct {
	ch  byte
	bit uint32
}{
	{'r', 0x01}, // read only
	{'h', 0x02}, // hidden
	{'s', 0x04}, // system
	{'v', 0x08}, // volume label
	{'d', 0x10}, // directory
	{'a', 0x20}, // archive
}

// Injected so the logic is testable without a FAT filesystem.
var (
	getAttrFn = func(path string) (uint32, error) {
		f, err := os.Open(path) //nolint:gosec // user-named file
		if err != nil {
			return 0, err
		}
		defer func() { _ = f.Close() }()
		var attr uint32
		_, _, errno := unix.Syscall(unix.SYS_IOCTL, f.Fd(), fatGetAttrs, uintptr(unsafe.Pointer(&attr)))
		if errno != 0 {
			return 0, errno
		}
		return attr, nil
	}
	setAttrFn = func(path string, attr uint32) error {
		f, err := os.Open(path) //nolint:gosec // user-named file
		if err != nil {
			return err
		}
		defer func() { _ = f.Close() }()
		_, _, errno := unix.Syscall(unix.SYS_IOCTL, f.Fd(), fatSetAttrs, uintptr(unsafe.Pointer(&attr)))
		if errno != 0 {
			return errno
		}
		return nil
	}
)

// Run executes fatattr.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[+|-attrs]... FILE...", stdio.Err).WithHelp(command.Help{
		Description: "Show or change the attributes of files on a FAT filesystem. With no +/- argument " +
			"each FILE's attributes are printed as a letter string (r read-only, h hidden, s system, " +
			"v volume, d directory, a archive). +X adds and -X removes the listed attributes.",
		Examples: []command.Example{
			{Command: "fatattr file.txt", Explain: "Show the attributes of file.txt."},
			{Command: "fatattr +r -a file.txt", Explain: "Set read-only and clear archive."},
		},
		ExitStatus: "0  all files were handled.\n1  an unknown attribute or an I/O error.",
	})
	// Leading "-X" looks like a flag; protect it and stop interspersed parsing so
	// the +/-attr tokens and file names are all treated as operands.
	if len(args) > 0 && strings.HasPrefix(args[0], "-") && args[0] != "--help" && args[0] != "--version" {
		args = append([]string{"--"}, args...)
	}
	fs.SetInterspersed(false)

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	var addBits, removeBits uint32
	var files []string
	modify := false
	for _, a := range fs.Args() {
		if (strings.HasPrefix(a, "+") || strings.HasPrefix(a, "-")) && len(a) > 1 {
			bits, err := parseAttrs(a[1:])
			if err != nil {
				return command.Failuref("%v", err)
			}
			if a[0] == '+' {
				addBits |= bits
			} else {
				removeBits |= bits
			}
			modify = true
			continue
		}
		files = append(files, a)
	}
	if len(files) == 0 {
		return command.Failuref("a file is required")
	}

	failed := false
	for _, file := range files {
		cur, err := getAttrFn(file)
		if err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "fatattr: %s: %v\n", file, err)
			failed = true
			continue
		}
		if !modify {
			_, _ = fmt.Fprintf(stdio.Out, "%s %s\n", decode(cur), file)
			continue
		}
		if err := setAttrFn(file, (cur|addBits)&^removeBits); err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "fatattr: %s: %v\n", file, err)
			failed = true
		}
	}
	if failed {
		return command.SilentFailure()
	}
	return nil
}

// parseAttrs turns attribute letters into their combined bits.
func parseAttrs(letters string) (uint32, error) {
	var bits uint32
	for i := 0; i < len(letters); i++ {
		bit, ok := bitFor(letters[i])
		if !ok {
			return 0, fmt.Errorf("unknown attribute: %q", string(letters[i]))
		}
		bits |= bit
	}
	return bits, nil
}

func bitFor(ch byte) (uint32, bool) {
	for _, a := range attrBits {
		if a.ch == ch {
			return a.bit, true
		}
	}
	return 0, false
}

// decode renders the attribute bits as the fatattr letter string.
func decode(attr uint32) string {
	b := make([]byte, len(attrBits))
	for i, a := range attrBits {
		if attr&a.bit != 0 {
			b[i] = a.ch
		} else {
			b[i] = '-'
		}
	}
	return string(b)
}
