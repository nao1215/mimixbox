package partprobe

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

type fakeReReader struct {
	seen []string
	err  error
}

func (f *fakeReReader) ReRead(device string) error {
	f.seen = append(f.seen, device)
	return f.err
}

func withReReader(t *testing.T, r ReReader) {
	t.Helper()
	prev := reread
	reread = r
	t.Cleanup(func() { reread = prev })
}

func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	var out, errBuf bytes.Buffer
	stdio := command.IO{In: strings.NewReader(""), Out: &out, Err: &errBuf}
	err := New().Run(context.Background(), stdio, args)
	return errBuf.String(), err
}

func TestPartprobeRereads(t *testing.T) {
	fake := &fakeReReader{}
	withReReader(t, fake)
	if _, err := run(t, "/dev/sda", "/dev/sdb"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fake.seen) != 2 {
		t.Errorf("expected 2 devices, got %v", fake.seen)
	}
}

func TestPartprobeNoArgs(t *testing.T) {
	withReReader(t, &fakeReReader{})
	if _, err := run(t); err == nil {
		t.Fatal("expected error with no device operands")
	}
}

func TestPartprobeCapabilityError(t *testing.T) {
	withReReader(t, &fakeReReader{err: errors.New("operation not permitted")})
	errOut, err := run(t, "/dev/sda")
	if err == nil {
		t.Fatal("expected error when re-read fails")
	}
	if !strings.Contains(errOut, "partprobe:") {
		t.Errorf("missing prefix: %q", errOut)
	}
}
