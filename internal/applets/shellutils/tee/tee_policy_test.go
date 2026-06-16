package tee

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

// syscallPipeErr returns an EPIPE error for the broken-pipe tolerance tests.
func syscallPipeErr() error { return syscall.EPIPE }

// failingWriter fails every Write with a non-pipe error, standing in for a
// destination that cannot be written.
type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errors.New("disk full") }

// TestParseOutputErrorMode covers every accepted MODE plus the invalid case.
func TestParseOutputErrorMode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in      string
		want    outputErrorMode
		wantErr bool
	}{
		{"", outputErrorWarn, false},
		{"warn", outputErrorWarn, false},
		{"warn-nopipe", outputErrorWarnNopipe, false},
		{"exit", outputErrorExit, false},
		{"exit-nopipe", outputErrorExitNopipe, false},
		{"bogus", 0, true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.in, func(t *testing.T) {
			t.Parallel()
			got, err := parseOutputErrorMode(tt.in)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseOutputErrorMode(%q) err = %v, wantErr %v", tt.in, err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("parseOutputErrorMode(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

// teeTo runs tee directly against a custom destination list so a failing writer
// can be injected. The writable file content and stderr are returned alongside
// the error.
func teeTo(t *testing.T, stdin string, dsts []*teeWriter, mode outputErrorMode) (string, error) {
	t.Helper()
	errBuf := &strings.Builder{}
	stdio := command.IO{In: strings.NewReader(stdin), Out: io.Discard, Err: errBuf}
	failed, stopped := copyStream(stdio, dsts, mode)
	if failed || stopped {
		return errBuf.String(), command.SilentFailure()
	}
	return errBuf.String(), nil
}

// TestWarnModeContinuesAndFails verifies that, with a failing writer present,
// warn mode still writes the writable file and reports a nonzero exit.
func TestWarnModeContinuesAndFails(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "out.txt")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	dsts := []*teeWriter{
		{w: failingWriter{}, label: "bad"},
		{w: f, label: path},
	}
	stderr, runErr := teeTo(t, "payload\n", dsts, outputErrorWarn)
	if runErr == nil {
		t.Fatal("warn mode: expected nonzero exit when a writer fails")
	}
	if !strings.Contains(stderr, "tee: ") {
		t.Errorf("warn mode: stderr = %q, want a tee diagnostic", stderr)
	}
	got, rerr := os.ReadFile(path)
	if rerr != nil {
		t.Fatal(rerr)
	}
	if string(got) != "payload\n" {
		t.Errorf("warn mode: writable file = %q, want %q (must keep writing)", string(got), "payload\n")
	}
}

// TestExitModeStopsEarly verifies that exit mode stops at the first write error,
// so a later writer in the list never receives the data.
func TestExitModeStopsEarly(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "out.txt")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	// The failing writer comes first; exit mode must stop before the file write.
	dsts := []*teeWriter{
		{w: failingWriter{}, label: "bad"},
		{w: f, label: path},
	}
	if _, runErr := teeTo(t, "payload\n", dsts, outputErrorExit); runErr == nil {
		t.Fatal("exit mode: expected nonzero exit when a writer fails")
	}
	got, rerr := os.ReadFile(path)
	if rerr != nil {
		t.Fatal(rerr)
	}
	if len(got) != 0 {
		t.Errorf("exit mode: writable file = %q, want empty (must stop on first error)", string(got))
	}
}

// TestWarnNopipeIgnoresPipe confirms a broken-pipe write error is silently
// tolerated by the *-nopipe modes but reported by warn.
func TestWarnNopipeIgnoresPipe(t *testing.T) {
	t.Parallel()
	if reportWriteError(command.IO{Err: io.Discard}, "bad", syscallPipeErr(), outputErrorWarnNopipe) {
		t.Error("warn-nopipe: broken pipe should not count toward failure")
	}
	if !reportWriteError(command.IO{Err: io.Discard}, "bad", syscallPipeErr(), outputErrorWarn) {
		t.Error("warn: broken pipe should be reported")
	}
}
