package pipeprogress

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func TestPassthrough(t *testing.T) {
	t.Parallel()
	in := bytes.Repeat([]byte("x"), 200*1024) // a few chunks
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: bytes.NewReader(in), Out: out, Err: errBuf}
	if err := New().Run(context.Background(), io, nil); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !bytes.Equal(out.Bytes(), in) {
		t.Errorf("stdout did not match stdin (%d vs %d bytes)", out.Len(), len(in))
	}
	if !strings.Contains(errBuf.String(), ".") {
		t.Errorf("expected progress dots on stderr, got %q", errBuf.String())
	}
}

func TestHelp(t *testing.T) {
	t.Parallel()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, []string{"--help"}); err != nil {
		t.Fatalf("--help err = %v", err)
	}
	if !strings.Contains(out.String(), "Usage: pipe_progress") {
		t.Errorf("--help = %q", out.String())
	}
}
