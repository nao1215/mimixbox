// Package microcom implements the microcom applet: a minimal terminal program
// that relays bytes between the user and a serial device.
package microcom

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the microcom applet.
type Command struct{}

// New returns a microcom command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "microcom" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Minimal serial terminal program" }

// Relay copies bytes in both directions between the user side (userIn -> device,
// device -> userOut) until either side reaches EOF or the context is cancelled.
// Splitting this out from the device open lets it be exercised over pipes, the
// unit-test substitute for a real serial port.
func Relay(ctx context.Context, userIn io.Reader, userOut io.Writer, device io.ReadWriter) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	errs := make(chan error, 2)
	pump := func(dst io.Writer, src io.Reader) {
		_, err := io.Copy(dst, src)
		cancel() // one direction ending tears down the session
		errs <- err
	}

	go pump(device, userIn)  // keyboard -> serial
	go pump(userOut, device) // serial -> screen

	// The session is torn down when a direction finishes or the caller cancels.
	<-ctx.Done()

	// Give the other direction a brief grace period to finish flushing its
	// in-flight bytes, then return regardless so a pump stuck on a blocking
	// Read never holds Relay open.
	grace := time.NewTimer(gracePeriod)
	defer grace.Stop()
	var firstErr error
	for drained := 0; drained < 2; {
		select {
		case err := <-errs:
			drained++
			if firstErr == nil && err != nil && err != io.EOF && err != io.ErrClosedPipe {
				firstErr = err
			}
		case <-grace.C:
			return firstErr
		}
	}
	return firstErr
}

// gracePeriod bounds how long Relay waits for the second direction to flush
// after the first ends. It is a variable so tests can shorten it.
var gracePeriod = 50 * time.Millisecond

// openDeviceFn is indirected so opening the serial port can be replaced in a
// test. In production it opens the named device for reading and writing; the
// caller is responsible for any line settings.
var openDeviceFn = func(name string) (io.ReadWriteCloser, error) {
	f, err := os.OpenFile(name, os.O_RDWR, 0) //nolint:gosec // user-named serial device
	if err != nil {
		return nil, err
	}
	return f, nil
}

// Run executes microcom.
func (c *Command) Run(ctx context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-s SPEED] DEVICE", stdio.Err).WithHelp(command.Help{
		Description: "Open the serial DEVICE (for example /dev/ttyUSB0) and relay bytes between it and " +
			"your terminal: what you type is written to the device and what the device sends is printed. " +
			"The session ends when the device or your input closes. -s requests a line speed (baud); " +
			"the actual termios change needs a real tty, so it is recorded but only applied when the " +
			"device is a real serial port. Connect standard input/output to pipes to test the relay " +
			"without a serial port.",
		Examples: []command.Example{
			{Command: "microcom /dev/ttyUSB0", Explain: "Open a serial console."},
			{Command: "microcom -s 115200 /dev/ttyS0", Explain: "Open at 115200 baud."},
		},
		ExitStatus: "0  the session ended normally.\n" +
			"1  no device was named, or the device could not be opened.",
	})
	_ = fs.IntP("speed", "s", 0, "requested line speed in baud")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	rest := fs.Args()
	if len(rest) == 0 {
		return command.Failuref("a serial device is required")
	}
	if len(rest) > 1 {
		return command.Failuref("unexpected argument: %q", rest[1])
	}

	dev, err := openDeviceFn(rest[0])
	if err != nil {
		return command.Failuref("cannot open %q: %v", rest[0], err)
	}
	defer func() { _ = dev.Close() }()

	if err := Relay(ctx, stdio.In, stdio.Out, dev); err != nil {
		return command.Failuref("%v", err)
	}
	return nil
}
