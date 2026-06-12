// Package runsv implements the runsv applet: supervise a single service by
// running its ./run script and restarting it when it exits.
package runsv

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the runsv applet.
type Command struct{}

// New returns a runsv command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "runsv" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Supervise a single service" }

// Injected so the run step and the restart delay are testable.
var (
	restartDelay = time.Second
	runOnceFn    = func(ctx context.Context, dir string, stdio command.IO) error {
		cmd := exec.CommandContext(ctx, "./run") //nolint:gosec // running the service's run script is the point
		cmd.Dir = dir
		cmd.Stdin, cmd.Stdout, cmd.Stderr = stdio.In, stdio.Out, stdio.Err
		return cmd.Run()
	}
)

// Run executes runsv.
func (c *Command) Run(ctx context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "DIR", stdio.Err).WithHelp(command.Help{
		Description: "Supervise the service directory DIR: run its ./run script and restart it whenever " +
			"it exits, until interrupted. While supervising, it maintains DIR/supervise/ok and the " +
			"control file so that sv/svok can see the service. If DIR/down exists, the service is left " +
			"stopped (but still supervised).",
		Examples: []command.Example{
			{Command: "runsv /etc/service/nginx", Explain: "Supervise the nginx service."},
		},
		ExitStatus: "0  the supervisor stopped cleanly.\n1  no directory was given or it was unwritable.",
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	rest := fs.Args()
	if len(rest) == 0 {
		return command.Failuref("a service directory is required")
	}
	dir := rest[0]

	cleanup, err := startSupervise(dir)
	if err != nil {
		return command.Failuref("%v", err)
	}
	defer cleanup()

	// A "down" file means the service should not be started, only supervised.
	if _, err := os.Stat(filepath.Join(dir, "down")); err == nil {
		<-ctx.Done()
		return nil
	}

	for {
		if ctx.Err() != nil {
			return nil
		}
		_ = runOnceFn(ctx, dir, stdio)
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(restartDelay):
		}
	}
}

// startSupervise creates the supervise directory and its ok/control files,
// returning a cleanup that removes the ok file when supervision ends.
func startSupervise(dir string) (func(), error) {
	sup := filepath.Join(dir, "supervise")
	if err := os.MkdirAll(sup, 0o755); err != nil {
		return nil, err
	}
	okPath := filepath.Join(sup, "ok")
	if err := os.WriteFile(okPath, nil, 0o644); err != nil {
		return nil, err
	}
	_ = os.WriteFile(filepath.Join(sup, "control"), nil, 0o644)
	_ = os.WriteFile(filepath.Join(sup, "pid"), []byte(strconv.Itoa(os.Getpid())+"\n"), 0o644)
	return func() { _ = os.Remove(okPath) }, nil
}
