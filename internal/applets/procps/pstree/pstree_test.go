package pstree

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

// fixture writes /proc/PID/stat files for the given pid->(comm, ppid) tree.
func fixture(t *testing.T, procs map[int][2]interface{}) {
	t.Helper()
	dir := t.TempDir()
	for pid, info := range procs {
		pdir := filepath.Join(dir, fmt.Sprintf("%d", pid))
		if err := os.MkdirAll(pdir, 0o755); err != nil {
			t.Fatal(err)
		}
		stat := fmt.Sprintf("%d (%s) S %d 0 0", pid, info[0], info[1])
		if err := os.WriteFile(filepath.Join(pdir, "stat"), []byte(stat), 0o644); err != nil {
			t.Fatal(err)
		}
	}
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

func TestTree(t *testing.T) {
	fixture(t, map[int][2]interface{}{
		1:   {"init", 0},
		100: {"sshd", 1},
		200: {"bash", 100},
		150: {"cron", 1},
	})
	out := run(t)
	for _, want := range []string{"init(1)", "sshd(100)", "bash(200)", "cron(150)"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
	// init is the root (printed first, no indentation).
	if !strings.HasPrefix(out, "init(1)\n") {
		t.Errorf("init should be the root: %q", out)
	}
	// bash appears after sshd (it is sshd's child).
	if strings.Index(out, "sshd(100)") > strings.Index(out, "bash(200)") {
		t.Errorf("bash should be under sshd:\n%s", out)
	}
}

func TestParseStat(t *testing.T) {
	t.Parallel()
	comm, ppid, ok := parseStat("1234 (cat) S 100 1234 1234")
	if !ok || comm != "cat" || ppid != 100 {
		t.Errorf("parseStat = %q, %d, %v", comm, ppid, ok)
	}
	// A comm containing spaces and parentheses.
	comm, ppid, ok = parseStat("5 (foo (bar) baz) R 7 0 0")
	if !ok || comm != "foo (bar) baz" || ppid != 7 {
		t.Errorf("parseStat with parens = %q, %d, %v", comm, ppid, ok)
	}
	if _, _, ok := parseStat("garbage"); ok {
		t.Errorf("garbage should not parse")
	}
}

func TestOrphanRoots(t *testing.T) {
	// A process whose parent is not in /proc is treated as a root.
	fixture(t, map[int][2]interface{}{
		500: {"daemon", 999}, // ppid 999 has no stat file
	})
	out := run(t)
	if !strings.HasPrefix(out, "daemon(500)\n") {
		t.Errorf("orphan should be a root: %q", out)
	}
}

func TestHelpExitStatus(t *testing.T) {
	t.Parallel()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, []string{"--help"}); err != nil {
		t.Fatalf("Run --help error = %v", err)
	}
	if !strings.Contains(out.String(), "Exit status:") {
		t.Errorf("--help missing Exit status section:\n%s", out.String())
	}
}
