package serial_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/shellutils/serial"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := serial.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func writeFiles(t *testing.T, dir string, names ...string) {
	t.Helper()
	for _, n := range names {
		if err := os.WriteFile(filepath.Join(dir, n), []byte(n), 0o600); err != nil {
			t.Fatal(err)
		}
	}
}

func listDir(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	sort.Strings(names)
	return names
}

func TestRunPrefixDefault(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, "a.txt", "b.txt", "c.txt")

	_, _, err := run(t, dir)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}

	got := listDir(t, dir)
	want := []string{"0_a.txt", "1_b.txt", "2_c.txt"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("renamed files = %v, want %v", got, want)
	}
}

func TestRunSuffixWithName(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, "a.txt", "b.txt", "c.txt")

	_, _, err := run(t, "--suffix", "--name=demo", dir)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}

	got := listDir(t, dir)
	want := []string{"demo_0.txt", "demo_1.txt", "demo_2.txt"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("renamed files = %v, want %v", got, want)
	}
}

func TestRunKeepCopiesOriginals(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, "a.txt", "b.txt")

	out, _, err := run(t, "--keep", dir)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(out, "Copy ") {
		t.Errorf("expected Copy output, got %q", out)
	}

	got := listDir(t, dir)
	want := []string{"0_a.txt", "1_b.txt", "a.txt", "b.txt"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("files = %v, want %v", got, want)
	}
}

func TestRunDryRunDoesNotRename(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, "a.txt", "b.txt")

	out, _, err := run(t, "--dry-run", dir)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(out, "Rename ") {
		t.Errorf("expected Rename output, got %q", out)
	}

	got := listDir(t, dir)
	want := []string{"a.txt", "b.txt"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("dry-run modified files = %v, want %v", got, want)
	}
}

func TestRunZeroPaddingWidth(t *testing.T) {
	dir := t.TempDir()
	var names []string
	for i := 0; i < 11; i++ {
		names = append(names, string(rune('a'+i))+".txt")
	}
	writeFiles(t, dir, names...)

	if _, _, err := run(t, dir); err != nil {
		t.Fatalf("Run error = %v", err)
	}

	got := listDir(t, dir)
	// 11 files -> width is len("11") = 2, so numbers are zero padded to 2.
	if got[0] != "00_a.txt" {
		t.Errorf("first file = %q, want 00_a.txt", got[0])
	}
	if got[len(got)-1] != "10_k.txt" {
		t.Errorf("last file = %q, want 10_k.txt", got[len(got)-1])
	}
}

func TestRunMissingOperand(t *testing.T) {
	out, errOut, err := run(t)
	if err == nil {
		t.Fatal("expected error for missing operand")
	}
	if out != "" {
		t.Errorf("out = %q, want empty", out)
	}
	if !strings.Contains(errOut, "serial: missing operand") {
		t.Errorf("stderr = %q, want missing operand message", errOut)
	}
}

func TestRunNonexistentDir(t *testing.T) {
	_, _, err := run(t, "/no/such/dir")
	if err == nil {
		t.Fatal("expected error for nonexistent directory")
	}
}

func TestRunEmptyDir(t *testing.T) {
	dir := t.TempDir()
	_, _, err := run(t, dir)
	if err == nil {
		t.Fatal("expected error for empty directory")
	}
}
