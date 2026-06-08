package fortune

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

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

func TestPrintsAKnownFortune(t *testing.T) {
	t.Parallel()
	out, _, err := run(t)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	got := strings.TrimRight(out, "\n")
	if !contains(fortunes, got) {
		t.Errorf("printed unknown fortune %q", got)
	}
}

func TestShortOnly(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "-s")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	got := strings.TrimRight(out, "\n")
	if len(got) > shortLimit {
		t.Errorf("short fortune too long (%d): %q", len(got), got)
	}
}

func TestCandidatesShortNonEmpty(t *testing.T) {
	t.Parallel()
	if len(candidates(true)) == 0 {
		t.Fatal("expected at least one short fortune")
	}
	for _, f := range candidates(true) {
		if len(f) > shortLimit {
			t.Errorf("candidate %q exceeds short limit", f)
		}
	}
}

func TestCandidatesAll(t *testing.T) {
	t.Parallel()
	if len(candidates(false)) != len(fortunes) {
		t.Errorf("candidates(false) = %d, want %d", len(candidates(false)), len(fortunes))
	}
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := New()
	if c.Name() != "fortune" {
		t.Errorf("Name() = %q", c.Name())
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}
