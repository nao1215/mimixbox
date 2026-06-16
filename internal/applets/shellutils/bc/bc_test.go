package bc

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, in string, args ...string) string {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(in), Out: out, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, args); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	return strings.TrimRight(out.String(), "\n")
}

func TestExpressions(t *testing.T) {
	t.Parallel()
	// All values verified against GNU bc.
	cases := map[string]string{
		"2 + 3 * 4":     "14",
		"(2 + 3) * 4":   "20",
		"2^10":          "1024",
		"10 % 3":        "1",
		"-3 + 1":        "-2",
		"2.5 * 2.5":     "6.2",
		"scale=2; 7/3":  "2.33",
		"scale=4; 1/3":  ".3333",
		"scale=2; -1/4": "-.25",
		"3^3":           "27",
		"7 - 2 - 1":     "4",
	}
	for in, want := range cases {
		if got := run(t, in); got != want {
			t.Errorf("bc %q = %q, want %q", in, got, want)
		}
	}
}

func TestVariables(t *testing.T) {
	t.Parallel()
	if got := run(t, "x = 5; x * x"); got != "25" {
		t.Errorf("variables -> %q, want 25", got)
	}
	if got := run(t, "a=2; b=3; a^b"); got != "8" {
		t.Errorf("two vars -> %q, want 8", got)
	}
	// An unset variable is zero.
	if got := run(t, "y + 7"); got != "7" {
		t.Errorf("unset var -> %q, want 7", got)
	}
}

func TestAssignmentDoesNotPrint(t *testing.T) {
	t.Parallel()
	// Only the bare expression prints; the assignment does not.
	if got := run(t, "x = 9"); got != "" {
		t.Errorf("assignment printed %q, want nothing", got)
	}
}

func TestPrecisionPersists(t *testing.T) {
	t.Parallel()
	// scale set on one line applies to a later line (same machine per run call).
	if got := run(t, "scale=3\n10/7"); got != "1.428" {
		t.Errorf("scale across lines -> %q, want 1.428", got)
	}
}

func TestDivideByZero(t *testing.T) {
	t.Parallel()
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader("1/0\n"), Out: &bytes.Buffer{}, Err: errBuf}
	_ = New().Run(context.Background(), io, nil)
	if !strings.Contains(errBuf.String(), "divide by zero") {
		t.Errorf("expected a divide-by-zero error, got %q", errBuf.String())
	}
}

func TestHelpExitStatus(t *testing.T) {
	t.Parallel()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, []string{"--help"}); err != nil {
		t.Fatalf("--help err = %v", err)
	}
	if !strings.Contains(out.String(), "Exit status:") {
		t.Errorf("--help missing Exit status section = %q", out.String())
	}
}
