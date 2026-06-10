package minips

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func fixture(t *testing.T, procs map[string]string) {
	t.Helper()
	dir := t.TempDir()
	for pid, c := range procs {
		pdir := filepath.Join(dir, pid)
		if err := os.MkdirAll(pdir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(pdir, "comm"), []byte(c+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	_ = os.MkdirAll(filepath.Join(dir, "notapid"), 0o755) // ignored
	orig := procDir
	procDir = dir
	t.Cleanup(func() { procDir = orig })
}

func run(t *testing.T) string {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, nil); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	return out.String()
}

func TestListing(t *testing.T) {
	fixture(t, map[string]string{"1": "init", "100": "bash", "50": "sshd"})
	out := run(t)
	if !strings.HasPrefix(out, "PID") || !strings.Contains(out, "USER") || !strings.Contains(out, "COMMAND") {
		t.Errorf("header wrong: %q", out)
	}
	for _, want := range []string{"init", "bash", "sshd"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
	// Sorted: 1 before 50 before 100.
	if !(strings.Index(out, "init") < strings.Index(out, "sshd") && strings.Index(out, "sshd") < strings.Index(out, "bash")) {
		t.Errorf("not sorted by PID:\n%s", out)
	}
	// The non-numeric directory is ignored (4 lines: header + 3 procs).
	if n := strings.Count(strings.TrimRight(out, "\n"), "\n"); n != 3 {
		t.Errorf("expected 3 process rows, got %d:\n%s", n, out)
	}
}
