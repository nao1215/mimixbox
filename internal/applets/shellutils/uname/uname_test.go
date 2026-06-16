package uname

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func withStub(t *testing.T) {
	t.Helper()
	orig := sysInfo
	sysInfo = func() (info, error) {
		return info{
			sysname:  "Linux",
			nodename: "host",
			release:  "6.6.0",
			version:  "#1 SMP",
			machine:  "x86_64",
			os:       "GNU/Linux",
		}, nil
	}
	t.Cleanup(func() { sysInfo = orig })
}

func run(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestDefaultIsSysname(t *testing.T) {
	withStub(t)
	out, _, err := run(t)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "Linux\n" {
		t.Errorf("out = %q", out)
	}
}

func TestAll(t *testing.T) {
	withStub(t)
	out, _, err := run(t, "-a")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "Linux host 6.6.0 #1 SMP x86_64 GNU/Linux\n" {
		t.Errorf("out = %q", out)
	}
}

func TestIndividualFlagsCombine(t *testing.T) {
	withStub(t)
	out, _, err := run(t, "-s", "-m")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "Linux x86_64\n" {
		t.Errorf("out = %q", out)
	}
}

func TestMachineAndOS(t *testing.T) {
	withStub(t)
	out, _, err := run(t, "-m")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "x86_64\n" {
		t.Errorf("out = %q", out)
	}
	out, _, err = run(t, "-o")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "GNU/Linux\n" {
		t.Errorf("out = %q", out)
	}
}

func TestRealUnameDoesNotError(t *testing.T) {
	t.Parallel()
	// Exercise the real uname(2)-backed path; the value is host-specific.
	in, err := uts()
	if err != nil {
		t.Fatalf("uts() error = %v", err)
	}
	if in.sysname == "" {
		t.Error("sysname is empty")
	}
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := New()
	if c.Name() != "uname" {
		t.Errorf("Name() = %q", c.Name())
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}

func TestHelpSections(t *testing.T) {
	t.Parallel()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, []string{"--help"}); err != nil {
		t.Fatalf("Run --help error = %v", err)
	}
	if !strings.Contains(out.String(), "Examples:") || !strings.Contains(out.String(), "Exit status:") {
		t.Errorf("--help missing structured sections:\n%s", out.String())
	}
}
