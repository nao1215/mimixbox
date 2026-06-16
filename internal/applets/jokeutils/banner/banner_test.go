package banner

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

func TestRenderHasFiveRows(t *testing.T) {
	t.Parallel()
	out := Render("HI")
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != glyphHeight {
		t.Fatalf("got %d rows, want %d", len(lines), glyphHeight)
	}
}

func TestRenderLetterArt(t *testing.T) {
	t.Parallel()
	out := Render("A")
	if !strings.Contains(out, "#####") {
		t.Errorf("letter A should contain a full row, got %q", out)
	}
	if strings.Contains(out, "A") {
		t.Errorf("art should not contain the literal letter: %q", out)
	}
}

func TestUppercased(t *testing.T) {
	t.Parallel()
	if Render("a") != Render("A") {
		t.Error("lowercase input should render the same as uppercase")
	}
}

func TestUnknownRuneIsBlank(t *testing.T) {
	t.Parallel()
	out := Render("~")
	if strings.TrimSpace(out) != "" {
		t.Errorf("unknown rune should render blank, got %q", out)
	}
}

func TestDefined(t *testing.T) {
	t.Parallel()
	if !defined('a') || !defined('Z') || !defined('5') {
		t.Error("expected letters and digits to be defined")
	}
	if defined('~') {
		t.Error("~ should not be defined")
	}
}

func TestRunPrints(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "HI")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(out, "#") {
		t.Errorf("expected art in output, got %q", out)
	}
}

func TestMissingOperand(t *testing.T) {
	t.Parallel()
	_, _, err := run(t)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "missing message operand") {
		t.Errorf("err = %v", err)
	}
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := New()
	if c.Name() != "banner" {
		t.Errorf("Name() = %q", c.Name())
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}

func TestHelpSections(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "--help")
	if err != nil {
		t.Fatalf("--help error = %v", err)
	}
	if !strings.Contains(out, "Examples:") {
		t.Errorf("--help missing Examples: %q", out)
	}
	if !strings.Contains(out, "Exit status:") {
		t.Errorf("--help missing Exit status: %q", out)
	}
}
