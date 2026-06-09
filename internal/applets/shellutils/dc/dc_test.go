package dc

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

func TestArithmetic(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"6 3 / p":     "2",
		"2k 7 3 / p":  "2.33",
		"2.5 2.5 * p": "6.25",
		"_5 3 + p":    "-2",
		"10 3 % p":    "1",
		"2 10 ^ p":    "1024",
		"7 2 - p":     "5",
		"4k 1 3 / p":  "0.3333",
		"_2 3 ^ p":    "-8",
	}
	for in, want := range cases {
		if got := run(t, in); got != want {
			t.Errorf("dc %q = %q, want %q", in, got, want)
		}
	}
}

func TestStackOps(t *testing.T) {
	t.Parallel()
	if got := run(t, "4 d * p"); got != "16" { // duplicate then multiply
		t.Errorf("d -> %q, want 16", got)
	}
	if got := run(t, "3 5 r - p"); got != "2" { // swap: 5 3 - = 2
		t.Errorf("r -> %q, want 2", got)
	}
	if got := run(t, "1 2 3 f"); got != "3\n2\n1" { // f prints top-down
		t.Errorf("f -> %q", got)
	}
	if got := run(t, "1 2 c 9 p"); got != "9" { // clear then push
		t.Errorf("c -> %q, want 9", got)
	}
}

func TestRegisters(t *testing.T) {
	t.Parallel()
	if got := run(t, "5 sa 3 la + p"); got != "8" { // store 5 in a, load it, add
		t.Errorf("registers -> %q, want 8", got)
	}
}

func TestExprFlag(t *testing.T) {
	t.Parallel()
	if got := run(t, "", "-e", "1 2 + p"); got != "3" {
		t.Errorf("-e -> %q, want 3", got)
	}
}

func TestPrintPopVsKeep(t *testing.T) {
	t.Parallel()
	// p keeps the value; a second p still prints it.
	if got := run(t, "7 p p"); got != "7\n7" {
		t.Errorf("p p -> %q", got)
	}
	// n pops the value; a following p sees an empty stack (no output).
	if got := run(t, "7 n"); got != "7" {
		t.Errorf("n -> %q", got)
	}
}
