package shellfront

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, c *Command, in string, args ...string) (string, string) {
	t.Helper()
	// Use a real pipe so launched commands share an *os.File stdin.
	pr, pw, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	go func() { _, _ = pw.WriteString(in); _ = pw.Close() }()
	out, errBuf := &bytes.Buffer{}, &bytes.Buffer{}
	io := command.IO{In: pr, Out: out, Err: errBuf}
	if rerr := c.Run(context.Background(), io, args); rerr != nil {
		t.Fatalf("Run error = %v (stderr=%q)", rerr, errBuf.String())
	}
	_ = pr.Close()
	return out.String(), errBuf.String()
}

func requireCmd(t *testing.T, name string) {
	t.Helper()
	if _, err := exec.LookPath(name); err != nil {
		t.Skipf("%s not on PATH: %v", name, err)
	}
}

func TestDashCRunsCommand(t *testing.T) {
	requireCmd(t, "echo")
	out, _ := run(t, NewSh(), "", "-c", "echo hello")
	if !strings.Contains(out, "hello") {
		t.Errorf("sh -c output = %q", out)
	}
	if strings.Contains(out, "mbsh:") {
		t.Errorf("non-interactive sh must not print a prompt: %q", out)
	}
}

func TestScriptFile(t *testing.T) {
	requireCmd(t, "echo")
	dir := t.TempDir()
	script := filepath.Join(dir, "s.sh")
	if err := os.WriteFile(script, []byte("echo fromfile\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	out, _ := run(t, NewBash(), "", script)
	if !strings.Contains(out, "fromfile") {
		t.Errorf("script output = %q", out)
	}
}

func TestStdinNonTTYNoPrompt(t *testing.T) {
	requireCmd(t, "echo")
	out, _ := run(t, NewAsh(), "echo fromstdin\n")
	if !strings.Contains(out, "fromstdin") || strings.Contains(out, "mbsh:") {
		t.Errorf("ash from stdin output = %q (should have no prompt)", out)
	}
}

func TestHelp(t *testing.T) {
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := NewBash().Run(context.Background(), io, []string{"--help"}); err != nil {
		t.Fatalf("--help err = %v", err)
	}
	if !strings.Contains(out.String(), "Usage: bash") {
		t.Errorf("--help out = %q", out.String())
	}
}
