package shuf_test

import (
	"bytes"
	"context"
	"sort"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/textutils/shuf"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, stdin string, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(stdin), Out: out, Err: errBuf}
	err := shuf.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func sortedLines(s string) []string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	sort.Strings(lines)
	return lines
}

func TestPermutation(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "a\nb\nc\nd\ne\n")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	got := sortedLines(out)
	want := []string{"a", "b", "c", "d", "e"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("multiset = %v, want %v", got, want)
	}
}

func TestHeadCount(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "a\nb\nc\nd\ne\n", "-n", "2")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 2 {
		t.Errorf("got %d lines, want 2", len(lines))
	}
}

func TestEcho(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "", "-e", "x", "y", "z")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	got := sortedLines(out)
	if strings.Join(got, ",") != "x,y,z" {
		t.Errorf("multiset = %v", got)
	}
}

func TestInputRange(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "", "-i", "1-4")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	got := sortedLines(out)
	if strings.Join(got, ",") != "1,2,3,4" {
		t.Errorf("multiset = %v", got)
	}
}

func TestInvalidRange(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, "", "-i", "5-1")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(errOut, "invalid input range") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestMissingFile(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, "", "/no/such/file")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(errOut, "shuf: /no/such/file:") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := shuf.New()
	if c.Name() != "shuf" {
		t.Errorf("Name() = %q", c.Name())
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}
