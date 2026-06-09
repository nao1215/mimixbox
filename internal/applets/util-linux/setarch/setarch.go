// Package setarch implements the setarch, linux32, and linux64 applets: run a
// program with a changed execution domain (personality), most usefully so that
// uname reports a 32-bit machine on a 64-bit kernel.
package setarch

import (
	"context"
	"errors"
	"fmt"
	"os/exec"

	"github.com/nao1215/mimixbox/internal/command"
	"golang.org/x/sys/unix"
)

// Linux execution-domain personalities (not exported by x/sys/unix here).
const (
	perLinux   uintptr = 0x0000
	perLinux32 uintptr = 0x0008
)

// setPersonality is indirected so tests do not change the test process's domain.
var setPersonality = func(p uintptr) error {
	if _, _, errno := unix.Syscall(unix.SYS_PERSONALITY, p, 0, 0); errno != 0 {
		return errno
	}
	return nil
}

// Command is the setarch/linux32/linux64 applet.
type Command struct{ name string }

// NewSetarch returns the setarch applet.
func NewSetarch() *Command { return &Command{name: "setarch"} }

// NewLinux32 returns the linux32 applet.
func NewLinux32() *Command { return &Command{name: "linux32"} }

// NewLinux64 returns the linux64 applet.
func NewLinux64() *Command { return &Command{name: "linux64"} }

// Name returns the command name.
func (c *Command) Name() string { return c.name }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string {
	switch c.name {
	case "linux32":
		return "Run a program with a 32-bit execution domain"
	case "linux64":
		return "Run a program with a 64-bit execution domain"
	default:
		return "Run a program with a changed architecture personality"
	}
}

// Run executes the applet.
func (c *Command) Run(ctx context.Context, stdio command.IO, args []string) error {
	usage := "[-c] COMMAND [ARG]..."
	if c.name == "setarch" {
		usage = "ARCH [-c] COMMAND [ARG]..."
	}
	fs := command.NewFlagSet(c.Name(), usage, stdio.Err).WithHelp(command.Help{
		Description: "Run COMMAND with a changed execution-domain personality. linux32 selects a " +
			"32-bit domain (so uname reports a 32-bit machine), linux64 a 64-bit one, and setarch " +
			"selects it from the leading ARCH name (e.g. i686, x86_64).",
		Examples: []command.Example{
			{Command: "linux32 uname -m", Explain: "Report the 32-bit machine name."},
			{Command: "setarch i686 ./configure", Explain: "Configure as if on a 32-bit host."},
		},
		ExitStatus: "The exit status of COMMAND (127 if it could not be run).",
	})
	fs.SetInterspersed(false)
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	rest := fs.Args()
	persona := perLinux32
	if c.name == "linux64" {
		persona = perLinux
	}
	if c.name == "setarch" {
		if len(rest) == 0 {
			_, _ = fmt.Fprintln(stdio.Err, "setarch: missing ARCH operand")
			return command.SilentFailure()
		}
		persona = personaForArch(rest[0])
		rest = rest[1:]
	}

	if len(rest) == 0 {
		_, _ = fmt.Fprintf(stdio.Err, "%s: missing command\n", c.Name())
		return command.SilentFailure()
	}

	if err := setPersonality(persona); err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "%s: cannot set personality: %v\n", c.Name(), err)
		return command.SilentFailure()
	}

	cmd := exec.CommandContext(ctx, rest[0], rest[1:]...) //nolint:gosec // running the user's command is the point
	cmd.Stdin, cmd.Stdout, cmd.Stderr = stdio.In, stdio.Out, stdio.Err
	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return &command.ExitError{Code: exitErr.ExitCode()}
		}
		_, _ = fmt.Fprintf(stdio.Err, "%s: %s: %v\n", c.Name(), rest[0], err)
		return &command.ExitError{Code: 127}
	}
	return nil
}

// personaForArch maps an architecture name to its personality.
func personaForArch(arch string) uintptr {
	switch arch {
	case "linux32", "i386", "i486", "i586", "i686", "x86":
		return perLinux32
	default:
		return perLinux
	}
}
