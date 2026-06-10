// Package rtcwake implements the rtcwake applet: arm the real-time clock to wake
// the system at a future time.
package rtcwake

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the rtcwake applet.
type Command struct{}

// New returns a rtcwake command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "rtcwake" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Arm the RTC to wake the system" }

// Injected so the alarm can be set against a fixture and a fixed clock.
var (
	sysRTCDir = "/sys/class/rtc"
	now       = time.Now
)

// suspendModes are the rtcwake modes that put the system to sleep; this build
// only arms the alarm, so they are rejected.
var suspendModes = map[string]bool{
	"mem": true, "disk": true, "standby": true, "freeze": true, "off": true,
}

// Run executes rtcwake.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-d DEVICE] [-m no] {-s SECONDS|-t EPOCH}", stdio.Err).WithHelp(command.Help{
		Description: "Arm the real-time clock alarm to fire at a future time, by writing the wake time " +
			"to /sys/class/rtc/<device>/wakealarm. Give the time as -s SECONDS from now or -t as an " +
			"absolute Unix epoch. Only -m no (arm the alarm without suspending) is supported; the " +
			"suspend modes are not, since this build does not put the system to sleep.",
		Examples: []command.Example{
			{Command: "rtcwake -m no -s 300", Explain: "Arm the alarm for 5 minutes from now."},
		},
		ExitStatus: "0  the alarm was armed.\n1  invalid options or the alarm could not be written.",
	})
	device := fs.StringP("device", "d", "rtc0", "RTC device under /sys/class/rtc")
	mode := fs.StringP("mode", "m", "no", "standby mode (only 'no' is supported)")
	seconds := fs.IntP("seconds", "s", 0, "seconds from now to wake")
	epoch := fs.Int64P("time", "t", 0, "absolute Unix epoch to wake at")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	if suspendModes[*mode] {
		return command.Failuref("suspend mode %q is not supported; use -m no to arm the alarm only", *mode)
	}
	if *mode != "no" && *mode != "on" {
		return command.Failuref("unknown mode: %q", *mode)
	}

	var wake int64
	switch {
	case fs.Changed("time"):
		wake = *epoch
	case fs.Changed("seconds"):
		wake = now().Unix() + int64(*seconds)
	default:
		return command.Failuref("a wake time is required (-s SECONDS or -t EPOCH)")
	}

	alarm := filepath.Join(sysRTCDir, *device, "wakealarm")
	// Clear any existing alarm before setting the new one, as the kernel rejects
	// a write while an alarm is already pending.
	if err := os.WriteFile(alarm, []byte("0\n"), 0o644); err != nil {
		return command.Failuref("cannot access %s: %v", alarm, err)
	}
	if err := os.WriteFile(alarm, []byte(strconv.FormatInt(wake, 10)+"\n"), 0o644); err != nil {
		return command.Failuref("cannot set alarm on %s: %v", alarm, err)
	}

	_, _ = fmt.Fprintf(stdio.Out, "rtcwake: alarm armed for %s (epoch %d)\n",
		time.Unix(wake, 0).Format("2006-01-02 15:04:05"), wake)
	return nil
}
