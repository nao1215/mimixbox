package svc

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func service(t *testing.T, supervised bool) string {
	t.Helper()
	dir := t.TempDir()
	if supervised {
		sup := filepath.Join(dir, "supervise")
		if err := os.MkdirAll(sup, 0o755); err != nil {
			t.Fatal(err)
		}
		_ = os.WriteFile(filepath.Join(sup, "control"), nil, 0o644)
	}
	return dir
}

func run(t *testing.T, args ...string) error {
	t.Helper()
	io := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	return New().Run(context.Background(), io, args)
}

func control(t *testing.T, dir string) string {
	t.Helper()
	data, _ := os.ReadFile(filepath.Join(dir, "supervise", "control"))
	return string(data)
}

func TestSingleCommand(t *testing.T) {
	dir := service(t, true)
	if err := run(t, "-d", dir); err != nil {
		t.Fatal(err)
	}
	if control(t, dir) != "d" {
		t.Errorf("control = %q, want d", control(t, dir))
	}
}

func TestCombinedCommandsInOrder(t *testing.T) {
	dir := service(t, true)
	// -u given before -t on the line, but the canonical order sends "tu".
	if err := run(t, "-u", "-t", dir); err != nil {
		t.Fatal(err)
	}
	if control(t, dir) != "tu" {
		t.Errorf("control = %q, want tu", control(t, dir))
	}
}

func TestMultipleDirs(t *testing.T) {
	a, b := service(t, true), service(t, true)
	if err := run(t, "-x", a, b); err != nil {
		t.Fatal(err)
	}
	if control(t, a) != "x" || control(t, b) != "x" {
		t.Errorf("control a=%q b=%q", control(t, a), control(t, b))
	}
}

func TestErrors(t *testing.T) {
	dir := service(t, true)
	if err := run(t, dir); err == nil {
		t.Errorf("no command should fail")
	}
	if err := run(t, "-u"); err == nil {
		t.Errorf("no directory should fail")
	}
	if err := run(t, "-u", service(t, false)); err == nil {
		t.Errorf("an unsupervised service should fail")
	}
}
