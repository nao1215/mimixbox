package ar_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/archival/ar"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := ar.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := ar.New()
	if got := c.Name(); got != "ar" {
		t.Errorf("Name() = %q", got)
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis empty")
	}
}

func TestCreateListExtract(t *testing.T) {
	// Uses t.Chdir, so it cannot be parallel.
	dir := t.TempDir()
	a := filepath.Join(dir, "a.txt")
	b := filepath.Join(dir, "b.txt")
	write(t, a, "alpha")
	write(t, b, "beta-odd") // 8 bytes (even); use odd below too
	archive := filepath.Join(dir, "lib.a")

	if _, errOut, err := run(t, "rc", archive, a, b); err != nil {
		t.Fatalf("create err = %v (stderr=%q)", err, errOut)
	}

	out, _, err := run(t, "t", archive)
	if err != nil {
		t.Fatalf("list err = %v", err)
	}
	if !strings.Contains(out, "a.txt") || !strings.Contains(out, "b.txt") {
		t.Errorf("list = %q, want both members", out)
	}

	// Extract into a fresh dir (extract writes to the current directory).
	extractDir := filepath.Join(dir, "out")
	if err := os.Mkdir(extractDir, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(extractDir)
	if _, errOut, err := run(t, "x", archive); err != nil {
		t.Fatalf("extract err = %v (stderr=%q)", err, errOut)
	}
	got, err := os.ReadFile(filepath.Join(extractDir, "a.txt"))
	if err != nil {
		t.Fatalf("read extracted: %v", err)
	}
	if string(got) != "alpha" {
		t.Errorf("extracted a.txt = %q", got)
	}
}

func TestOddSizePadding(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	a := filepath.Join(dir, "odd.txt")
	write(t, a, "abc") // 3 bytes -> needs padding
	archive := filepath.Join(dir, "o.a")
	if _, _, err := run(t, "r", archive, a); err != nil {
		t.Fatalf("create err = %v", err)
	}
	out, _, err := run(t, "t", archive)
	if err != nil {
		t.Fatalf("list err = %v", err)
	}
	if strings.TrimSpace(out) != "odd.txt" {
		t.Errorf("list = %q", out)
	}
}

func TestExtractSpecificMember(t *testing.T) {
	// Uses t.Chdir, so it cannot be parallel.
	dir := t.TempDir()
	a := filepath.Join(dir, "one.txt")
	b := filepath.Join(dir, "two.txt")
	write(t, a, "1")
	write(t, b, "2")
	archive := filepath.Join(dir, "m.a")
	if _, _, err := run(t, "r", archive, a, b); err != nil {
		t.Fatalf("create err = %v", err)
	}
	extractDir := filepath.Join(dir, "ex")
	if err := os.Mkdir(extractDir, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(extractDir)
	if _, _, err := run(t, "x", archive, "two.txt"); err != nil {
		t.Fatalf("extract err = %v", err)
	}
	if _, err := os.Stat(filepath.Join(extractDir, "two.txt")); err != nil {
		t.Errorf("expected two.txt extracted: %v", err)
	}
	if _, err := os.Stat(filepath.Join(extractDir, "one.txt")); err == nil {
		t.Error("one.txt should not have been extracted")
	}
}

func TestVerbose(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	a := filepath.Join(dir, "v.txt")
	write(t, a, "x")
	archive := filepath.Join(dir, "v.a")
	out, _, err := run(t, "rv", archive, a)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if !strings.Contains(out, "v.txt") {
		t.Errorf("verbose out = %q", out)
	}
}

func TestErrors(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"too few args", []string{"t"}, "usage"},
		{"bad key", []string{"q", filepath.Join(dir, "x.a")}, "invalid operation key"},
		{"create no members", []string{"r", filepath.Join(dir, "y.a")}, "no members"},
		{"list missing archive", []string{"t", filepath.Join(dir, "nope.a")}, "ar:"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, errOut, err := run(t, tt.args...)
			if err == nil {
				t.Errorf("expected error for %v", tt.args)
			}
			if !strings.Contains(errOut, tt.want) {
				t.Errorf("stderr = %q, want %q", errOut, tt.want)
			}
		})
	}
}

func TestNotAnArchive(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	bad := filepath.Join(dir, "bad.a")
	write(t, bad, "garbage")
	_, errOut, err := run(t, "t", bad)
	if err == nil {
		t.Error("expected error for non-archive")
	}
	if !strings.Contains(errOut, "not an ar archive") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestHelp(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "--help")
	if err != nil {
		t.Fatalf("help err = %v", err)
	}
	if !strings.Contains(out, "Usage: ar") {
		t.Errorf("help = %q", out)
	}
}
