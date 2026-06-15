// Package watchdog implements the watchdog applet: periodically pet a hardware
// or software watchdog timer so it does not reset the system.
package watchdog

import (
	"context"
	"fmt"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the watchdog applet.
type Command struct{}

// New returns a watchdog command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "watchdog" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Pet a watchdog timer to prevent a reset" }

// pinger opens and keeps a watchdog device alive. It is injected so the option
// parsing and the petting loop can be tested without a real /dev/watchdog.
var pinger Pinger = osPinger{}

// Pinger abstracts the privileged watchdog device operations.
type Pinger interface {
	// Open opens the watchdog at device and programs its hardware timeout in
	// seconds. It returns a keepalive function (called once per interval) and
	// a close function (called on exit; a graceful close stops the watchdog).
	Open(device string, timeoutSec int) (keepalive func() error, closeFn func() error, err error)
}

// Run executes watchdog.
func (c *Command) Run(ctx context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-t interval] [-T timeout] [-n count] DEVICE", stdio.Err).WithHelp(command.Help{
		Description: "Open the watchdog DEVICE (e.g. /dev/watchdog) and keep petting it every -t seconds so " +
			"the timer never expires. -T sets the hardware timeout in seconds (default 60). By default the " +
			"command runs until interrupted; -n COUNT pets the watchdog COUNT times and exits, which is " +
			"useful for testing. WARNING: while running, the watchdog will reset the whole system if the " +
			"process stops without a graceful close. Opening the device requires privilege.",
		Examples: []command.Example{
			{Command: "watchdog -t 5 -T 30 /dev/watchdog", Explain: "Pet every 5s with a 30s hardware timeout."},
			{Command: "watchdog -n 3 -t 1 /dev/watchdog", Explain: "Pet 3 times then exit cleanly."},
		},
		ExitStatus: "0  the watchdog was petted (and closed for -n).\n1  bad arguments or the device could not be opened.",
	})
	interval := fs.IntP("interval", "t", 30, "seconds between keepalive pets")
	timeout := fs.IntP("timeout", "T", 60, "hardware timeout in seconds")
	count := fs.IntP("count", "n", 0, "pet this many times then exit (0 = run forever)")
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	rest := fs.Args()
	if len(rest) != 1 {
		return command.Failuref("usage: watchdog [-t interval] [-T timeout] [-n count] DEVICE")
	}
	if *interval <= 0 || *timeout <= 0 {
		return command.Failuref("interval and timeout must be positive")
	}
	if *count < 0 {
		return command.Failuref("count must not be negative")
	}

	keepalive, closeFn, err := pinger.Open(rest[0], *timeout)
	if err != nil {
		return command.Failuref("%s: %v", rest[0], err)
	}
	defer func() { _ = closeFn() }()

	return c.loop(ctx, stdio, keepalive, *interval, *count)
}

// loop pets the watchdog, either count times (count > 0) or until ctx is done.
func (c *Command) loop(ctx context.Context, stdio command.IO, keepalive func() error, interval, count int) error {
	tick := time.Duration(interval) * time.Second
	for i := 0; count == 0 || i < count; i++ {
		if err := keepalive(); err != nil {
			return command.Failuref("keepalive: %v", err)
		}
		// The final pet of a bounded run does not need to wait.
		if count != 0 && i == count-1 {
			break
		}
		select {
		case <-ctx.Done():
			_, _ = fmt.Fprintln(stdio.Err, "watchdog: interrupted, closing device")
			return nil
		case <-time.After(tick):
		}
	}
	return nil
}
