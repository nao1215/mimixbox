package fgconsole

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func stub(t *testing.T, active int, err error) {
	t.Helper()
	orig := getActiveFn
	getActiveFn = func() (int, error) { return active, err }
	t.Cleanup(func() { getActiveFn = orig })
}

func run(t *testing.T) (string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	err := New().Run(context.Background(), io, nil)
	return strings.TrimSpace(out.String()), err
}

func TestPrintsActiveVT(t *testing.T) {
	stub(t, 3, nil)
	out, err := run(t)
	if err != nil {
		t.Fatal(err)
	}
	if out != "3" {
		t.Errorf("fgconsole = %q, want 3", out)
	}
}

func TestFailure(t *testing.T) {
	stub(t, 0, errors.New("inappropriate ioctl"))
	if _, err := run(t); err == nil {
		t.Errorf("a console read failure should fail")
	}
}
