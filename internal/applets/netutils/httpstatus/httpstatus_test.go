package httpstatus

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

func TestSearchSingle(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "search", "404")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.HasPrefix(out, "404 Not Found (ref.=RFC9110") {
		t.Errorf("out = %q", out)
	}
}

func TestSearchMultiple(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "search", "200", "500")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(out, "200 OK") || !strings.Contains(out, "500 Internal Server Error") {
		t.Errorf("out = %q", out)
	}
}

func TestSearchUnknownCode(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, "search", "799")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(errOut, "unknown status code 799") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestSearchInvalidCode(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, "search", "abc")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(errOut, "invalid code") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestList(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "list")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != len(table) {
		t.Errorf("listed %d lines, want %d", len(lines), len(table))
	}
	// list is sorted ascending: the first line is the smallest code (100).
	if !strings.HasPrefix(lines[0], "100 ") {
		t.Errorf("first line = %q, want 100 first", lines[0])
	}
}

func TestNoSubcommand(t *testing.T) {
	t.Parallel()
	_, _, err := run(t)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "expected a subcommand") {
		t.Errorf("err = %v", err)
	}
}

func TestUnknownSubcommand(t *testing.T) {
	t.Parallel()
	_, _, err := run(t, "frobnicate")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "unknown subcommand") {
		t.Errorf("err = %v", err)
	}
}

func TestSearchNoCodes(t *testing.T) {
	t.Parallel()
	_, _, err := run(t, "search")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "at least one CODE") {
		t.Errorf("err = %v", err)
	}
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := New()
	if c.Name() != "http-status-code" {
		t.Errorf("Name() = %q", c.Name())
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}
