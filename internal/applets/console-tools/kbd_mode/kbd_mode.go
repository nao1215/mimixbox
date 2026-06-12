// Package kbdmode implements the kbd_mode applet: report or set the console
// keyboard mode.
package kbdmode

import (
	"context"
	"fmt"
	"os"
	"unsafe"

	"github.com/nao1215/mimixbox/internal/command"
	"golang.org/x/sys/unix"
)

// Command is the kbd_mode applet.
type Command struct{}

// New returns a kbd_mode command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "kbd_mode" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Report or set the keyboard mode" }

// Keyboard modes and the ioctls to get/set them.
const (
	kdgkbMode  = 0x4B44
	kdskbMode  = 0x4B45
	kRaw       = 0
	kXlate     = 1
	kMediumRaw = 2
	kUnicode   = 3
)

// modeNames describes each keyboard mode for the report output.
var modeNames = map[int]string{
	kRaw:       "raw (scancode)",
	kXlate:     "ASCII",
	kMediumRaw: "mediumraw (keycode)",
	kUnicode:   "Unicode (UTF-8)",
}

// Injected so the keyboard mode can be read/set without a console.
var (
	getModeFn = func() (int, error) {
		f, err := os.Open("/dev/tty") //nolint:gosec // the controlling terminal
		if err != nil {
			return 0, err
		}
		defer func() { _ = f.Close() }()
		var mode int
		if _, _, errno := unix.Syscall(unix.SYS_IOCTL, f.Fd(), kdgkbMode, uintptr(unsafe.Pointer(&mode))); errno != 0 {
			return 0, errno
		}
		return mode, nil
	}
	setModeFn = func(mode int) error {
		f, err := os.Open("/dev/tty") //nolint:gosec // the controlling terminal
		if err != nil {
			return err
		}
		defer func() { _ = f.Close() }()
		if _, _, errno := unix.Syscall(unix.SYS_IOCTL, f.Fd(), kdskbMode, uintptr(mode)); errno != 0 {
			return errno
		}
		return nil
	}
)

// Run executes kbd_mode.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-a|-u|-k|-s]", stdio.Err).WithHelp(command.Help{
		Description: "With no option, report the current keyboard mode of the console. With an option, " +
			"set it: -a ASCII (XLATE), -u Unicode (UTF-8), -k mediumraw (keycodes), -s raw (scancodes). " +
			"Setting the mode requires privilege and a Linux console.",
		Examples: []command.Example{
			{Command: "kbd_mode", Explain: "Report the current keyboard mode."},
			{Command: "kbd_mode -u", Explain: "Switch to Unicode mode."},
		},
		ExitStatus: "0  the mode was reported or set.\n1  conflicting options or the console was inaccessible.",
	})
	ascii := fs.BoolP("ascii", "a", false, "set ASCII (XLATE) mode")
	unicode := fs.BoolP("unicode", "u", false, "set Unicode (UTF-8) mode")
	keycode := fs.BoolP("keycode", "k", false, "set mediumraw (keycode) mode")
	scancode := fs.BoolP("scancode", "s", false, "set raw (scancode) mode")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	mode, set, n := -1, false, 0
	for flag, m := range map[*bool]int{ascii: kXlate, unicode: kUnicode, keycode: kMediumRaw, scancode: kRaw} {
		if *flag {
			mode, set = m, true
			n++
		}
	}
	if n > 1 {
		return command.Failuref("only one mode option may be given")
	}

	if set {
		if err := setModeFn(mode); err != nil {
			return command.Failuref("cannot set the keyboard mode: %v", err)
		}
		return nil
	}

	cur, err := getModeFn()
	if err != nil {
		return command.Failuref("cannot read the keyboard mode: %v", err)
	}
	name := modeNames[cur]
	if name == "" {
		name = "unknown"
	}
	_, _ = fmt.Fprintf(stdio.Out, "The keyboard is in %s mode\n", name)
	return nil
}
