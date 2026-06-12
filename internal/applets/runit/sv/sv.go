// Package sv implements the sv applet: control or query a runit service by
// writing to its supervise/control file and reading its supervise state.
package sv

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the sv applet.
type Command struct{}

// New returns a sv command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "sv" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Control or query a runit service" }

// controlChars maps each sv command to the character written to the supervise
// control file (the runit/daemontools control protocol).
var controlChars = map[string]string{
	"up": "u", "down": "d", "once": "o", "pause": "p", "cont": "c",
	"hup": "h", "alarm": "a", "interrupt": "i", "term": "t", "kill": "k",
	"quit": "q", "exit": "x", "restart": "t",
}

// Run executes sv.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "COMMAND DIR...", stdio.Err).WithHelp(command.Help{
		Description: "Control or query the runit service(s) in each DIR. COMMAND is status, or a control " +
			"command (up, down, once, pause, cont, hup, alarm, interrupt, term, kill, quit, exit, " +
			"restart) that is written to DIR/supervise/control. 'status' reports whether each service " +
			"is supervised and, if available, its pid.",
		Examples: []command.Example{
			{Command: "sv status /etc/service/nginx", Explain: "Show nginx's status."},
			{Command: "sv restart /etc/service/nginx", Explain: "Restart nginx."},
		},
		ExitStatus: "0  the command was delivered (or status was reported).\n1  a usage or I/O error.",
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	rest := fs.Args()
	if len(rest) < 2 {
		return command.Failuref("a command and at least one service directory are required")
	}
	cmd, dirs := rest[0], rest[1:]

	if cmd == "status" {
		return reportStatus(stdio, dirs)
	}

	ch, ok := controlChars[cmd]
	if !ok {
		return command.Failuref("unknown command: %q", cmd)
	}
	failed := false
	for _, dir := range dirs {
		if err := sendControl(dir, ch); err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "sv: %s: %v\n", dir, err)
			failed = true
		}
	}
	if failed {
		return command.SilentFailure()
	}
	return nil
}

// sendControl writes the control character to the service's control file.
func sendControl(dir, ch string) error {
	control := filepath.Join(dir, "supervise", "control")
	f, err := os.OpenFile(control, os.O_WRONLY, 0) //nolint:gosec // the supervise control fifo
	if err != nil {
		return fmt.Errorf("not supervised (no %s): %v", control, err)
	}
	defer func() { _ = f.Close() }()
	_, err = f.WriteString(ch)
	return err
}

// reportStatus prints the supervision state of each service directory.
func reportStatus(stdio command.IO, dirs []string) error {
	failed := false
	for _, dir := range dirs {
		if _, err := os.Stat(filepath.Join(dir, "supervise", "ok")); err != nil {
			_, _ = fmt.Fprintf(stdio.Out, "%s: not supervised\n", dir)
			failed = true
			continue
		}
		state, detail := "run", ""
		if pid, err := os.ReadFile(filepath.Join(dir, "supervise", "pid")); err == nil { //nolint:gosec // supervise dir
			if p := strings.TrimSpace(string(pid)); p != "" {
				detail = " (pid " + p + ")"
			}
		}
		if _, err := os.Stat(filepath.Join(dir, "down")); err == nil {
			state = "down"
		}
		_, _ = fmt.Fprintf(stdio.Out, "%s: %s%s\n", dir, state, detail)
	}
	if failed {
		return command.SilentFailure()
	}
	return nil
}
