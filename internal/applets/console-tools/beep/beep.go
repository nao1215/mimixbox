// Package beep implements the beep applet: sound the console speaker.
package beep

import (
	"context"
	"os"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
	"golang.org/x/sys/unix"
)

// Command is the beep applet.
type Command struct{}

// New returns a beep command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "beep" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Sound the console speaker" }

// KIOCSOUND drives the console speaker; the PIT runs at this clock.
const (
	kiocSound     = 0x4B2F
	pitClock      = 1193180
	defaultFreq   = 4000
	defaultLength = 30
)

// beepFn is indirected so the speaker is not actually driven in tests.
var beepFn = func(freq, lengthMs int) error {
	f, err := os.Open("/dev/console") //nolint:gosec // the system console
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	tick := uintptr(0)
	if freq > 0 {
		tick = uintptr(pitClock / freq)
	}
	if _, _, errno := unix.Syscall(unix.SYS_IOCTL, f.Fd(), kiocSound, tick); errno != 0 {
		return errno
	}
	time.Sleep(time.Duration(lengthMs) * time.Millisecond)
	_, _, _ = unix.Syscall(unix.SYS_IOCTL, f.Fd(), kiocSound, 0) // silence
	return nil
}

// Run executes beep.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-f FREQ] [-l MS] [-r N]", stdio.Err).WithHelp(command.Help{
		Description: "Sound the console speaker. -f sets the frequency in hertz (default 4000), -l the " +
			"length in milliseconds (default 30), and -r the number of repetitions. Requires access to " +
			"the console and privilege.",
		Examples: []command.Example{
			{Command: "beep -f 1000 -l 200", Explain: "A 1 kHz tone for 200 ms."},
			{Command: "beep -r 3", Explain: "Three default beeps."},
		},
		ExitStatus: "0  the speaker was sounded.\n1  invalid options or the console was inaccessible.",
	})
	freq := fs.IntP("frequency", "f", defaultFreq, "frequency in hertz")
	length := fs.IntP("length", "l", defaultLength, "length in milliseconds")
	repeats := fs.IntP("repeats", "r", 1, "number of beeps")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	if *freq <= 0 {
		return command.Failuref("frequency must be positive")
	}
	if *repeats < 1 {
		return command.Failuref("the repeat count must be at least 1")
	}

	for i := 0; i < *repeats; i++ {
		if err := beepFn(*freq, *length); err != nil {
			return command.Failuref("cannot sound the speaker: %v", err)
		}
	}
	return nil
}
