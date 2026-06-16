package printenv_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/shellutils/printenv"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := printenv.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestRunNamedVariable(t *testing.T) {
	t.Setenv("MIMIXBOX_PRINTENV_TEST", "hello")

	out, _, err := run(t, "MIMIXBOX_PRINTENV_TEST")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "hello\n" {
		t.Errorf("out = %q, want %q", out, "hello\n")
	}
}

func TestRunUnsetVariable(t *testing.T) {
	out, _, err := run(t, "MIMIXBOX_PRINTENV_DEFINITELY_UNSET")
	if err == nil {
		t.Fatal("expected non-nil error for unset variable")
	}
	if out != "" {
		t.Errorf("out = %q, want empty", out)
	}
}

func TestRunMixedSetAndUnset(t *testing.T) {
	t.Setenv("MIMIXBOX_PRINTENV_SET", "value")

	out, _, err := run(t, "MIMIXBOX_PRINTENV_SET", "MIMIXBOX_PRINTENV_UNSET")
	if err == nil {
		t.Fatal("expected non-nil error when any variable is unset")
	}
	if out != "value\n" {
		t.Errorf("out = %q, want %q", out, "value\n")
	}
}

func TestRunNull(t *testing.T) {
	t.Setenv("MIMIXBOX_PRINTENV_TEST", "world")

	out, _, err := run(t, "-0", "MIMIXBOX_PRINTENV_TEST")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "world\x00" {
		t.Errorf("out = %q, want %q", out, "world\x00")
	}
}

func TestRunAllContainsSetVariable(t *testing.T) {
	t.Setenv("MIMIXBOX_PRINTENV_ALL", "present")

	out, _, err := run(t)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(out, "MIMIXBOX_PRINTENV_ALL=present\n") {
		t.Errorf("out did not contain expected NAME=VALUE line: %q", out)
	}
}

func TestHelpSections(t *testing.T) {
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := printenv.New().Run(context.Background(), io, []string{"--help"}); err != nil {
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
