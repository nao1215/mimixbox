package acpid

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
)

func withSource(t *testing.T, r io.ReadCloser, openErr error) {
	t.Helper()
	oo := openSourceFn
	openSourceFn = func() (io.ReadCloser, error) {
		if openErr != nil {
			return nil, openErr
		}
		return r, nil
	}
	t.Cleanup(func() { openSourceFn = oo })
}

func TestDispatchesEvents(t *testing.T) {
	withSource(t, io.NopCloser(strings.NewReader(
		"button/power PBTN 00000080 00000000\n\nac_adapter ACAD 00000080 00000001\n")), nil)
	out := &bytes.Buffer{}
	io2 := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io2, []string{"-f"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "button/power PBTN") || !strings.Contains(out.String(), "ac_adapter ACAD") {
		t.Errorf("events not dispatched:\n%s", out)
	}
}

func TestForegroundRequired(t *testing.T) {
	io2 := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io2, nil); err == nil {
		t.Errorf("acpid without -f should fail")
	}
}

func TestOpenFailure(t *testing.T) {
	withSource(t, nil, errors.New("no such file"))
	io2 := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io2, []string{"-f"}); err == nil {
		t.Errorf("an unavailable source should fail")
	}
}

type blockingReader struct{ closed chan struct{} }

func (b *blockingReader) Read([]byte) (int, error) { <-b.closed; return 0, io.EOF }
func (b *blockingReader) Close() error {
	select {
	case <-b.closed:
	default:
		close(b.closed)
	}
	return nil
}

func TestStopsOnCancel(t *testing.T) {
	withSource(t, &blockingReader{closed: make(chan struct{})}, nil)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		io2 := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
		done <- New().Run(ctx, io2, []string{"-f"})
	}()
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Run returned %v after cancel", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("acpid did not stop after cancellation")
	}
}

// TestHelpNotes asserts the --help output documents a Notes section.
func TestHelpNotes(t *testing.T) {
	t.Parallel()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, []string{"--help"}); err != nil {
		t.Fatalf("--help err = %v", err)
	}
	if !strings.Contains(out.String(), "Notes:") {
		t.Errorf("--help missing Notes section: %q", out.String())
	}
}
