package truncate_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/fileutils/truncate"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := truncate.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func size(t *testing.T, p string) int64 {
	t.Helper()
	info, err := os.Stat(p)
	if err != nil {
		t.Fatal(err)
	}
	return info.Size()
}

func TestCreateAndExtend(t *testing.T) {
	t.Parallel()
	p := filepath.Join(t.TempDir(), "f")
	if _, _, err := run(t, "-s", "10", p); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if size(t, p) != 10 {
		t.Errorf("size = %d, want 10", size(t, p))
	}
}

func TestShrink(t *testing.T) {
	t.Parallel()
	p := filepath.Join(t.TempDir(), "f")
	if err := os.WriteFile(p, []byte("0123456789"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := run(t, "-s", "4", p); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if size(t, p) != 4 {
		t.Errorf("size = %d, want 4", size(t, p))
	}
}

func TestRelativeGrow(t *testing.T) {
	t.Parallel()
	p := filepath.Join(t.TempDir(), "f")
	if err := os.WriteFile(p, []byte("abc"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := run(t, "-s", "+5", p); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if size(t, p) != 8 {
		t.Errorf("size = %d, want 8", size(t, p))
	}
}

func TestRelativeShrinkClampsToZero(t *testing.T) {
	t.Parallel()
	p := filepath.Join(t.TempDir(), "f")
	if err := os.WriteFile(p, []byte("abc"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := run(t, "-s", "-10", p); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if size(t, p) != 0 {
		t.Errorf("size = %d, want 0", size(t, p))
	}
}

func TestSuffix(t *testing.T) {
	t.Parallel()
	p := filepath.Join(t.TempDir(), "f")
	if _, _, err := run(t, "-s", "1K", p); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if size(t, p) != 1024 {
		t.Errorf("size = %d, want 1024", size(t, p))
	}
}

func TestNoCreate(t *testing.T) {
	t.Parallel()
	p := filepath.Join(t.TempDir(), "absent")
	if _, _, err := run(t, "-c", "-s", "5", p); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if _, err := os.Stat(p); !os.IsNotExist(err) {
		t.Error("file should not have been created")
	}
}

func TestMissingSize(t *testing.T) {
	t.Parallel()
	_, _, err := run(t, "file")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "specify a size") {
		t.Errorf("err = %v", err)
	}
}

func TestInvalidSize(t *testing.T) {
	t.Parallel()
	p := filepath.Join(t.TempDir(), "f")
	_, errOut, err := run(t, "-s", "xyz", p)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(errOut, "invalid number") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := truncate.New()
	if c.Name() != "truncate" {
		t.Errorf("Name() = %q", c.Name())
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}
