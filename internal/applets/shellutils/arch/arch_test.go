package arch

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

func TestPrintsMachine(t *testing.T) {
	orig := machine
	machine = func() (string, error) { return "x86_64", nil }
	t.Cleanup(func() { machine = orig })

	out, _, err := run(t)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "x86_64\n" {
		t.Errorf("out = %q", out)
	}
}

func TestError(t *testing.T) {
	orig := machine
	machine = func() (string, error) { return "", errors.New("boom") }
	t.Cleanup(func() { machine = orig })

	_, _, err := run(t)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRealMachine(t *testing.T) {
	t.Parallel()
	got, err := realMachine()
	if err != nil {
		t.Fatalf("realMachine() error = %v", err)
	}
	if got == "" {
		t.Error("machine name is empty")
	}
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := New()
	if c.Name() != "arch" {
		t.Errorf("Name() = %q", c.Name())
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}
