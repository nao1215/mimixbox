package setarch

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
	"github.com/nao1215/mimixbox/internal/testutil/fakecmd"
)

// withStub captures the personality passed to setPersonality without touching
// the test process's execution domain.
func withStub(t *testing.T) *uintptr {
	t.Helper()
	var got uintptr = 0xffff
	orig := setPersonality
	setPersonality = func(p uintptr) error { got = p; return nil }
	t.Cleanup(func() { setPersonality = orig })
	return &got
}

func run(t *testing.T, c *Command, args ...string) (string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	err := c.Run(context.Background(), io, args)
	return out.String(), err
}

// requireEcho installs a repo-local fake echo and points PATH at it so the
// test exercises the exec path deterministically without relying on a host
// /bin/echo being present.
func requireEcho(t *testing.T) {
	t.Helper()
	fakecmd.UseOnly(t, "echo")
}

func TestLinux32Personality(t *testing.T) {
	requireEcho(t)
	got := withStub(t)
	out, err := run(t, NewLinux32(), "echo", "ok")
	if err != nil {
		t.Fatal(err)
	}
	if *got != perLinux32 {
		t.Errorf("persona = %#x, want perLinux32", *got)
	}
	if !strings.Contains(out, "ok") {
		t.Errorf("command output = %q", out)
	}
}

func TestLinux64Personality(t *testing.T) {
	requireEcho(t)
	got := withStub(t)
	if _, err := run(t, NewLinux64(), "echo", "x"); err != nil {
		t.Fatal(err)
	}
	if *got != perLinux {
		t.Errorf("persona = %#x, want perLinux", *got)
	}
}

func TestSetarchArch(t *testing.T) {
	requireEcho(t)
	got := withStub(t)
	if _, err := run(t, NewSetarch(), "i686", "echo", "x"); err != nil {
		t.Fatal(err)
	}
	if *got != perLinux32 {
		t.Errorf("setarch i686 persona = %#x, want perLinux32", *got)
	}
	if _, err := run(t, NewSetarch(), "x86_64", "echo", "x"); err != nil {
		t.Fatal(err)
	}
	if *got != perLinux {
		t.Errorf("setarch x86_64 persona = %#x, want perLinux", *got)
	}
}

func TestMissingCommand(t *testing.T) {
	withStub(t)
	if _, err := run(t, NewLinux32()); err == nil {
		t.Errorf("missing command should fail")
	}
}

func TestSetarchMissingArch(t *testing.T) {
	withStub(t)
	if _, err := run(t, NewSetarch()); err == nil {
		t.Errorf("setarch with no arch should fail")
	}
}

func TestPersonaForArch(t *testing.T) {
	t.Parallel()
	for _, a := range []string{"i686", "i386", "linux32", "x86"} {
		if personaForArch(a) != perLinux32 {
			t.Errorf("personaForArch(%q) != perLinux32", a)
		}
	}
	for _, a := range []string{"x86_64", "linux64", "amd64"} {
		if personaForArch(a) != perLinux {
			t.Errorf("personaForArch(%q) != perLinux", a)
		}
	}
}
