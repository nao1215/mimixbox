package tree

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func build(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	for _, d := range []string{"a/b", "c"} {
		if err := os.MkdirAll(filepath.Join(root, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	for _, f := range []string{"a/file1.txt", "a/b/deep.txt", "c/x.txt", "top.txt", ".hidden"} {
		if err := os.WriteFile(filepath.Join(root, f), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func run(t *testing.T, args ...string) string {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, args); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	return out.String()
}

// body returns the output with the first line (the root path) removed.
func body(out string) string {
	_, rest, _ := strings.Cut(out, "\n")
	return rest
}

func TestTreeDefault(t *testing.T) {
	t.Parallel()
	root := build(t)
	want := `├── a
│   ├── b
│   │   └── deep.txt
│   └── file1.txt
├── c
│   └── x.txt
└── top.txt

4 directories, 4 files
`
	if got := body(run(t, root)); got != want {
		t.Errorf("tree =\n%q\nwant\n%q", got, want)
	}
}

func TestTreeAllIncludesHidden(t *testing.T) {
	t.Parallel()
	root := build(t)
	if !strings.Contains(run(t, "-a", root), ".hidden") {
		t.Errorf("tree -a should include .hidden")
	}
	if strings.Contains(run(t, root), ".hidden") {
		t.Errorf("tree without -a should hide .hidden")
	}
}

func TestTreeDirsOnly(t *testing.T) {
	t.Parallel()
	root := build(t)
	out := run(t, "-d", root)
	if strings.Contains(out, "file1.txt") {
		t.Errorf("tree -d should not list files: %q", out)
	}
	if !strings.HasSuffix(out, "4 directories\n") {
		t.Errorf("tree -d summary = %q", out)
	}
}

func TestTreeLevel(t *testing.T) {
	t.Parallel()
	root := build(t)
	out := run(t, "-L", "1", root)
	if strings.Contains(out, "deep.txt") || strings.Contains(out, "file1.txt") {
		t.Errorf("tree -L 1 should not descend: %q", out)
	}
}
