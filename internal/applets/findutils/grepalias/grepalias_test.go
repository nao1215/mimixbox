package grepalias

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, c *Command, in string, args ...string) string {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(in), Out: out, Err: &bytes.Buffer{}}
	_ = c.Run(context.Background(), io, args)
	return out.String()
}

func TestEgrepUsesExtendedRegex(t *testing.T) {
	t.Parallel()
	got := run(t, NewEgrep(), "foo\nbar\nbaz\n", "ba(r|z)")
	if got != "bar\nbaz\n" {
		t.Errorf("egrep = %q, want bar/baz", got)
	}
}

func TestFgrepIsFixedString(t *testing.T) {
	t.Parallel()
	got := run(t, NewFgrep(), "a.b\naxb\n", "a.b")
	if got != "a.b\n" {
		t.Errorf("fgrep = %q, want only the literal a.b line", got)
	}
}

// TestHelpSections asserts egrep/fgrep --help render alias-named structured help.
func TestHelpSections(t *testing.T) {
	t.Parallel()
	for _, c := range []*Command{NewEgrep(), NewFgrep()} {
		out := &bytes.Buffer{}
		io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
		if err := c.Run(context.Background(), io, []string{"--help"}); err != nil {
			t.Fatalf("%s --help err = %v", c.Name(), err)
		}
		for _, want := range []string{"Usage: " + c.Name(), "Examples:", "Exit status:"} {
			if !strings.Contains(out.String(), want) {
				t.Errorf("%s --help missing %q: %q", c.Name(), want, out.String())
			}
		}
	}
}
