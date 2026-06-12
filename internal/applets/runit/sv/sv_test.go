package sv

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

// service creates a service dir; supervised adds supervise/ok and control.
func service(t *testing.T, supervised bool, pid string, down bool) string {
	t.Helper()
	dir := t.TempDir()
	if supervised {
		sup := filepath.Join(dir, "supervise")
		if err := os.MkdirAll(sup, 0o755); err != nil {
			t.Fatal(err)
		}
		_ = os.WriteFile(filepath.Join(sup, "ok"), nil, 0o644)
		_ = os.WriteFile(filepath.Join(sup, "control"), nil, 0o644)
		if pid != "" {
			_ = os.WriteFile(filepath.Join(sup, "pid"), []byte(pid+"\n"), 0o644)
		}
	}
	if down {
		_ = os.WriteFile(filepath.Join(dir, "down"), nil, 0o644)
	}
	return dir
}

func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	err := New().Run(context.Background(), io, args)
	return out.String(), err
}

func TestControlWritesChar(t *testing.T) {
	dir := service(t, true, "", false)
	for cmd, want := range map[string]string{"up": "u", "down": "d", "restart": "t", "kill": "k"} {
		if _, err := run(t, cmd, dir); err != nil {
			t.Fatalf("%s: %v", cmd, err)
		}
		got, _ := os.ReadFile(filepath.Join(dir, "supervise", "control"))
		if string(got) != want {
			t.Errorf("sv %s wrote %q, want %q", cmd, got, want)
		}
		_ = os.WriteFile(filepath.Join(dir, "supervise", "control"), nil, 0o644) // reset
	}
}

func TestStatusRunning(t *testing.T) {
	dir := service(t, true, "4242", false)
	out, err := run(t, "status", dir)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "run") || !strings.Contains(out, "pid 4242") {
		t.Errorf("status = %q", out)
	}
}

func TestStatusDown(t *testing.T) {
	dir := service(t, true, "", true)
	out, err := run(t, "status", dir)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "down") {
		t.Errorf("status = %q, want down", out)
	}
}

func TestStatusNotSupervised(t *testing.T) {
	dir := service(t, false, "", false)
	out, err := run(t, "status", dir)
	if err == nil {
		t.Errorf("an unsupervised service should fail")
	}
	if !strings.Contains(out, "not supervised") {
		t.Errorf("status = %q", out)
	}
}

func TestErrors(t *testing.T) {
	dir := service(t, true, "", false)
	if _, err := run(t, "bogus", dir); err == nil {
		t.Errorf("an unknown command should fail")
	}
	if _, err := run(t, "up", service(t, false, "", false)); err == nil {
		t.Errorf("controlling an unsupervised service should fail")
	}
	if _, err := run(t, "up"); err == nil {
		t.Errorf("missing directory should fail")
	}
}
