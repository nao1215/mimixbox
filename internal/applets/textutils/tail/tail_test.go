package tail_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/nao1215/mimixbox/internal/applets/textutils/tail"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, stdin string, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(stdin), Out: out, Err: errBuf}
	err := tail.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestRunStdin(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		stdin string
		args  []string
		want  string
	}{
		{"default 10", "1\n2\n3\n", nil, "1\n2\n3\n"},
		{"lines flag", "1\n2\n3\n4\n", []string{"-n", "2"}, "3\n4\n"},
		{"bytes flag", "hello world", []string{"-c", "5"}, "world"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out, _, err := run(t, tt.stdin, tt.args...)
			if err != nil {
				t.Fatalf("Run error = %v", err)
			}
			if out != tt.want {
				t.Errorf("out = %q, want %q", out, tt.want)
			}
		})
	}
}

// TestRunZeroTerminated checks that -z/--zero-terminated treats NUL as the
// record delimiter, counts NUL-delimited records for -n, and preserves any
// newlines embedded within a record.
func TestRunZeroTerminated(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		stdin string
		args  []string
		want  string
	}{
		{"z last record", "a\nb\x00c\nd\x00", []string{"-z", "-n", "1"}, "c\nd\x00"},
		{"z long flag", "a\nb\x00c\nd\x00", []string{"--zero-terminated", "-n", "1"}, "c\nd\x00"},
		{"z two records", "a\nb\x00c\nd\x00", []string{"-z", "-n", "2"}, "a\nb\x00c\nd\x00"},
		{"z bytes unaffected", "a\nb\x00c\nd\x00", []string{"-z", "-c", "3"}, "\nd\x00"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out, _, err := run(t, tt.stdin, tt.args...)
			if err != nil {
				t.Fatalf("Run error = %v", err)
			}
			if out != tt.want {
				t.Errorf("out = %q, want %q", out, tt.want)
			}
		})
	}
}

func TestRunMultipleFilesHaveHeaders(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	a := filepath.Join(dir, "a.txt")
	b := filepath.Join(dir, "b.txt")
	_ = os.WriteFile(a, []byte("aaa\n"), 0o600)
	_ = os.WriteFile(b, []byte("bbb\n"), 0o600)

	out, _, err := run(t, "", "-n", "1", a, b)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	want := "==> " + a + " <==\naaa\n\n==> " + b + " <==\nbbb\n"
	if out != want {
		t.Errorf("out = %q, want %q", out, want)
	}
}

func TestRunMissingFile(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, "", "/no/such/file")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(errOut, "tail: /no/such/file:") {
		t.Errorf("stderr = %q", errOut)
	}
}

func appendToFile(t *testing.T, path, data string) {
	t.Helper()
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		t.Fatalf("open for append: %v", err)
	}
	defer func() { _ = f.Close() }()
	if _, err := f.WriteString(data); err != nil {
		t.Fatalf("append: %v", err)
	}
}

// runFollowBackground starts tail in a goroutine and returns the live stdout
// buffer, a cancel function that stops the follow loop, and a channel that is
// closed once Run returns. Only the tail goroutine writes to the buffer, and
// callers read it after receiving from done, so access is synchronized.
func runFollowBackground(ctx context.Context, args ...string) (*bytes.Buffer, <-chan error) {
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	done := make(chan error, 1)
	go func() { done <- tail.New().Run(ctx, io, args) }()
	return out, done
}

func TestFollowEmitsAppendedData(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(path, []byte("line1\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	out, done := runFollowBackground(ctx, "-f", "-s", "0.05", path)

	time.Sleep(150 * time.Millisecond)
	appendToFile(t, path, "line2\nline3\n")
	time.Sleep(250 * time.Millisecond)
	cancel()
	if err := <-done; err != nil {
		t.Fatalf("Run error = %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "line1\n") {
		t.Errorf("initial tail missing in %q", got)
	}
	if !strings.Contains(got, "line2\nline3\n") {
		t.Errorf("appended data missing in %q", got)
	}
}

func TestFollowReturnsOnCancel(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(path, []byte("x\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	_, done := runFollowBackground(ctx, "-f", "-s", "0.05", path)
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run error = %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("follow did not stop after cancel")
	}
}

func TestFollowNameReopensRecreatedFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rotating.txt")
	if err := os.WriteFile(path, []byte("first\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	out, done := runFollowBackground(ctx, "-F", "-s", "0.05", path)

	time.Sleep(150 * time.Millisecond)
	// Rotate: rename the original away and recreate the path with new content.
	if err := os.Rename(path, path+".1"); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("second\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	time.Sleep(300 * time.Millisecond)
	cancel()
	if err := <-done; err != nil {
		t.Fatalf("Run error = %v", err)
	}

	if got := out.String(); !strings.Contains(got, "second\n") {
		t.Errorf("recreated file content missing in %q", got)
	}
}

func TestInvalidSleepIntervalRejected(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(path, []byte("x\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, _, err := run(t, "", "-f", "-s", "0", path)
	if err == nil {
		t.Fatal("expected error for non-positive sleep interval")
	}
	if !strings.Contains(err.Error(), "invalid number of seconds") {
		t.Errorf("err = %v", err)
	}
}

// TestInvalidPidRejected covers --pid validation: a non-positive PID is an
// error, matching GNU tail's "invalid PID" diagnostic.
func TestInvalidPidRejected(t *testing.T) {
	t.Parallel()
	for _, arg := range []string{"--pid=0", "--pid=-3"} {
		_, _, err := run(t, "", "-f", "-s", "0.05", arg, "/dev/null")
		if err == nil {
			t.Fatalf("%s: expected error", arg)
		}
		if !strings.Contains(err.Error(), "invalid PID") {
			t.Errorf("%s: err = %v, want invalid PID", arg, err)
		}
	}
}

// TestPidWithoutFollowWarns confirms --pid without a follow flag emits the GNU
// "PID ignored" warning and otherwise behaves like a plain tail.
func TestPidWithoutFollowWarns(t *testing.T) {
	t.Parallel()
	out, errOut, err := run(t, "a\nb\nc\n", "--pid=123", "-n", "1")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "c\n" {
		t.Errorf("out = %q, want %q", out, "c\n")
	}
	if !strings.Contains(errOut, "PID ignored") {
		t.Errorf("stderr = %q, want PID ignored warning", errOut)
	}
}

// TestFollowDescriptorDoesNotReopen verifies the -f / --follow=descriptor
// default: tail keeps reading the original descriptor across a rotation and does
// not switch to the recreated path, unlike --follow=name.
func TestFollowDescriptorDoesNotReopen(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rotating.txt")
	if err := os.WriteFile(path, []byte("first\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	out, done := runFollowBackground(ctx, "--follow=descriptor", "-s", "0.05", path)

	time.Sleep(150 * time.Millisecond)
	// Rotate the original away and recreate the path with new content. A
	// descriptor follower keeps the old fd and must not pick up "second".
	if err := os.Rename(path, path+".1"); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("second\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	time.Sleep(300 * time.Millisecond)
	cancel()
	if err := <-done; err != nil {
		t.Fatalf("Run error = %v", err)
	}

	if got := out.String(); strings.Contains(got, "second\n") {
		t.Errorf("descriptor follow should not show recreated content, got %q", got)
	}
}

func TestInvalidFollowModeRejected(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(path, []byte("x\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, _, err := run(t, "", "--follow=bogus", path)
	if err == nil {
		t.Fatal("expected error for invalid --follow mode")
	}
	if !strings.Contains(err.Error(), "invalid argument") {
		t.Errorf("err = %v", err)
	}
}
