// Package chattr implements the chattr applet: change ext2/ext4 file attributes.
package chattr

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
	"golang.org/x/sys/unix"
)

// Command is the chattr applet.
type Command struct{}

// New returns a chattr command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "chattr" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Change ext2/ext4 file attributes" }

// attrBits maps the chattr attribute letters to their inode flag bits.
var attrBits = map[byte]uint{
	's': 0x00000001, 'u': 0x00000002, 'c': 0x00000004, 'S': 0x00000008,
	'i': 0x00000010, 'a': 0x00000020, 'd': 0x00000040, 'A': 0x00000080,
	'I': 0x00001000, 'j': 0x00004000, 't': 0x00008000, 'D': 0x00010000,
	'T': 0x00020000, 'e': 0x00080000, 'C': 0x00800000,
}

// getFlags and setFlags are indirected so the logic is testable without ioctls.
var (
	getFlags = func(path string) (int, error) {
		f, err := os.Open(path) //nolint:gosec // user-named file
		if err != nil {
			return 0, err
		}
		defer func() { _ = f.Close() }()
		return unix.IoctlGetInt(int(f.Fd()), unix.FS_IOC_GETFLAGS)
	}
	setFlags = func(path string, flags int) error {
		f, err := os.Open(path) //nolint:gosec // user-named file
		if err != nil {
			return err
		}
		defer func() { _ = f.Close() }()
		return unix.IoctlSetPointerInt(int(f.Fd()), unix.FS_IOC_SETFLAGS, flags)
	}
)

// Run executes chattr.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "{+|-|=}ATTRS FILE...", stdio.Err).WithHelp(command.Help{
		Description: "Change the ext2/ext4 inode attributes of each FILE. The mode is +ATTRS to add, " +
			"-ATTRS to remove, or =ATTRS to set exactly. Each attribute is a letter such as i " +
			"(immutable), a (append-only), or A (no atime). Changing immutable/append attributes " +
			"requires privilege.",
		Examples: []command.Example{
			{Command: "chattr +i file", Explain: "Make file immutable."},
			{Command: "chattr -a file", Explain: "Clear the append-only attribute."},
		},
		ExitStatus: "0  all files were updated.\n1  the mode was invalid or a file could not be changed.",
	})
	// A "-ATTRS" mode looks like a flag to the parser; protect it with "--".
	if len(args) > 0 && strings.HasPrefix(args[0], "-") && args[0] != "--help" && args[0] != "--version" {
		args = append([]string{"--"}, args...)
	}

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	rest := fs.Args()
	if len(rest) < 2 {
		_, _ = fmt.Fprintln(stdio.Err, "chattr: a mode and at least one file are required")
		return command.SilentFailure()
	}

	op := rest[0][0]
	if op != '+' && op != '-' && op != '=' {
		return command.Failuref("invalid mode: %q (must start with +, - or =)", rest[0])
	}
	bits, err := parseAttrs(rest[0][1:])
	if err != nil {
		return command.Failuref("%v", err)
	}

	failed := false
	for _, path := range rest[1:] {
		cur, err := getFlags(path)
		if err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "chattr: %s: %v\n", path, err)
			failed = true
			continue
		}
		if err := setFlags(path, apply(op, cur, bits)); err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "chattr: %s: %v\n", path, err)
			failed = true
		}
	}
	if failed {
		return command.SilentFailure()
	}
	return nil
}

// parseAttrs turns a string of attribute letters into the combined flag bits.
func parseAttrs(letters string) (uint, error) {
	var bits uint
	for i := 0; i < len(letters); i++ {
		bit, ok := attrBits[letters[i]]
		if !ok {
			return 0, fmt.Errorf("unknown attribute: %q", string(letters[i]))
		}
		bits |= bit
	}
	return bits, nil
}

// apply computes the new flags from the operator, current flags, and bits.
func apply(op byte, cur int, bits uint) int {
	switch op {
	case '+':
		return int(uint(cur) | bits)
	case '-':
		return int(uint(cur) &^ bits)
	default: // '='
		return int(bits)
	}
}
