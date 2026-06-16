package split_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/textutils/split"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, stdin string, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(stdin), Out: out, Err: errBuf}
	err := split.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestByLines(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	prefix := filepath.Join(dir, "part-")
	_, _, err := run(t, "1\n2\n3\n4\n5\n", "-l", "2", "-", prefix)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	checkFile(t, prefix+"aa", "1\n2\n")
	checkFile(t, prefix+"ab", "3\n4\n")
	checkFile(t, prefix+"ac", "5\n")
}

func TestByBytes(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	prefix := filepath.Join(dir, "b-")
	_, _, err := run(t, "abcdefg", "-b", "3", "-", prefix)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	checkFile(t, prefix+"aa", "abc")
	checkFile(t, prefix+"ab", "def")
	checkFile(t, prefix+"ac", "g")
}

func TestInvalidLines(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, "x\n", "-l", "0")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(errOut, "invalid number of lines") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestInvalidBytes(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, "x\n", "-b", "notanumber")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(errOut, "invalid number of bytes") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestMissingInput(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, "", "/no/such/file")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(errOut, "split: /no/such/file:") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := split.New()
	if c.Name() != "split" {
		t.Errorf("Name() = %q", c.Name())
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}

func checkFile(t *testing.T, path, want string) {
	t.Helper()
	got, err := os.ReadFile(path) //nolint:gosec // test reads a file it just wrote
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if string(got) != want {
		t.Errorf("%s = %q, want %q", path, got, want)
	}
}

func TestHelpSections(t *testing.T) {
	t.Parallel()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := split.New().Run(context.Background(), io, []string{"--help"}); err != nil {
		t.Fatalf("Run --help error = %v", err)
	}
	for _, want := range []string{"Examples:", "Exit status:"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("--help missing %q section:\n%s", want, out.String())
		}
	}
}
