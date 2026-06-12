package svok

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, args ...string) error {
	t.Helper()
	io := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	return New().Run(context.Background(), io, args)
}

func TestSupervised(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "supervise"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "supervise", "ok"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := run(t, dir); err != nil {
		t.Errorf("a supervised service should succeed, got %v", err)
	}
}

func TestNotSupervised(t *testing.T) {
	err := run(t, t.TempDir())
	ee, ok := err.(*command.ExitError)
	if !ok || ee.Code != 100 {
		t.Errorf("err = %v, want exit 100", err)
	}
}

func TestNoDir(t *testing.T) {
	if err := run(t); err == nil {
		t.Errorf("a missing directory should fail")
	}
}
