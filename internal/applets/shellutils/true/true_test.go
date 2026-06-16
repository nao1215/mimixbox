package booltrue_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	booltrue "github.com/nao1215/mimixbox/internal/applets/shellutils/true"
	"github.com/nao1215/mimixbox/internal/command"
)

func TestRunAlwaysSucceeds(t *testing.T) {
	t.Parallel()
	io := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	code := command.Execute(context.Background(), booltrue.New(), io, []string{"ignored", "args"})
	if code != command.ExitSuccess {
		t.Errorf("exit code = %d, want %d", code, command.ExitSuccess)
	}
}

func TestRunHelp(t *testing.T) {
	t.Parallel()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	code := command.Execute(context.Background(), booltrue.New(), io, []string{"--help"})
	if code != command.ExitSuccess {
		t.Errorf("exit code = %d, want 0", code)
	}
	for _, want := range []string{"Usage: true", "Examples:", "Exit status:"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("help missing %q: %q", want, out.String())
		}
	}
}

func TestRunVersion(t *testing.T) {
	t.Parallel()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	code := command.Execute(context.Background(), booltrue.New(), io, []string{"--version"})
	if code != command.ExitSuccess {
		t.Errorf("exit code = %d, want 0", code)
	}
	if !strings.Contains(out.String(), "true (mimixbox)") {
		t.Errorf("version out = %q", out.String())
	}
}
