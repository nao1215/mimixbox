package ed

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

// edit runs ed on a file seeded with content, feeding script as commands. It
// returns ed's stdout and the file's contents afterward.
func edit(t *testing.T, content, script string) (string, string) {
	t.Helper()
	dir := t.TempDir()
	f := filepath.Join(dir, "buf.txt")
	if err := os.WriteFile(f, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(script), Out: out, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, []string{f}); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	data, _ := os.ReadFile(f)
	return out.String(), string(data)
}

func TestHelpExitStatus(t *testing.T) {
	t.Parallel()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, []string{"--help"}); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(out.String(), "Exit status:") {
		t.Errorf("help missing exit status section = %q", out.String())
	}
}

func TestLoadAndPrint(t *testing.T) {
	t.Parallel()
	out, _ := edit(t, "one\ntwo\nthree\n", "1,$p\nq\n")
	if out != "14\none\ntwo\nthree\n" {
		t.Errorf("print = %q", out)
	}
}

func TestAppendAndWrite(t *testing.T) {
	t.Parallel()
	out, file := edit(t, "one\ntwo\nthree\n", "2a\nINSERTED\n.\nw\nq\n")
	if file != "one\ntwo\nINSERTED\nthree\n" {
		t.Errorf("file = %q", file)
	}
	// 14 on load, 23 after write.
	if out != "14\n23\n" {
		t.Errorf("byte counts = %q", out)
	}
}

func TestInsertBefore(t *testing.T) {
	t.Parallel()
	_, file := edit(t, "one\ntwo\n", "1i\nHEAD\n.\nw\nq\n")
	if file != "HEAD\none\ntwo\n" {
		t.Errorf("file = %q", file)
	}
}

func TestChange(t *testing.T) {
	t.Parallel()
	_, file := edit(t, "one\ntwo\nthree\n", "2c\nCHANGED\n.\nw\nq\n")
	if file != "one\nCHANGED\nthree\n" {
		t.Errorf("file = %q", file)
	}
}

func TestDelete(t *testing.T) {
	t.Parallel()
	_, file := edit(t, "one\ntwo\nthree\n", "1d\nw\nq\n")
	if file != "two\nthree\n" {
		t.Errorf("file = %q", file)
	}
}

func TestSubstitute(t *testing.T) {
	t.Parallel()
	_, file := edit(t, "foo bar\n", "1s/bar/BAZ/\nw\nq\n")
	if file != "foo BAZ\n" {
		t.Errorf("file = %q", file)
	}
	_, file = edit(t, "aaa\n", "1s/a/b/g\nw\nq\n")
	if file != "bbb\n" {
		t.Errorf("global sub = %q", file)
	}
}

func TestLineNumber(t *testing.T) {
	t.Parallel()
	out, _ := edit(t, "a\nb\nc\n", "=\nq\n")
	if out != "6\n3\n" { // 6 bytes on load, $ = 3
		t.Errorf("= -> %q", out)
	}
}

func TestRangeDelete(t *testing.T) {
	t.Parallel()
	_, file := edit(t, "1\n2\n3\n4\n", "2,3d\nw\nq\n")
	if file != "1\n4\n" {
		t.Errorf("range delete = %q", file)
	}
}
