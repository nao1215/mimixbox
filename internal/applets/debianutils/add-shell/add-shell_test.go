package addShell_test

import (
	"bufio"
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	addShell "github.com/nao1215/mimixbox/internal/applets/debianutils/add-shell"
	"github.com/nao1215/mimixbox/internal/command"
)

func TestNew(t *testing.T) {
	t.Parallel()
	if addShell.New() == nil {
		t.Fatal("New() returned nil")
	}
}

func writeShells(t *testing.T, lines ...string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "shells")
	content := ""
	if len(lines) > 0 {
		content = strings.Join(lines, "\n") + "\n"
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func readLines(t *testing.T, path string) []string {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	var lines []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	return lines
}

func TestAddShellsAppendsNew(t *testing.T) {
	t.Parallel()
	path := writeShells(t, "/bin/sh", "/bin/bash")

	if err := addShell.AddShellsForTest(path, []string{"/bin/zsh"}); err != nil {
		t.Fatalf("addShells error = %v", err)
	}

	got := readLines(t, path)
	want := []string{"/bin/sh", "/bin/bash", "/bin/zsh"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("lines = %v, want %v", got, want)
	}
}

func TestAddShellsExistingIsNoOp(t *testing.T) {
	t.Parallel()
	path := writeShells(t, "/bin/sh", "/bin/bash")

	if err := addShell.AddShellsForTest(path, []string{"/bin/bash"}); err != nil {
		t.Fatalf("addShells error = %v", err)
	}

	got := readLines(t, path)
	want := []string{"/bin/sh", "/bin/bash"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("lines = %v, want %v (existing shell should be a no-op)", got, want)
	}
}

func TestHelpSections(t *testing.T) {
	t.Parallel()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := addShell.New().Run(context.Background(), io, []string{"--help"}); err != nil {
		t.Fatalf("--help error = %v", err)
	}
	if !strings.Contains(out.String(), "Examples:") {
		t.Errorf("--help missing Examples: %q", out.String())
	}
	if !strings.Contains(out.String(), "Exit status:") {
		t.Errorf("--help missing Exit status: %q", out.String())
	}
}
