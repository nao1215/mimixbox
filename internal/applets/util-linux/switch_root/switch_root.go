// Package switchroot implements the switch_root applet: switch to another root
// filesystem and execute a new init. Intended to run as PID 1 in an initramfs.
package switchroot

import (
	"context"
	"os"
	"syscall"

	"github.com/nao1215/mimixbox/internal/command"
	"golang.org/x/sys/unix"
)

// Command is the switch_root applet.
type Command struct{}

// New returns a switch_root command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "switch_root" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Switch to another root and run init" }

// Injected so the privileged switch and the exec are testable.
var (
	// switchFn moves the current mount onto NEW_ROOT and chroots into it.
	switchFn = func(newRoot string) error {
		if err := os.Chdir(newRoot); err != nil {
			return err
		}
		if err := unix.Mount(newRoot, "/", "", unix.MS_MOVE, ""); err != nil {
			return err
		}
		if err := unix.Chroot("."); err != nil {
			return err
		}
		return os.Chdir("/")
	}
	// execFn replaces the current process with init; it does not return on success.
	execFn = func(path string, argv []string) error {
		return syscall.Exec(path, argv, os.Environ())
	}
)

// Run executes switch_root.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "NEW_ROOT INIT [ARG...]", stdio.Err).WithHelp(command.Help{
		Description: "Make NEW_ROOT the root filesystem of the current process and execute INIT (with " +
			"any ARGs) as the new init. This is meant to run as PID 1 from an initramfs and requires " +
			"privilege. The destructive removal of the old initramfs contents that the real switch_root " +
			"performs is intentionally not done by this build.",
		Examples: []command.Example{
			{Command: "switch_root /mnt/root /sbin/init", Explain: "Switch to /mnt/root and run init."},
		},
		ExitStatus: "1  the arguments were wrong or the switch failed (it does not return on success).",
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	rest := fs.Args()
	if len(rest) < 2 {
		return command.Failuref("a NEW_ROOT directory and an INIT program are required")
	}
	newRoot := rest[0]

	info, err := os.Stat(newRoot)
	if err != nil || !info.IsDir() {
		return command.Failuref("%s: not a directory", newRoot)
	}

	if err := switchFn(newRoot); err != nil {
		return command.Failuref("cannot switch root to %s: %v", newRoot, err)
	}

	// execFn normally replaces the process; reaching here means it failed.
	if err := execFn(rest[1], rest[1:]); err != nil {
		return command.Failuref("cannot exec %s: %v", rest[1], err)
	}
	return nil
}
