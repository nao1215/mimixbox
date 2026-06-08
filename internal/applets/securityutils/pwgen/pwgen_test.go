package pwgen

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestDefaultLengthAndCount(t *testing.T) {
	t.Parallel()
	out, _, err := run(t)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 1 {
		t.Fatalf("got %d passwords, want 1", len(lines))
	}
	if len(lines[0]) != 16 {
		t.Errorf("length = %d, want 16", len(lines[0]))
	}
}

func TestCountAndLength(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "-n", "5", "-l", "8")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 5 {
		t.Fatalf("got %d passwords, want 5", len(lines))
	}
	for _, l := range lines {
		if len(l) != 8 {
			t.Errorf("length = %d, want 8 (%q)", len(l), l)
		}
	}
}

func TestNoDigits(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "-0", "-l", "200")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if strings.ContainsAny(strings.TrimSpace(out), digits) {
		t.Errorf("password contains digits despite -0: %q", out)
	}
}

func TestSymbolsCharsetMembership(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "-s", "-l", "500")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	pw := strings.TrimSpace(out)
	allowed := lowers + uppers + digits + symbols
	for _, r := range pw {
		if !strings.ContainsRune(allowed, r) {
			t.Errorf("password contains out-of-charset rune %q", r)
		}
	}
}

func TestOutputFile(t *testing.T) {
	t.Parallel()
	p := filepath.Join(t.TempDir(), "wl.txt")
	out, _, err := run(t, "-n", "3", "-o", p)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "" {
		t.Errorf("stdout should be empty when -o is used, got %q", out)
	}
	data, err := os.ReadFile(p) //nolint:gosec // reading a file the test wrote
	if err != nil {
		t.Fatal(err)
	}
	if n := len(strings.Split(strings.TrimRight(string(data), "\n"), "\n")); n != 3 {
		t.Errorf("file has %d lines, want 3", n)
	}
}

func TestInvalidLength(t *testing.T) {
	t.Parallel()
	_, _, err := run(t, "-l", "0")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "length must be at least 1") {
		t.Errorf("err = %v", err)
	}
}

func TestInvalidCount(t *testing.T) {
	t.Parallel()
	_, _, err := run(t, "-n", "0")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "number must be at least 1") {
		t.Errorf("err = %v", err)
	}
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := New()
	if c.Name() != "pwgen" {
		t.Errorf("Name() = %q", c.Name())
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}
