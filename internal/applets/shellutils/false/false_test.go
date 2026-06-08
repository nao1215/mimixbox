package boolfalse_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	boolfalse "github.com/nao1215/mimixbox/internal/applets/shellutils/false"
	"github.com/nao1215/mimixbox/internal/command"
)

func TestRunAlwaysFails(t *testing.T) {
	t.Parallel()
	io := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	code := command.Execute(context.Background(), boolfalse.New(), io, []string{"ignored"})
	if code != command.ExitFailure {
		t.Errorf("exit code = %d, want %d", code, command.ExitFailure)
	}
}

func TestRunHelpSucceeds(t *testing.T) {
	t.Parallel()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	code := command.Execute(context.Background(), boolfalse.New(), io, []string{"--help"})
	if code != command.ExitSuccess {
		t.Errorf("exit code = %d, want 0", code)
	}
	if !strings.Contains(out.String(), "Usage: false") {
		t.Errorf("help out = %q", out.String())
	}
}

func TestRunVersionSucceeds(t *testing.T) {
	t.Parallel()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	code := command.Execute(context.Background(), boolfalse.New(), io, []string{"--version"})
	if code != command.ExitSuccess {
		t.Errorf("exit code = %d, want 0", code)
	}
	if !strings.Contains(out.String(), "false (mimixbox)") {
		t.Errorf("version out = %q", out.String())
	}
}
