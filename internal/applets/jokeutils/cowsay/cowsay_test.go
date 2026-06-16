package cowsay_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/jokeutils/cowsay"
	"github.com/nao1215/mimixbox/internal/command"
)

const wantHi = "------------------------------------------------------------\n" +
	"hi\n" +
	"------------------------------------------------------------\n" +
	"   \\ \n" +
	"    \\   ^__^\n" +
	"     \\  (oo)\\_______\n" +
	"        (__)\\       )\\/\\\n" +
	"            ||----w |\n" +
	"            ||     ||\n"

func run(t *testing.T, in string, args ...string) (string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(in), Out: out, Err: &bytes.Buffer{}}
	err := cowsay.New().Run(context.Background(), io, args)
	return out.String(), err
}

func TestRunFromOperand(t *testing.T) {
	t.Parallel()
	out, err := run(t, "", "hi")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != wantHi {
		t.Errorf("out = %q, want %q", out, wantHi)
	}
}

func TestRunFromStdin(t *testing.T) {
	t.Parallel()
	out, err := run(t, "hi\n")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != wantHi {
		t.Errorf("out = %q, want %q", out, wantHi)
	}
}

func TestRunContainsMessageAndCow(t *testing.T) {
	t.Parallel()
	out, err := run(t, "", "moo", "world")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(out, "moo world") {
		t.Errorf("output missing message; got %q", out)
	}
	if !strings.Contains(out, "^__^") || !strings.Contains(out, "(oo)") {
		t.Errorf("output missing cow art; got %q", out)
	}
}

func TestMeta(t *testing.T) {
	t.Parallel()
	c := cowsay.New()
	if c.Name() != "cowsay" {
		t.Errorf("Name() = %q, want %q", c.Name(), "cowsay")
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}

// TestHelpSections verifies that --help renders both the Examples and the
// Exit status sections supplied through WithHelp.
func TestHelpSections(t *testing.T) {
	out, err := run(t, "", "--help")
	if err != nil {
		t.Fatalf("--help err = %v", err)
	}
	for _, want := range []string{"Examples:", "Exit status:"} {
		if !strings.Contains(out, want) {
			t.Errorf("--help missing %q section:\n%s", want, out)
		}
	}
}
