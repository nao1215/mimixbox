// Package softlimit implements the softlimit applet: run a program under a set
// of resource limits.
package softlimit

import (
	"context"
	"errors"
	"os/exec"

	"github.com/nao1215/mimixbox/internal/command"
	"golang.org/x/sys/unix"
)

// Command is the softlimit applet.
type Command struct{}

// New returns a softlimit command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "softlimit" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Run a program under resource limits" }

// setRlimitFn is indirected so the limit changes can be tested without altering
// the test process. It sets both the soft and hard limit to value.
var setRlimitFn = func(resource int, value uint64) error {
	return unix.Setrlimit(resource, &unix.Rlimit{Cur: value, Max: value})
}

// limitFlag binds a command-line flag to an rlimit resource.
type limitFlag struct {
	short    string
	resource int
	help     string
}

var limitFlags = []limitFlag{
	{"m", unix.RLIMIT_AS, "limit total address space to BYTES"},
	{"d", unix.RLIMIT_DATA, "limit the data segment to BYTES"},
	{"s", unix.RLIMIT_STACK, "limit the stack to BYTES"},
	{"o", unix.RLIMIT_NOFILE, "limit open file descriptors to N"},
	{"p", unix.RLIMIT_NPROC, "limit processes to N"},
	{"f", unix.RLIMIT_FSIZE, "limit created file size to BYTES"},
	{"c", unix.RLIMIT_CORE, "limit core dump size to BYTES"},
	{"t", unix.RLIMIT_CPU, "limit CPU time to SECONDS"},
}

// Run executes softlimit.
func (c *Command) Run(ctx context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-m BYTES] [-o N] [...] PROG [ARG...]", stdio.Err).WithHelp(command.Help{
		Description: "Set resource limits, then run PROG with its arguments under them: -m total " +
			"address space, -d data segment, -s stack, -f file size, -c core size (bytes); -o open " +
			"files, -p processes (counts); -t CPU seconds. Each flag sets the soft (and hard) limit. " +
			"This is the daemontools/runit softlimit.",
		Examples: []command.Example{
			{Command: "softlimit -m 100000000 -o 64 mydaemon", Explain: "Cap memory and open files."},
		},
		ExitStatus: "PROG's exit status, or 1 on a usage error or a limit that could not be set.",
	})
	values := make(map[int]*int64, len(limitFlags))
	for _, lf := range limitFlags {
		values[lf.resource] = fs.Int64P("limit-"+lf.short, lf.short, -1, lf.help)
	}
	fs.SetInterspersed(false)

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	rest := fs.Args()
	if len(rest) == 0 {
		return command.Failuref("a program is required")
	}

	for _, lf := range limitFlags {
		if v := *values[lf.resource]; v >= 0 {
			if err := setRlimitFn(lf.resource, uint64(v)); err != nil {
				return command.Failuref("cannot set -%s limit: %v", lf.short, err)
			}
		}
	}

	cmd := exec.CommandContext(ctx, rest[0], rest[1:]...) //nolint:gosec // running the user's program is the point
	cmd.Stdin, cmd.Stdout, cmd.Stderr = stdio.In, stdio.Out, stdio.Err
	if err := cmd.Run(); err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			return &command.ExitError{Code: ee.ExitCode()}
		}
		return command.Failuref("%v", err)
	}
	return nil
}
