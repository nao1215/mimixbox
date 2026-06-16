package posixer

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

func TestHeaderAndRows(t *testing.T) {
	orig := lookPath
	lookPath = func(name string) (string, error) {
		if name == "cat" {
			return "/bin/cat", nil
		}
		return "", errors.New("not found")
	}
	t.Cleanup(func() { lookPath = orig })

	out, _, err := run(t)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(out, "NAME") || !strings.Contains(out, "INSTALLED") || !strings.Contains(out, "PATH") {
		t.Errorf("missing header in %q", out)
	}
	if !strings.Contains(out, "/bin/cat") {
		t.Errorf("expected cat resolved to /bin/cat in %q", out)
	}
}

func TestInstalledColumn(t *testing.T) {
	orig := lookPath
	lookPath = func(string) (string, error) { return "", errors.New("nope") }
	t.Cleanup(func() { lookPath = orig })

	out, _, err := run(t)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	// Every utility is missing, so no row should report "yes".
	if strings.Contains(out, "yes") {
		t.Errorf("expected all rows to be missing, got %q", out)
	}
}

func TestCheckSubcommand(t *testing.T) {
	orig := lookPath
	lookPath = func(string) (string, error) { return "/bin/x", nil }
	t.Cleanup(func() { lookPath = orig })

	out, _, err := run(t, "check")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(out, "yes") {
		t.Errorf("expected installed rows in %q", out)
	}
}

func TestUnknownSubcommand(t *testing.T) {
	t.Parallel()
	_, _, err := run(t, "bogus")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "unknown subcommand") {
		t.Errorf("err = %v", err)
	}
}

func TestUtilityListNonEmpty(t *testing.T) {
	t.Parallel()
	if len(posixUtilities) == 0 {
		t.Fatal("expected a non-empty POSIX utility list")
	}
	for _, u := range posixUtilities {
		if u.kind != "required" && u.kind != "optional" {
			t.Errorf("%s has invalid kind %q", u.name, u.kind)
		}
	}
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := New()
	if c.Name() != "posixer" {
		t.Errorf("Name() = %q", c.Name())
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}

func TestHelpSections(t *testing.T) {
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, []string{"--help"}); err != nil {
		t.Fatalf("Run --help error = %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "Examples:") {
		t.Errorf("--help missing Examples section:\n%s", got)
	}
	if !strings.Contains(got, "Exit status:") {
		t.Errorf("--help missing Exit status section:\n%s", got)
	}
}
