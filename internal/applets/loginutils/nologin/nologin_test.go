package nologin

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

func TestDefaultMessageAndExit(t *testing.T) {
	orig := messageFile
	messageFile = "/no/such/nologin.txt"
	defer func() { messageFile = orig }()
	out, err := run(t)
	if err == nil {
		t.Errorf("nologin must exit non-zero")
	}
	if !strings.Contains(out, defaultMessage) {
		t.Errorf("output = %q", out)
	}
}

func TestCustomMessage(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "nologin.txt")
	_ = os.WriteFile(f, []byte("System is down for maintenance.\n"), 0o644)
	orig := messageFile
	messageFile = f
	defer func() { messageFile = orig }()
	out, err := run(t)
	if err == nil {
		t.Errorf("nologin must exit non-zero")
	}
	if out != "System is down for maintenance.\n" {
		t.Errorf("custom message = %q", out)
	}
}

func TestIgnoresArguments(t *testing.T) {
	orig := messageFile
	messageFile = "/no/such/nologin.txt"
	defer func() { messageFile = orig }()
	// A shell-style "-c command" must be ignored and still refuse.
	out, err := run(t, "-c", "echo pwned")
	if err == nil {
		t.Errorf("nologin must exit non-zero even with arguments")
	}
	if strings.Contains(out, "pwned") {
		t.Errorf("nologin must not run the command")
	}
}
