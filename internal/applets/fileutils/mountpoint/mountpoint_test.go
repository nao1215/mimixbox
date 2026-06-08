package mountpoint_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/fileutils/mountpoint"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := mountpoint.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestRootIsMountpoint(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "/")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(out, "/ is a mountpoint") {
		t.Errorf("out = %q", out)
	}
}

func TestRegularDirIsNotMountpoint(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	out, _, err := run(t, dir)
	if err == nil {
		t.Fatal("expected non-zero exit for a non-mountpoint")
	}
	if !strings.Contains(out, "is not a mountpoint") {
		t.Errorf("out = %q", out)
	}
}

func TestQuiet(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "-q", "/")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "" {
		t.Errorf("quiet output should be empty, got %q", out)
	}
}

func TestQuietNonMountpoint(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	out, _, err := run(t, "-q", dir)
	if err == nil {
		t.Fatal("expected non-zero exit")
	}
	if out != "" {
		t.Errorf("quiet output should be empty, got %q", out)
	}
}

func TestMissingDir(t *testing.T) {
	t.Parallel()
	_, _, err := run(t, "/no/such/dir")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestWrongArgCount(t *testing.T) {
	t.Parallel()
	_, _, err := run(t)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "exactly one argument") {
		t.Errorf("err = %v", err)
	}
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := mountpoint.New()
	if c.Name() != "mountpoint" {
		t.Errorf("Name() = %q", c.Name())
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}
