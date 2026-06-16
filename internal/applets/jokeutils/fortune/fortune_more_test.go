package fortune

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

// errWriter fails every write, to exercise the output-error path.
type errWriter struct{}

func (errWriter) Write([]byte) (int, error) { return 0, errors.New("write failed") }

// TestRunNoFortunesAvailable covers the empty-pool branch by emptying the
// built-in collection for the duration of the test.
func TestRunNoFortunesAvailable(t *testing.T) {
	orig := fortunes
	fortunes = nil
	t.Cleanup(func() { fortunes = orig })

	io := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, nil); err == nil {
		t.Fatal("expected an error when no fortunes are available")
	}
}

// TestRunShortFiltersToEmpty covers candidates(short=true) yielding an empty
// pool when no short fortunes exist.
func TestRunShortFiltersToEmpty(t *testing.T) {
	orig := fortunes
	fortunes = []string{strings.Repeat("x", shortLimit+50)}
	t.Cleanup(func() { fortunes = orig })

	io := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, []string{"-s"}); err == nil {
		t.Fatal("expected an error when no short fortunes are available")
	}
}

// TestRunWriteError covers the Fprintln failure branch.
func TestRunWriteError(t *testing.T) {
	io := command.IO{In: strings.NewReader(""), Out: errWriter{}, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, nil); err == nil {
		t.Fatal("expected an error when the adage cannot be written")
	}
}
