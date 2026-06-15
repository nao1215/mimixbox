package popmaildir

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func makeMaildir(t *testing.T, msgs map[string]string) string {
	t.Helper()
	root := t.TempDir()
	newDir := filepath.Join(root, "new")
	if err := os.MkdirAll(newDir, 0o700); err != nil {
		t.Fatal(err)
	}
	for name, content := range msgs {
		if err := os.WriteFile(filepath.Join(newDir, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func TestMoveToDest(t *testing.T) {
	root := makeMaildir(t, map[string]string{"1.msg": "alpha", "2.msg": "beta"})
	dest := filepath.Join(t.TempDir(), "out")
	io := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, []string{"-d", dest, root}); err != nil {
		t.Fatal(err)
	}
	for _, n := range []string{"1.msg", "2.msg"} {
		if _, err := os.Stat(filepath.Join(dest, n)); err != nil {
			t.Errorf("message %s not moved to dest: %v", n, err)
		}
		if _, err := os.Stat(filepath.Join(root, "new", n)); err == nil {
			t.Errorf("message %s still in maildir", n)
		}
	}
}

func TestPrintAndDrain(t *testing.T) {
	root := makeMaildir(t, map[string]string{"a": "first", "b": "second"})
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, []string{root}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "first") || !strings.Contains(out.String(), "second") {
		t.Errorf("output missing messages:\n%s", out.String())
	}
	entries, _ := os.ReadDir(filepath.Join(root, "new"))
	if len(entries) != 0 {
		t.Errorf("maildir not drained, %d files remain", len(entries))
	}
}

func TestKeep(t *testing.T) {
	root := makeMaildir(t, map[string]string{"a": "x"})
	io := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, []string{"-k", root}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "new", "a")); err != nil {
		t.Errorf("message removed despite -k: %v", err)
	}
}

func TestMissingMaildir(t *testing.T) {
	io := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, []string{filepath.Join(t.TempDir(), "nope")}); err == nil {
		t.Fatal("expected error for missing maildir")
	}
}

func TestRequiresOneArg(t *testing.T) {
	io := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, nil); err == nil {
		t.Fatal("expected error with no maildir")
	}
}
