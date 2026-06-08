package hostname

import (
	"bytes"
	"context"
	"errors"
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

func TestPrintsHostname(t *testing.T) {
	orig := hostFn
	hostFn = func() (string, error) { return "host.example.com", nil }
	t.Cleanup(func() { hostFn = orig })

	out, _, err := run(t)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "host.example.com\n" {
		t.Errorf("out = %q", out)
	}
}

func TestShort(t *testing.T) {
	orig := hostFn
	hostFn = func() (string, error) { return "host.example.com", nil }
	t.Cleanup(func() { hostFn = orig })

	out, _, err := run(t, "-s")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "host\n" {
		t.Errorf("out = %q", out)
	}
}

func TestError(t *testing.T) {
	orig := hostFn
	hostFn = func() (string, error) { return "", errors.New("boom") }
	t.Cleanup(func() { hostFn = orig })

	_, _, err := run(t)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRealHostname(t *testing.T) {
	t.Parallel()
	out, _, err := run(t)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if strings.TrimSpace(out) == "" {
		t.Error("hostname is empty")
	}
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := New()
	if c.Name() != "hostname" {
		t.Errorf("Name() = %q", c.Name())
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}
