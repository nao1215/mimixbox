package logread

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	err := New().Run(context.Background(), io, args)
	return out.String(), err
}

func writeFile(t *testing.T, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "log")
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestExplicitFile(t *testing.T) {
	f := writeFile(t, "alpha\nbeta\n")
	out, err := run(t, f)
	if err != nil {
		t.Fatal(err)
	}
	if out != "alpha\nbeta\n" {
		t.Errorf("logread FILE = %q", out)
	}
}

func TestDefaultCandidate(t *testing.T) {
	f := writeFile(t, "system log\n")
	orig := logCandidates
	logCandidates = []string{"/no/such/log", f}
	defer func() { logCandidates = orig }()
	out, err := run(t)
	if err != nil {
		t.Fatal(err)
	}
	if out != "system log\n" {
		t.Errorf("fallback candidate = %q", out)
	}
}

func TestNoLog(t *testing.T) {
	orig := logCandidates
	logCandidates = []string{"/no/such/a", "/no/such/b"}
	defer func() { logCandidates = orig }()
	if _, err := run(t); err == nil {
		t.Errorf("no readable log should fail")
	}
}
