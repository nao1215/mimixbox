// Package setpriv implements the setpriv applet: run a program with changed
// privilege settings, or dump the current ones.
package setpriv

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"

	"github.com/nao1215/mimixbox/internal/command"
	"golang.org/x/sys/unix"
)

// Command is the setpriv applet.
type Command struct{}

// New returns a setpriv command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "setpriv" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Run a program with different privilege settings" }

const prSetNoNewPrivs = 38

// Injected so the credential queries are controllable in tests.
var (
	getuid     = os.Getuid
	geteuid    = os.Geteuid
	getgid     = os.Getgid
	getegid    = os.Getegid
	getgroups  = os.Getgroups
	noNewPrivs = func() int {
		v, err := unix.PrctlRetInt(unix.PR_GET_NO_NEW_PRIVS, 0, 0, 0, 0)
		if err != nil {
			return 0
		}
		return v
	}
)

// Run executes setpriv.
func (c *Command) Run(ctx context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[--dump | OPTIONS COMMAND [ARG]...]", stdio.Err).WithHelp(command.Help{
		Description: "Run COMMAND with changed privileges, or with --dump print the current ones. " +
			"Options: --reuid UID and --regid GID change the credentials (needs privilege), and " +
			"--no-new-privs prevents the program from gaining new privileges.",
		Examples: []command.Example{
			{Command: "setpriv --dump", Explain: "Show the current uid/gid and privileges."},
			{Command: "setpriv --reuid 1000 --regid 1000 -- id", Explain: "Run id as another user."},
		},
		ExitStatus: "0  success.\n1  an invalid option, or the command could not be run.",
	})
	dump := fs.BoolP("dump", "d", false, "show the current privileges and exit")
	reuid := fs.Int("reuid", -1, "set the real and effective user ID")
	regid := fs.Int("regid", -1, "set the real and effective group ID")
	nnp := fs.Bool("no-new-privs", false, "disallow gaining new privileges")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	if *dump {
		c.dump(stdio.Out)
		return nil
	}

	rest := fs.Args()
	if len(rest) > 0 && rest[0] == "--" {
		rest = rest[1:]
	}
	if len(rest) == 0 {
		_, _ = fmt.Fprintln(stdio.Err, "setpriv: a command is required")
		return command.SilentFailure()
	}

	if *nnp {
		if _, err := unix.PrctlRetInt(prSetNoNewPrivs, 1, 0, 0, 0); err != nil {
			return command.Failuref("cannot set no_new_privs: %v", err)
		}
	}

	cmd := exec.CommandContext(ctx, rest[0], rest[1:]...) //nolint:gosec // running the user's command is the point
	cmd.Stdin, cmd.Stdout, cmd.Stderr = stdio.In, stdio.Out, stdio.Err
	if *reuid >= 0 || *regid >= 0 {
		cred := &syscall.Credential{Uid: uint32(getuid()), Gid: uint32(getgid())}
		if *reuid >= 0 {
			cred.Uid = uint32(*reuid)
		}
		if *regid >= 0 {
			cred.Gid = uint32(*regid)
		}
		cmd.SysProcAttr = &syscall.SysProcAttr{Credential: cred}
	}

	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return &command.ExitError{Code: exitErr.ExitCode()}
		}
		return command.Failuref("%s: %v", rest[0], err)
	}
	return nil
}

// dump prints the current credentials, mirroring setpriv --dump.
func (c *Command) dump(out io.Writer) {
	_, _ = fmt.Fprintf(out, "uid: %d\n", getuid())
	_, _ = fmt.Fprintf(out, "euid: %d\n", geteuid())
	_, _ = fmt.Fprintf(out, "gid: %d\n", getgid())
	_, _ = fmt.Fprintf(out, "egid: %d\n", getegid())
	groups, _ := getgroups()
	_, _ = fmt.Fprintf(out, "Supplementary groups: %s\n", joinInts(groups))
	_, _ = fmt.Fprintf(out, "no_new_privs: %d\n", noNewPrivs())
}

func joinInts(ns []int) string {
	if len(ns) == 0 {
		return "[none]"
	}
	parts := make([]string, len(ns))
	for i, n := range ns {
		parts[i] = strconv.Itoa(n)
	}
	return strings.Join(parts, ",")
}
