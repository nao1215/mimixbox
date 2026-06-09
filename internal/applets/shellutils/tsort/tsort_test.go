package tsort

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, in string, args ...string) (string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(in), Out: out, Err: &bytes.Buffer{}}
	err := New().Run(context.Background(), io, args)
	return out.String(), err
}

// indexOf returns the line index of s in the newline-separated output.
func indexOf(out, s string) int {
	for i, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if line == s {
			return i
		}
	}
	return -1
}

func TestTopologicalOrder(t *testing.T) {
	t.Parallel()
	out, err := run(t, "a b\nb c\nd c\n")
	if err != nil {
		t.Fatalf("tsort error = %v", err)
	}
	// Constraints: a before b, b before c, d before c.
	if indexOf(out, "a") >= indexOf(out, "b") || indexOf(out, "b") >= indexOf(out, "c") || indexOf(out, "d") >= indexOf(out, "c") {
		t.Errorf("order violates constraints:\n%s", out)
	}
}

func TestDeterministic(t *testing.T) {
	t.Parallel()
	a, _ := run(t, "x y\nz y\n")
	b, _ := run(t, "x y\nz y\n")
	if a != b {
		t.Errorf("tsort not deterministic: %q vs %q", a, b)
	}
}

func TestCycle(t *testing.T) {
	t.Parallel()
	out, err := run(t, "a b\nb a\n")
	if err == nil {
		t.Errorf("a cycle should fail")
	}
	// Output still lists the nodes.
	if !strings.Contains(out, "a") || !strings.Contains(out, "b") {
		t.Errorf("cycle output should still list nodes, got %q", out)
	}
}

func TestOddTokens(t *testing.T) {
	t.Parallel()
	if _, err := run(t, "a b c\n"); err == nil {
		t.Errorf("an odd number of tokens should fail")
	}
}
