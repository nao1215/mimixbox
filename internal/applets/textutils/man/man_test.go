package man

import (
	"bytes"
	"compress/gzip"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func fixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "man1"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "man5"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "man1", "foo.1"), []byte("FOO(1)\nfoo page\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	_, _ = zw.Write([]byte("BAR(5)\nbar page\n"))
	_ = zw.Close()
	if err := os.WriteFile(filepath.Join(root, "man5", "bar.5.gz"), buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: bytes.NewReader(nil), Out: out, Err: &bytes.Buffer{}}
	err := New().Run(context.Background(), io, args)
	return out.String(), err
}

func TestPlainPage(t *testing.T) {
	t.Parallel()
	out, err := run(t, "-M", fixture(t), "foo")
	if err != nil {
		t.Fatal(err)
	}
	if out != "FOO(1)\nfoo page\n" {
		t.Errorf("page = %q", out)
	}
}

func TestGzippedPage(t *testing.T) {
	t.Parallel()
	out, err := run(t, "-M", fixture(t), "5", "bar")
	if err != nil {
		t.Fatal(err)
	}
	if out != "BAR(5)\nbar page\n" {
		t.Errorf("gzipped page = %q", out)
	}
}

func TestSectionSearchAll(t *testing.T) {
	t.Parallel()
	// No section: bar lives in section 5 and is still found.
	out, err := run(t, "-M", fixture(t), "bar")
	if err != nil || out != "BAR(5)\nbar page\n" {
		t.Errorf("search-all = %q, %v", out, err)
	}
}

func TestNotFound(t *testing.T) {
	t.Parallel()
	_, err := run(t, "-M", fixture(t), "missing")
	var ee *command.ExitError
	if err == nil {
		t.Fatal("missing page should fail")
	}
	if e, ok := err.(*command.ExitError); ok {
		ee = e
	}
	if ee == nil || ee.Code != 16 {
		t.Errorf("err = %v, want exit 16", err)
	}
}

func TestIsSection(t *testing.T) {
	t.Parallel()
	for _, s := range []string{"1", "3", "3p", "8"} {
		if !isSection(s) {
			t.Errorf("isSection(%q) = false", s)
		}
	}
	for _, s := range []string{"foo", "", "ls"} {
		if isSection(s) {
			t.Errorf("isSection(%q) = true", s)
		}
	}
}
