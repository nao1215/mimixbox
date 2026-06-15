package microcom

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
)

// fakeDevice is a bidirectional in-memory ReadWriteCloser: reads drain its
// preset "from device" data, writes accumulate into a buffer.
type fakeDevice struct {
	in  *strings.Reader
	out bytes.Buffer
}

func (d *fakeDevice) Read(p []byte) (int, error)  { return d.in.Read(p) }
func (d *fakeDevice) Write(p []byte) (int, error) { return d.out.Write(p) }
func (d *fakeDevice) Close() error                { return nil }

func TestRelayBothDirections(t *testing.T) {
	t.Parallel()
	dev := &fakeDevice{in: strings.NewReader("hello from device")}
	userIn := strings.NewReader("typed by user")
	var screen bytes.Buffer

	if err := Relay(context.Background(), userIn, &screen, dev); err != nil {
		t.Fatalf("Relay: %v", err)
	}
	if !strings.Contains(screen.String(), "hello from device") {
		t.Errorf("screen = %q", screen.String())
	}
	if !strings.Contains(dev.out.String(), "typed by user") {
		t.Errorf("device received = %q", dev.out.String())
	}
}

func TestRelayContextCancel(t *testing.T) {
	t.Parallel()
	// Device that never EOFs; cancellation must end the relay.
	dev := &endlessDevice{}
	pr, _ := io.Pipe() // user input that blocks forever
	var screen bytes.Buffer

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- Relay(ctx, pr, &screen, dev) }()
	time.Sleep(10 * time.Millisecond)
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Relay did not return after cancel")
	}
}

type endlessDevice struct{}

func (endlessDevice) Read(p []byte) (int, error) {
	time.Sleep(time.Millisecond)
	for i := range p {
		p[i] = '.'
	}
	return len(p), nil
}
func (endlessDevice) Write(p []byte) (int, error) { return len(p), nil }
func (endlessDevice) Close() error                { return nil }

func TestRunNoDevice(t *testing.T) {
	t.Parallel()
	var out, errBuf bytes.Buffer
	io := command.IO{In: strings.NewReader(""), Out: &out, Err: &errBuf}
	if err := New().Run(context.Background(), io, nil); err == nil {
		t.Error("expected error when no device given")
	}
}

func TestRunOpenError(t *testing.T) {
	t.Parallel()
	var out, errBuf bytes.Buffer
	io := command.IO{In: strings.NewReader(""), Out: &out, Err: &errBuf}
	if err := New().Run(context.Background(), io, []string{"/no/such/device"}); err == nil {
		t.Error("expected error opening missing device")
	}
}

func TestRunInjectedDevice(t *testing.T) {
	orig := openDeviceFn
	openDeviceFn = func(_ string) (io.ReadWriteCloser, error) {
		return &fakeDevice{in: strings.NewReader("PROMPT> ")}, nil
	}
	defer func() { openDeviceFn = orig }()

	var out, errBuf bytes.Buffer
	io := command.IO{In: strings.NewReader("AT\r"), Out: &out, Err: &errBuf}
	if err := New().Run(context.Background(), io, []string{"/dev/ttyUSB0"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "PROMPT>") {
		t.Errorf("out = %q", out.String())
	}
}
