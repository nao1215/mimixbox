package busybox

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func out(t *testing.T, args ...string) string {
	t.Helper()
	o := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: o, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, args); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	return o.String()
}

func TestVersion(t *testing.T) {
	t.Parallel()
	if got := out(t, "--version"); !strings.Contains(got, "busybox") {
		t.Errorf("--version = %q", got)
	}
}

func TestHelpAndNoArgs(t *testing.T) {
	t.Parallel()
	if got := out(t, "--help"); !strings.Contains(got, "Usage: busybox") {
		t.Errorf("--help = %q", got)
	}
	if got := out(t); !strings.Contains(got, "Usage: busybox") {
		t.Errorf("no args = %q", got)
	}
}
