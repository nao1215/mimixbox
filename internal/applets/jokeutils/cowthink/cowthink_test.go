package cowthink_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/jokeutils/cowthink"
	"github.com/nao1215/mimixbox/internal/command"
)

const wantHi = "------------------------------------------------------------\n" +
	"hi\n" +
	"------------------------------------------------------------\n" +
	"   o\n" +
	"    o   ^__^\n" +
	"        (oo)\\_______\n" +
	"        (__)\\       )\\/\\\n" +
	"            ||----w |\n" +
	"            ||     ||\n"

func run(t *testing.T, in string, args ...string) (string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(in), Out: out, Err: &bytes.Buffer{}}
	err := cowthink.New().Run(context.Background(), io, args)
	return out.String(), err
}

func TestFromOperand(t *testing.T) {
	t.Parallel()
	out, err := run(t, "", "hi")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != wantHi {
		t.Errorf("out = %q, want %q", out, wantHi)
	}
}

func TestFromStdin(t *testing.T) {
	t.Parallel()
	out, err := run(t, "hi")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != wantHi {
		t.Errorf("out = %q", out)
	}
}

func TestThoughtBubbleUsesO(t *testing.T) {
	t.Parallel()
	out, err := run(t, "", "thinking")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(out, "   o\n") {
		t.Errorf("expected an 'o' thought connector in %q", out)
	}
}

func TestLongMessageWraps(t *testing.T) {
	t.Parallel()
	msg := strings.Repeat("x", 130)
	out, err := run(t, "", msg)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(out, strings.Repeat("x", 60)+"\n") {
		t.Errorf("message was not wrapped at 60 columns: %q", out)
	}
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := cowthink.New()
	if c.Name() != "cowthink" {
		t.Errorf("Name() = %q", c.Name())
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
