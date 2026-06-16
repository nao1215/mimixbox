// Package acpid implements the acpid applet: a foreground ACPI event daemon that
// reads kernel ACPI events and dispatches each to a handler.
package acpid

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the acpid applet.
type Command struct{}

// New returns an acpid command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "acpid" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Dispatch ACPI events (foreground)" }

// Injected so the event source and the handler are testable.
var (
	openSourceFn = func() (io.ReadCloser, error) {
		return os.Open("/proc/acpi/event") //nolint:gosec // well-known ACPI event source
	}
	// handlerFn handles one event line. The default runs /etc/acpi/handler.sh with
	// the event fields if it exists, and always reports the event.
	handlerFn = func(stdio command.IO, event string) {
		_, _ = fmt.Fprintf(stdio.Out, "acpid: %s\n", event)
		const handler = "/etc/acpi/handler.sh"
		if _, err := os.Stat(handler); err == nil {
			cmd := exec.Command(handler, strings.Fields(event)...) //nolint:gosec // configured ACPI handler
			cmd.Stdout, cmd.Stderr = stdio.Out, stdio.Err
			_ = cmd.Run()
		}
	}
)

// Run executes acpid.
func (c *Command) Run(ctx context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "-f", stdio.Err).WithHelp(command.Help{
		Description: "Read kernel ACPI events (from /proc/acpi/event) and dispatch each to a handler, " +
			"running /etc/acpi/handler.sh with the event fields when it exists. Only the foreground " +
			"mode (-f) is supported: acpid stays in the foreground until interrupted.",
		Examples: []command.Example{
			{Command: "acpid -f", Explain: "Run the ACPI event daemon in the foreground."},
		},
		ExitStatus: "0  the daemon stopped cleanly.\n1  -f was not given or the event source was unavailable.",
		Notes: []string{
			"Only foreground mode (-f) is supported; MimixBox does not daemonize or write a PID file.",
			"Reading /proc/acpi/event requires a Linux host that still exposes the legacy ACPI event interface.",
		},
	})
	foreground := fs.BoolP("foreground", "f", false, "run in the foreground (the only supported mode)")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	if !*foreground {
		_, _ = fmt.Fprintln(stdio.Err, "acpid: only the foreground mode (-f) is supported by this build")
		return command.SilentFailure()
	}

	src, err := openSourceFn()
	if err != nil {
		return command.Failuref("cannot open the ACPI event source: %v", err)
	}
	defer func() { _ = src.Close() }()

	// Closing the source on cancellation unblocks the read.
	go func() {
		<-ctx.Done()
		_ = src.Close()
	}()

	sc := bufio.NewScanner(src)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		handlerFn(stdio, line)
	}
	return nil
}
