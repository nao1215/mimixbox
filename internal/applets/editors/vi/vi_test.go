package vi_test

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/editors/vi"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, stdin string, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	// A bytes.Reader is not an *os.File, so vi runs in batch (non-terminal) mode.
	io := command.IO{In: strings.NewReader(stdin), Out: out, Err: errBuf}
	err := vi.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

// errReader fails on the first read, modeling an unreadable stdin (e.g. stdin
// redirected from a directory, which makes read(0) fail with EISDIR).
type errReader struct{}

func (errReader) Read([]byte) (int, error) { return 0, errors.New("is a directory") }

func TestBatchStdinReadErrorIsReported(t *testing.T) {
	t.Parallel()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: errReader{}, Out: out, Err: errBuf}
	err := vi.New().Run(context.Background(), io, nil)
	if err == nil {
		t.Fatal("vi must report an unreadable stdin instead of treating it as empty")
	}
	if !strings.Contains(errBuf.String(), "is a directory") {
		t.Errorf("stderr = %q, want the read error", errBuf.String())
	}
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := vi.New()
	if got := c.Name(); got != "vi" {
		t.Errorf("Name() = %q", got)
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis empty")
	}
}

func TestEditAndSave(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	f := filepath.Join(dir, "doc.txt")
	if err := os.WriteFile(f, []byte("hello\nworld\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Delete the first character of line 1, then write & quit.
	if _, errOut, err := run(t, "x:wq\r", f); err != nil {
		t.Fatalf("err = %v (stderr=%q)", err, errOut)
	}
	got, err := os.ReadFile(f)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "ello\nworld\n" {
		t.Errorf("file = %q, want %q", got, "ello\nworld\n")
	}
}

func TestInsertAndSave(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	f := filepath.Join(dir, "doc.txt")
	if err := os.WriteFile(f, []byte("bar\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Insert "foo" before the buffer, ESC, write & quit.
	if _, errOut, err := run(t, "ifoo\x1b:wq\r", f); err != nil {
		t.Fatalf("err = %v (stderr=%q)", err, errOut)
	}
	got, _ := os.ReadFile(f)
	if string(got) != "foobar\n" {
		t.Errorf("file = %q, want foobar", got)
	}
}

func TestNoWriteWithoutSave(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	f := filepath.Join(dir, "doc.txt")
	if err := os.WriteFile(f, []byte("keep\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Edit but quit with :q! (no save) -> file unchanged.
	if _, _, err := run(t, "x:q!\r", f); err != nil {
		t.Fatalf("err = %v", err)
	}
	got, _ := os.ReadFile(f)
	if string(got) != "keep\n" {
		t.Errorf("file = %q, want unchanged 'keep'", got)
	}
}

func TestCreateNewFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	f := filepath.Join(dir, "new.txt")
	// File does not exist yet; create content and save.
	if _, errOut, err := run(t, "inew content\x1b:wq\r", f); err != nil {
		t.Fatalf("err = %v (stderr=%q)", err, errOut)
	}
	got, err := os.ReadFile(f)
	if err != nil {
		t.Fatalf("expected file created: %v", err)
	}
	if string(got) != "new content\n" {
		t.Errorf("file = %q", got)
	}
}

func TestSaveWithoutFilename(t *testing.T) {
	t.Parallel()
	// No filename operand, but the script asks to write -> error.
	_, errOut, err := run(t, "ihi\x1b:wq\r")
	if err == nil {
		t.Error("expected error saving with no file name")
	}
	if !strings.Contains(errOut, "no file name") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestHelp(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "", "--help")
	if err != nil {
		t.Fatalf("help err = %v", err)
	}
	if !strings.Contains(out, "Usage: vi") {
		t.Errorf("help = %q", out)
	}
	if !strings.Contains(out, "Exit status:") {
		t.Errorf("help missing exit status section = %q", out)
	}
}
