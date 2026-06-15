package readahead

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	var out, errBuf bytes.Buffer
	stdio := command.IO{In: strings.NewReader(""), Out: &out, Err: &errBuf}
	err := New().Run(context.Background(), stdio, args)
	return errBuf.String(), err
}

func TestReadaheadPreloadsFiles(t *testing.T) {
	var got []string
	prev := preload
	preload = func(p string) error { got = append(got, p); return nil }
	t.Cleanup(func() { preload = prev })

	if _, err := run(t, "a.txt", "b.txt"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 || got[0] != "a.txt" || got[1] != "b.txt" {
		t.Errorf("unexpected preload list: %v", got)
	}
}

func TestReadaheadRealFile(t *testing.T) {
	p := filepath.Join(t.TempDir(), "data")
	if err := os.WriteFile(p, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := run(t, p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestReadaheadMissingFile(t *testing.T) {
	errOut, err := run(t, filepath.Join(t.TempDir(), "absent"))
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	if !strings.Contains(errOut, "readahead:") {
		t.Errorf("missing prefix: %q", errOut)
	}
}

func TestReadaheadNoArgs(t *testing.T) {
	if _, err := run(t); err == nil {
		t.Fatal("expected error with no operands")
	}
}
