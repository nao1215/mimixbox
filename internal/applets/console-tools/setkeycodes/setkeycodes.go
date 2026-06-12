// Package setkeycodes implements the setkeycodes applet: map scancodes to
// keycodes in the kernel keyboard driver.
package setkeycodes

import (
	"context"
	"strconv"
	"unsafe"

	"github.com/nao1215/mimixbox/internal/command"
	"golang.org/x/sys/unix"
	"os"
)

// Command is the setkeycodes applet.
type Command struct{}

// New returns a setkeycodes command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "setkeycodes" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Map scancodes to keycodes" }

// kdSetKeycode is the KDSETKEYCODE ioctl, not exported by this x/sys version.
const kdSetKeycode = 0x4B4D

// kbkeycode mirrors struct kbkeycode.
type kbkeycode struct {
	scancode, keycode uint32
}

// setKeycodeFn is indirected so the ioctl can be tested without a console.
var setKeycodeFn = func(scancode, keycode int) error {
	f, err := os.Open("/dev/console") //nolint:gosec // the system console
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	kc := kbkeycode{scancode: uint32(scancode), keycode: uint32(keycode)}
	if _, _, errno := unix.Syscall(unix.SYS_IOCTL, f.Fd(), kdSetKeycode, uintptr(unsafe.Pointer(&kc))); errno != 0 {
		return errno
	}
	return nil
}

// Run executes setkeycodes.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "SCANCODE KEYCODE...", stdio.Err).WithHelp(command.Help{
		Description: "Map raw keyboard SCANCODEs to kernel KEYCODEs (via the KDSETKEYCODE ioctl), given " +
			"as one or more pairs. The scancode is hexadecimal (e.g. e060) and the keycode is decimal. " +
			"Requires privilege and a Linux console.",
		Examples: []command.Example{
			{Command: "setkeycodes e060 122", Explain: "Map scancode 0xe060 to keycode 122."},
		},
		ExitStatus: "0  the mappings were set.\n1  an odd or invalid argument, or the ioctl failed.",
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	rest := fs.Args()
	if len(rest) == 0 || len(rest)%2 != 0 {
		return command.Failuref("scancode/keycode arguments must be given in pairs")
	}

	for i := 0; i < len(rest); i += 2 {
		scancode, err := strconv.ParseInt(rest[i], 16, 32)
		if err != nil || scancode < 0 {
			return command.Failuref("invalid scancode: %q", rest[i])
		}
		keycode, err := strconv.Atoi(rest[i+1])
		if err != nil || keycode < 0 {
			return command.Failuref("invalid keycode: %q", rest[i+1])
		}
		if err := setKeycodeFn(int(scancode), keycode); err != nil {
			return command.Failuref("cannot set scancode %s: %v", rest[i], err)
		}
	}
	return nil
}
