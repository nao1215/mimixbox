package ls

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func fixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	for _, f := range []string{"b.txt", "a.txt", ".hidden"} {
		if err := os.WriteFile(filepath.Join(dir, f), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.MkdirAll(filepath.Join(dir, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "sub", "inner.txt"), []byte("y"), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	err := New().Run(context.Background(), io, args)
	return out.String(), err
}

func TestDefaultSorted(t *testing.T) {
	t.Parallel()
	out, err := run(t, fixture(t))
	if err != nil {
		t.Fatal(err)
	}
	if out != "a.txt\nb.txt\nsub\n" {
		t.Errorf("default = %q", out)
	}
}

func TestAll(t *testing.T) {
	t.Parallel()
	out, _ := run(t, "-a", fixture(t))
	if out != ".\n..\n.hidden\na.txt\nb.txt\nsub\n" {
		t.Errorf("-a = %q", out)
	}
}

func TestAlmostAll(t *testing.T) {
	t.Parallel()
	out, _ := run(t, "-A", fixture(t))
	if out != ".hidden\na.txt\nb.txt\nsub\n" {
		t.Errorf("-A = %q", out)
	}
}

func TestClassify(t *testing.T) {
	t.Parallel()
	out, _ := run(t, "-F", fixture(t))
	if !strings.Contains(out, "sub/") {
		t.Errorf("-F should mark directories: %q", out)
	}
}

func TestDirectorySelf(t *testing.T) {
	t.Parallel()
	dir := fixture(t)
	out, _ := run(t, "-d", dir)
	if out != dir+"\n" {
		t.Errorf("-d = %q, want %q", out, dir+"\n")
	}
}

func TestLong(t *testing.T) {
	t.Parallel()
	out, _ := run(t, "-l", fixture(t))
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 3 {
		t.Fatalf("-l produced %d lines: %q", len(lines), out)
	}
	// Each line begins with a 10-char mode string and ends with the name.
	if !strings.HasPrefix(lines[0], "-rw") || !strings.HasSuffix(lines[0], "a.txt") {
		t.Errorf("-l line = %q", lines[0])
	}
	if !strings.HasPrefix(lines[2], "d") {
		t.Errorf("directory mode should start with d: %q", lines[2])
	}
}

func TestRecursive(t *testing.T) {
	t.Parallel()
	out, _ := run(t, "-R", fixture(t))
	if !strings.Contains(out, "sub:") || !strings.Contains(out, "inner.txt") {
		t.Errorf("-R = %q", out)
	}
}

func TestMissingFile(t *testing.T) {
	t.Parallel()
	_, err := run(t, "/no/such/ls/file")
	var ee *command.ExitError
	if err == nil {
		t.Fatal("missing file should fail")
	}
	if e, ok := err.(*command.ExitError); ok {
		ee = e
	}
	if ee == nil || ee.Code != 2 {
		t.Errorf("err = %v, want exit 2", err)
	}
}

func TestModeString(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	f := filepath.Join(dir, "x")
	if err := os.WriteFile(f, []byte("x"), 0o640); err != nil {
		t.Fatal(err)
	}
	info, _ := os.Lstat(f)
	if got := modeString(info); got != "-rw-r-----" {
		t.Errorf("modeString = %q, want -rw-r-----", got)
	}
}

func TestSizeString(t *testing.T) {
	t.Parallel()
	cases := map[int64]string{0: "0", 512: "512", 1024: "1.0K", 1048576: "1.0M"}
	for in, want := range cases {
		if got := sizeString(in, true); got != want {
			t.Errorf("sizeString(%d) = %q, want %q", in, got, want)
		}
	}
	if got := sizeString(2048, false); got != "2048" {
		t.Errorf("sizeString plain = %q", got)
	}
}
