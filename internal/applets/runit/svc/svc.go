// Package svc implements the svc applet: send control commands to a daemontools
// service by writing to its supervise/control file.
package svc

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the svc applet.
type Command struct{}

// New returns a svc command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "svc" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Send control commands to a service" }

// controlFlag binds a command-line flag to a supervise control character.
type controlFlag struct {
	short string
	ch    byte
	help  string
}

// flags are applied in this order — signal commands before the up/down/exit
// state commands — so a combination like -t -u sends "tu" (terminate, then keep
// the service up), the conventional restart.
var flags = []controlFlag{
	{"t", 't', "terminate (SIGTERM)"},
	{"k", 'k', "kill (SIGKILL)"},
	{"p", 'p', "pause (SIGSTOP)"},
	{"c", 'c', "continue (SIGCONT)"},
	{"h", 'h', "hangup (SIGHUP)"},
	{"a", 'a', "alarm (SIGALRM)"},
	{"i", 'i', "interrupt (SIGINT)"},
	{"o", 'o', "once: start, but do not restart"},
	{"u", 'u', "up: start and restart on exit"},
	{"d", 'd', "down: stop and do not restart"},
	{"x", 'x', "exit: ask the supervisor to exit"},
}

// Run executes svc.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "-{udopchaitkx} DIR...", stdio.Err).WithHelp(command.Help{
		Description: "Send control commands to the daemontools service(s) in each DIR by writing the " +
			"corresponding characters to DIR/supervise/control: -u up, -d down, -o once, -p pause, " +
			"-c continue, -h hup, -a alarm, -i interrupt, -t term, -k kill, -x exit. Multiple commands " +
			"may be combined.",
		Examples: []command.Example{
			{Command: "svc -d /service/nginx", Explain: "Stop nginx and keep it down."},
			{Command: "svc -t -u /service/nginx", Explain: "Restart nginx (term, then up)."},
		},
		ExitStatus: "0  the commands were delivered.\n1  no command/dir was given, or an I/O error.",
	})
	set := make(map[string]*bool, len(flags))
	for _, f := range flags {
		set[f.short] = fs.BoolP("cmd-"+f.short, f.short, false, f.help)
	}

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	var control []byte
	for _, f := range flags {
		if *set[f.short] {
			control = append(control, f.ch)
		}
	}
	if len(control) == 0 {
		return command.Failuref("at least one control command is required")
	}

	dirs := fs.Args()
	if len(dirs) == 0 {
		return command.Failuref("at least one service directory is required")
	}

	failed := false
	for _, dir := range dirs {
		if err := sendControl(dir, control); err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "svc: %s: %v\n", dir, err)
			failed = true
		}
	}
	if failed {
		return command.SilentFailure()
	}
	return nil
}

// sendControl writes the control bytes to the service's control file.
func sendControl(dir string, control []byte) error {
	path := filepath.Join(dir, "supervise", "control")
	f, err := os.OpenFile(path, os.O_WRONLY, 0) //nolint:gosec // the supervise control fifo
	if err != nil {
		return fmt.Errorf("not supervised (no %s): %v", path, err)
	}
	defer func() { _ = f.Close() }()
	_, err = f.Write(control)
	return err
}
