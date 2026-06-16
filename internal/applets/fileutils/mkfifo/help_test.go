package mkfifo_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/fileutils/mkfifo"
	"github.com/nao1215/mimixbox/internal/command"
)

func TestHelpSections(t *testing.T) {
	t.Parallel()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := mkfifo.New().Run(context.Background(), io, []string{"--help"}); err != nil {
		t.Fatalf("--help returned error: %v", err)
	}
	if !strings.Contains(out.String(), "Examples:") {
		t.Errorf("help output missing Examples section:\n%s", out.String())
	}
	if !strings.Contains(out.String(), "Exit status:") {
		t.Errorf("help output missing Exit status section:\n%s", out.String())
	}
}
