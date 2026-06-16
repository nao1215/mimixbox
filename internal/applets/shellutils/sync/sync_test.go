package sync_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/shellutils/sync"
	"github.com/nao1215/mimixbox/internal/command"
)

func TestNew(t *testing.T) {
	t.Parallel()
	if sync.New() == nil {
		t.Fatal("New() = nil")
	}
}

func TestName(t *testing.T) {
	t.Parallel()
	if got := sync.New().Name(); got != "sync" {
		t.Errorf("Name() = %q, want %q", got, "sync")
	}
}

func TestSynopsis(t *testing.T) {
	t.Parallel()
	want := "Synchronize cached writes to persistent storage"
	if got := sync.New().Synopsis(); got != want {
		t.Errorf("Synopsis() = %q, want %q", got, want)
	}
}

func TestRun(t *testing.T) {
	t.Parallel()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}

	if err := sync.New().Run(context.Background(), io, nil); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out.Len() != 0 {
		t.Errorf("stdout = %q, want empty", out.String())
	}
	if errBuf.Len() != 0 {
		t.Errorf("stderr = %q, want empty", errBuf.String())
	}
}

func TestHelpSections(t *testing.T) {
	t.Parallel()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := sync.New().Run(context.Background(), io, []string{"--help"}); err != nil {
		t.Fatalf("Run --help error = %v", err)
	}
	for _, want := range []string{"Examples:", "Exit status:"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("--help missing %q section:\n%s", want, out.String())
		}
	}
}
