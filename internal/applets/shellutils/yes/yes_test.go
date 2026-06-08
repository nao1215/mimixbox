package yes_test

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/shellutils/yes"
	"github.com/nao1215/mimixbox/internal/command"
)

// errorWriter fails every write, standing in for a closed pipe so yes stops.
type errorWriter struct{}

func (errorWriter) Write([]byte) (int, error) { return 0, errors.New("broken pipe") }

// limitedWriter records output but reports a failure once it has accepted limit
// bytes, so yes terminates after producing a bounded, inspectable prefix.
type limitedWriter struct {
	buf   bytes.Buffer
	limit int
}

func (w *limitedWriter) Write(p []byte) (int, error) {
	if w.buf.Len() >= w.limit {
		return 0, errors.New("limit reached")
	}
	return w.buf.Write(p)
}

func runUntilError(t *testing.T, args ...string) (string, error) {
	t.Helper()
	w := &limitedWriter{limit: 256}
	io := command.IO{In: strings.NewReader(""), Out: w, Err: &bytes.Buffer{}}
	err := yes.New().Run(context.Background(), io, args)
	return w.buf.String(), err
}

func TestStopsOnWriteError(t *testing.T) {
	t.Parallel()
	io := command.IO{In: strings.NewReader(""), Out: errorWriter{}, Err: &bytes.Buffer{}}
	if err := yes.New().Run(context.Background(), io, nil); err != nil {
		t.Fatalf("Run returned %v, want nil on write error", err)
	}
}

func TestContextCancel(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // already cancelled: Run should return promptly
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := yes.New().Run(ctx, io, nil); err != nil {
		t.Fatalf("Run returned %v, want nil", err)
	}
}

func TestDefaultString(t *testing.T) {
	t.Parallel()
	got, err := runUntilError(t)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.HasPrefix(got, "y\ny\n") {
		t.Errorf("output prefix = %q", got[:min(8, len(got))])
	}
}

func TestCustomString(t *testing.T) {
	t.Parallel()
	got, err := runUntilError(t, "hello", "world")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.HasPrefix(got, "hello world\nhello world\n") {
		t.Errorf("output prefix = %q", got[:min(24, len(got))])
	}
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := yes.New()
	if c.Name() != "yes" {
		t.Errorf("Name() = %q", c.Name())
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
