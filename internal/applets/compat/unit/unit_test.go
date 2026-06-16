package unit

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func TestUnitExitsNonZero(t *testing.T) {
	t.Parallel()
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: errBuf}
	err := New().Run(context.Background(), io, nil)
	var ee *command.ExitError
	if !errors.As(err, &ee) || ee.Code != 2 {
		t.Errorf("err = %v, want ExitError code 2", err)
	}
	if !strings.Contains(errBuf.String(), "go test") {
		t.Errorf("stderr should point at the real tests, got %q", errBuf.String())
	}
}

func TestUnitHelp(t *testing.T) {
	t.Parallel()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, []string{"--help"}); err != nil {
		t.Fatalf("--help err = %v", err)
	}
	for _, want := range []string{"Usage: unit", "Examples:", "Exit status:", "Notes:"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("--help missing %q; out = %q", want, out.String())
		}
	}
}
