package watchdog

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

type fakePinger struct {
	device   string
	timeout  int
	pets     int
	closed   bool
	openErr  error
	petErr   error
}

func (f *fakePinger) Open(device string, timeoutSec int) (func() error, func() error, error) {
	if f.openErr != nil {
		return nil, nil, f.openErr
	}
	f.device = device
	f.timeout = timeoutSec
	keepalive := func() error {
		f.pets++
		return f.petErr
	}
	closeFn := func() error {
		f.closed = true
		return nil
	}
	return keepalive, closeFn, nil
}

func withPinger(t *testing.T, p Pinger) {
	t.Helper()
	prev := pinger
	pinger = p
	t.Cleanup(func() { pinger = prev })
}

func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	var out, errBuf bytes.Buffer
	stdio := command.IO{In: strings.NewReader(""), Out: &out, Err: &errBuf}
	err := New().Run(context.Background(), stdio, args)
	return errBuf.String(), err
}

func TestWatchdogBoundedPets(t *testing.T) {
	fake := &fakePinger{}
	withPinger(t, fake)
	// -n 1 pets exactly once with no wait, then exits and closes.
	if _, err := run(t, "-n", "1", "-t", "1", "-T", "20", "/dev/watchdog"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fake.pets != 1 {
		t.Errorf("expected 1 pet, got %d", fake.pets)
	}
	if !fake.closed {
		t.Error("device should be closed on exit")
	}
	if fake.device != "/dev/watchdog" || fake.timeout != 20 {
		t.Errorf("open args wrong: device=%q timeout=%d", fake.device, fake.timeout)
	}
}

func TestWatchdogContextCancel(t *testing.T) {
	fake := &fakePinger{}
	withPinger(t, fake)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel before the first wait so the run-forever loop returns

	var out, errBuf bytes.Buffer
	stdio := command.IO{In: strings.NewReader(""), Out: &out, Err: &errBuf}
	if err := New().Run(ctx, stdio, []string{"-t", "1", "/dev/watchdog"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// At least one pet happens before the cancellation is observed.
	if fake.pets < 1 || !fake.closed {
		t.Errorf("expected pet+close on cancel: pets=%d closed=%v", fake.pets, fake.closed)
	}
}

func TestWatchdogOpenError(t *testing.T) {
	withPinger(t, &fakePinger{openErr: errors.New("permission denied")})
	if _, err := run(t, "-n", "1", "/dev/watchdog"); err == nil {
		t.Fatal("expected error when open fails")
	}
}

func TestWatchdogPetError(t *testing.T) {
	withPinger(t, &fakePinger{petErr: errors.New("device gone")})
	if _, err := run(t, "-n", "1", "/dev/watchdog"); err == nil {
		t.Fatal("expected error when keepalive fails")
	}
}

func TestWatchdogUsageErrors(t *testing.T) {
	withPinger(t, &fakePinger{})
	if _, err := run(t); err == nil {
		t.Error("expected usage error with no device")
	}
	if _, err := run(t, "-t", "0", "/dev/watchdog"); err == nil {
		t.Error("expected error for non-positive interval")
	}
}
