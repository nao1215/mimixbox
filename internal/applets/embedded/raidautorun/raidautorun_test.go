package raidautorun

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

type fakeRunner struct {
	seen string
	err  error
}

func (f *fakeRunner) Run(device string) error {
	f.seen = device
	return f.err
}

func withRunner(t *testing.T, r AutoRunner) {
	t.Helper()
	prev := autorun
	autorun = r
	t.Cleanup(func() { autorun = prev })
}

func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	var out, errBuf bytes.Buffer
	stdio := command.IO{In: strings.NewReader(""), Out: &out, Err: &errBuf}
	err := New().Run(context.Background(), stdio, args)
	return errBuf.String(), err
}

func TestRaidautorunTriggers(t *testing.T) {
	fake := &fakeRunner{}
	withRunner(t, fake)
	if _, err := run(t, "/dev/md0"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fake.seen != "/dev/md0" {
		t.Errorf("unexpected device: %q", fake.seen)
	}
}

func TestRaidautorunUsage(t *testing.T) {
	withRunner(t, &fakeRunner{})
	if _, err := run(t); err == nil {
		t.Fatal("expected usage error with no device")
	}
}

func TestRaidautorunCapabilityError(t *testing.T) {
	withRunner(t, &fakeRunner{err: errors.New("operation not permitted")})
	if _, err := run(t, "/dev/md0"); err == nil {
		t.Fatal("expected error when ioctl fails")
	}
}
