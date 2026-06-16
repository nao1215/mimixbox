package find_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/findutils/find"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := find.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

// tree builds a fixed directory layout under a temp dir and returns its root:
//
//	root/
//	  a.txt        (regular, "hi")
//	  empty.txt    (regular, empty)
//	  sub/         (directory)
//	    b.log      (regular, "log")
//	  emptydir/    (directory, empty)
func tree(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "a.txt"), "hi")
	mustWrite(t, filepath.Join(root, "empty.txt"), "")
	if err := os.Mkdir(filepath.Join(root, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	mustWrite(t, filepath.Join(root, "sub", "b.log"), "log")
	if err := os.Mkdir(filepath.Join(root, "emptydir"), 0o755); err != nil {
		t.Fatal(err)
	}
	return root
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func lines(out string) []string {
	out = strings.TrimRight(out, "\n")
	if out == "" {
		return nil
	}
	ls := strings.Split(out, "\n")
	sort.Strings(ls)
	return ls
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := find.New()
	if got := c.Name(); got != "find" {
		t.Errorf("Name() = %q", got)
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis empty")
	}
}

func TestPrintAll(t *testing.T) {
	t.Parallel()
	root := tree(t)
	out, _, err := run(t, root)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	got := lines(out)
	// root itself plus 5 entries.
	if len(got) != 6 {
		t.Errorf("got %d paths, want 6: %v", len(got), got)
	}
}

func TestName(t *testing.T) {
	t.Parallel()
	root := tree(t)
	out, _, err := run(t, root, "-name", "*.txt")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	got := lines(out)
	want := lines(filepath.Join(root, "a.txt") + "\n" + filepath.Join(root, "empty.txt"))
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestINameCaseInsensitive(t *testing.T) {
	t.Parallel()
	root := tree(t)
	out, _, err := run(t, root, "-iname", "*.TXT")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if len(lines(out)) != 2 {
		t.Errorf("-iname got %v, want 2 files", lines(out))
	}
}

func TestTypeFile(t *testing.T) {
	t.Parallel()
	root := tree(t)
	out, _, err := run(t, root, "-type", "f")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if len(lines(out)) != 3 {
		t.Errorf("-type f got %v, want 3", lines(out))
	}
}

func TestTypeDir(t *testing.T) {
	t.Parallel()
	root := tree(t)
	out, _, err := run(t, root, "-type", "d")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	// root, sub, emptydir
	if len(lines(out)) != 3 {
		t.Errorf("-type d got %v, want 3", lines(out))
	}
}

func TestMaxDepth(t *testing.T) {
	t.Parallel()
	root := tree(t)
	out, _, err := run(t, root, "-maxdepth", "1", "-type", "f")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	// Only top-level files: a.txt, empty.txt (b.log is at depth 2).
	if len(lines(out)) != 2 {
		t.Errorf("-maxdepth 1 -type f got %v, want 2", lines(out))
	}
}

func TestMinDepth(t *testing.T) {
	t.Parallel()
	root := tree(t)
	out, _, err := run(t, root, "-mindepth", "2")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	got := lines(out)
	// Only sub/b.log is at depth 2.
	if len(got) != 1 || !strings.HasSuffix(got[0], filepath.Join("sub", "b.log")) {
		t.Errorf("-mindepth 2 got %v, want sub/b.log", got)
	}
}

func TestEmpty(t *testing.T) {
	t.Parallel()
	root := tree(t)
	out, _, err := run(t, root, "-empty")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	got := lines(out)
	// empty.txt and emptydir
	if len(got) != 2 {
		t.Errorf("-empty got %v, want 2", got)
	}
}

func TestPrint0(t *testing.T) {
	t.Parallel()
	root := tree(t)
	out, _, err := run(t, root, "-name", "a.txt", "-print0")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if !strings.HasSuffix(out, "a.txt\x00") {
		t.Errorf("-print0 out = %q, want NUL terminator", out)
	}
}

func TestDefaultPath(t *testing.T) {
	t.Parallel()
	// With no path, find walks "." which always contains the test binary cwd.
	out, _, err := run(t, "-maxdepth", "0")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if strings.TrimSpace(out) != "." {
		t.Errorf("default path out = %q, want '.'", out)
	}
}

func TestMultipleRoots(t *testing.T) {
	t.Parallel()
	root := tree(t)
	out, _, err := run(t, filepath.Join(root, "sub"), filepath.Join(root, "emptydir"))
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if len(lines(out)) != 3 { // sub, sub/b.log, emptydir
		t.Errorf("got %v, want 3", lines(out))
	}
}

func TestErrors(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"unknown predicate", []string{".", "-bogus"}, "unknown predicate"},
		{"missing -name arg", []string{".", "-name"}, "missing argument"},
		{"bad maxdepth", []string{".", "-maxdepth", "x"}, "invalid argument"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, errOut, err := run(t, tt.args...)
			if err == nil {
				t.Errorf("expected error for %v", tt.args)
			}
			if !strings.Contains(errOut, tt.want) {
				t.Errorf("stderr = %q, want %q", errOut, tt.want)
			}
		})
	}
}

func TestMissingRoot(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, "/no/such/path/xyz")
	if err == nil {
		t.Error("expected error for missing root")
	}
	if !strings.Contains(errOut, "find:") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestHelp(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "--help")
	if err != nil {
		t.Fatalf("help err = %v", err)
	}
	if !strings.Contains(out, "Usage: find") {
		t.Errorf("help = %q", out)
	}
	// GNU-style help must carry an Options: block listing --help and --version,
	// while still documenting the supported subset tokens.
	for _, want := range []string{
		"Options:", "--help", "--version",
		"-name", "-type", "-print0", "-maxdepth", "-mindepth",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("help missing %q:\n%s", want, out)
		}
	}
}

// TestVersion verifies that --version prints the version line, not usage text
// (the bug fixed for issue #236).
func TestVersion(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "--version")
	if err != nil {
		t.Fatalf("version err = %v", err)
	}
	if !strings.Contains(out, "find (mimixbox)") {
		t.Errorf("version = %q", out)
	}
	if strings.Contains(out, "Usage:") {
		t.Errorf("version output should not contain usage text: %q", out)
	}
}
