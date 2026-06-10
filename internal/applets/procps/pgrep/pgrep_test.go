package pgrep

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func fixture(t *testing.T, procs map[string]string) {
	t.Helper()
	dir := t.TempDir()
	for pid, comm := range procs {
		pdir := filepath.Join(dir, pid)
		if err := os.MkdirAll(pdir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(pdir, "comm"), []byte(comm+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	// A non-numeric entry must be ignored.
	_ = os.MkdirAll(filepath.Join(dir, "self"), 0o755)
	orig := procDir
	procDir = dir
	t.Cleanup(func() { procDir = orig })
}

type sig struct {
	pid int
	num syscall.Signal
}

func stubSignal(t *testing.T) *[]sig {
	t.Helper()
	var calls []sig
	orig := sendSignal
	sendSignal = func(pid int, s syscall.Signal) error { calls = append(calls, sig{pid, s}); return nil }
	t.Cleanup(func() { sendSignal = orig })
	return &calls
}

func run(t *testing.T, c *Command, args ...string) (string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	err := c.Run(context.Background(), io, args)
	return strings.TrimSpace(out.String()), err
}

func TestPgrepMatches(t *testing.T) {
	fixture(t, map[string]string{"10": "sshd", "20": "bash", "30": "sshd"})
	out, err := run(t, NewPgrep(), "sshd")
	if err != nil {
		t.Fatal(err)
	}
	if out != "10\n30" {
		t.Errorf("pgrep = %q, want \"10\\n30\"", out)
	}
}

func TestPgrepNoMatch(t *testing.T) {
	fixture(t, map[string]string{"10": "bash"})
	if _, err := run(t, NewPgrep(), "nope"); err == nil {
		t.Errorf("no match should exit non-zero")
	}
}

func TestPkillSignals(t *testing.T) {
	fixture(t, map[string]string{"10": "vim", "20": "vim", "30": "less"})
	calls := stubSignal(t)
	if _, err := run(t, NewPkill(), "vim"); err != nil {
		t.Fatal(err)
	}
	if len(*calls) != 2 {
		t.Fatalf("signalled %d processes, want 2", len(*calls))
	}
	for _, c := range *calls {
		if c.num != syscall.SIGTERM {
			t.Errorf("default signal = %v, want SIGTERM", c.num)
		}
	}
}

func TestPkillSignalShorthand(t *testing.T) {
	fixture(t, map[string]string{"10": "vim"})
	calls := stubSignal(t)
	if _, err := run(t, NewPkill(), "-9", "vim"); err != nil {
		t.Fatal(err)
	}
	if len(*calls) != 1 || (*calls)[0].num != syscall.SIGKILL {
		t.Errorf("pkill -9 calls = %v, want SIGKILL", *calls)
	}
}

func TestNoPattern(t *testing.T) {
	fixture(t, map[string]string{"10": "x"})
	if _, err := run(t, NewPgrep()); err == nil {
		t.Errorf("no pattern should fail")
	}
}
