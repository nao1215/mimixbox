package fmt_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/textutils/fmt"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, stdin string, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(stdin), Out: out, Err: errBuf}
	err := fmt.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestReflow(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "aa bb cc dd\n", "-w", "5")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "aa bb\ncc dd\n" {
		t.Errorf("out = %q", out)
	}
}

func TestJoinShortLines(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "aa\nbb\ncc\n", "-w", "10")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "aa bb cc\n" {
		t.Errorf("out = %q", out)
	}
}

func TestParagraphSeparation(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "aa bb\n\ncc dd\n", "-w", "20")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "aa bb\n\ncc dd\n" {
		t.Errorf("out = %q", out)
	}
}

func TestInvalidWidth(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, "x\n", "-w", "0")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(errOut, "invalid width") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestMissingFile(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, "", "/no/such/file")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(errOut, "fmt: /no/such/file:") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := fmt.New()
	if c.Name() != "fmt" {
		t.Errorf("Name() = %q", c.Name())
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}

func TestHelpSections(t *testing.T) {
	out, _, err := run(t, "", "--help")
	if err != nil {
		t.Fatalf("--help error = %v", err)
	}
	if !strings.Contains(out, "Examples:") || !strings.Contains(out, "Exit status:") {
		t.Errorf("--help missing structured sections:\n%s", out)
	}
}
