// Package hwclock implements the hwclock applet: read the hardware (RTC) clock.
package hwclock

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
	"golang.org/x/sys/unix"
)

// Command is the hwclock applet.
type Command struct{}

// New returns a hwclock command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "hwclock" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Read the hardware (RTC) clock" }

// readRTC is indirected so the formatting can be tested without a real RTC. It
// returns the RTC time, which the kernel keeps in UTC.
var readRTC = func() (time.Time, error) {
	f, err := os.OpenFile("/dev/rtc0", os.O_RDONLY, 0)
	if err != nil {
		f, err = os.OpenFile("/dev/rtc", os.O_RDONLY, 0)
		if err != nil {
			return time.Time{}, err
		}
	}
	defer func() { _ = f.Close() }()

	rt, err := unix.IoctlGetRTCTime(int(f.Fd()))
	if err != nil {
		return time.Time{}, err
	}
	return time.Date(int(rt.Year)+1900, time.Month(rt.Mon+1), int(rt.Mday),
		int(rt.Hour), int(rt.Min), int(rt.Sec), 0, time.UTC), nil
}

// Run executes hwclock.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-r] [-u]", stdio.Err).WithHelp(command.Help{
		Description: "Read the hardware real-time clock and print it. With -u, print it in UTC; " +
			"otherwise in local time. Setting the clock is not implemented in this slice.",
		Examples: []command.Example{
			{Command: "hwclock", Explain: "Print the RTC time in local time."},
			{Command: "hwclock -u", Explain: "Print it in UTC."},
		},
		ExitStatus: "0  success.\n1  the RTC could not be read.",
	})
	_ = fs.BoolP("show", "r", false, "read and print the RTC (the default)")
	utc := fs.BoolP("utc", "u", false, "print the time in UTC")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	t, err := readRTC()
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "hwclock: cannot read the RTC: %v\n", err)
		return command.SilentFailure()
	}

	if !*utc {
		t = t.Local()
	}
	_, _ = fmt.Fprintln(stdio.Out, t.Format("2006-01-02 15:04:05"))
	return nil
}
