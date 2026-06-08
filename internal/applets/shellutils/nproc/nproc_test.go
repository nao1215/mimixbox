package nproc

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestPrintsCount(t *testing.T) {
	orig := cpuCount
	cpuCount = func() int { return 8 }
	t.Cleanup(func() { cpuCount = orig })

	out, _, err := run(t)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "8\n" {
		t.Errorf("out = %q", out)
	}
}

func TestIgnore(t *testing.T) {
	orig := cpuCount
	cpuCount = func() int { return 8 }
	t.Cleanup(func() { cpuCount = orig })

	out, _, err := run(t, "--ignore", "3")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "5\n" {
		t.Errorf("out = %q", out)
	}
}

func TestIgnoreClampsToOne(t *testing.T) {
	orig := cpuCount
	cpuCount = func() int { return 2 }
	t.Cleanup(func() { cpuCount = orig })

	out, _, err := run(t, "--ignore", "10")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "1\n" {
		t.Errorf("out = %q", out)
	}
}

func TestRealCount(t *testing.T) {
	t.Parallel()
	out, _, err := run(t)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if strings.TrimSpace(out) == "" {
		t.Error("output is empty")
	}
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := New()
	if c.Name() != "nproc" {
		t.Errorf("Name() = %q", c.Name())
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}
