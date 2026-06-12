package ttysize

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func stub(t *testing.T, w, h int) {
	t.Helper()
	orig := getSizeFn
	getSizeFn = func() (int, int) { return w, h }
	t.Cleanup(func() { getSizeFn = orig })
}

func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	err := New().Run(context.Background(), io, args)
	return strings.TrimSpace(out.String()), err
}

func TestPrintsBoth(t *testing.T) {
	stub(t, 100, 40)
	out, err := run(t)
	if err != nil {
		t.Fatal(err)
	}
	if out != "100 40" {
		t.Errorf("ttysize = %q, want \"100 40\"", out)
	}
}

func TestSelective(t *testing.T) {
	stub(t, 100, 40)
	if out, _ := run(t, "w"); out != "100" {
		t.Errorf("ttysize w = %q", out)
	}
	if out, _ := run(t, "h"); out != "40" {
		t.Errorf("ttysize h = %q", out)
	}
	if out, _ := run(t, "h", "w"); out != "40 100" {
		t.Errorf("ttysize h w = %q, want \"40 100\"", out)
	}
}

func TestUnknownArg(t *testing.T) {
	stub(t, 80, 24)
	if _, err := run(t, "x"); err == nil {
		t.Errorf("an unknown argument should fail")
	}
}
