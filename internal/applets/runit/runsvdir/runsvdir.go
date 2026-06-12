// Package runsvdir implements the runsvdir applet: start and supervise a runsv
// for every service directory under a directory.
package runsvdir

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/nao1215/mimixbox/internal/applets/runit/runsv"
	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the runsvdir applet.
type Command struct{}

// New returns a runsvdir command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "runsvdir" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Supervise a directory of services" }

// Injected so the per-service supervisor and the rescan interval are testable.
var (
	rescanInterval = 5 * time.Second
	startRunsvFn   = func(ctx context.Context, dir string, stdio command.IO) {
		_ = runsv.New().Run(ctx, stdio, []string{dir})
	}
)

// Run executes runsvdir.
func (c *Command) Run(ctx context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "DIR", stdio.Err).WithHelp(command.Help{
		Description: "Supervise every service in DIR: start a runsv for each subdirectory and keep it " +
			"running, rescanning periodically so newly added services are picked up, until " +
			"interrupted. Subdirectories whose names start with a dot are ignored.",
		Examples: []command.Example{
			{Command: "runsvdir /etc/service", Explain: "Supervise all services under /etc/service."},
		},
		ExitStatus: "0  the supervisor stopped cleanly.\n1  no directory was given or it was unreadable.",
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	rest := fs.Args()
	if len(rest) == 0 {
		return command.Failuref("a services directory is required")
	}
	dir := rest[0]
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		return command.Failuref("%s: not a directory", dir)
	}

	started := map[string]bool{}
	for {
		if err := scanAndStart(ctx, dir, started, stdio); err != nil {
			return command.Failuref("%s: %v", dir, err)
		}
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(rescanInterval):
		}
	}
}

// scanAndStart starts a runsv for each not-yet-started service subdirectory.
func scanAndStart(ctx context.Context, dir string, started map[string]bool, stdio command.IO) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") || started[e.Name()] {
			continue
		}
		started[e.Name()] = true
		service := filepath.Join(dir, e.Name())
		go startRunsvFn(ctx, service, stdio)
	}
	return nil
}
