package stat_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/fileutils/stat"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := stat.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func tmpFile(t *testing.T, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "f.txt")
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestFormatSizeAndName(t *testing.T) {
	t.Parallel()
	p := tmpFile(t, "hello")
	out, _, err := run(t, "-c", "%n %s", p)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != p+" 5\n" {
		t.Errorf("out = %q", out)
	}
}

func TestFormatPermsAndType(t *testing.T) {
	t.Parallel()
	p := tmpFile(t, "x")
	out, _, err := run(t, "-c", "%a %F", p)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "644 regular file\n" {
		t.Errorf("out = %q", out)
	}
}

func TestFormatDirectoryType(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	out, _, err := run(t, "-c", "%F", dir)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "directory\n" {
		t.Errorf("out = %q", out)
	}
}

func TestFormatEscapesAndLiteralPercent(t *testing.T) {
	t.Parallel()
	p := tmpFile(t, "ab")
	out, _, err := run(t, "-c", `%s\t100%%`, p)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "2\t100%\n" {
		t.Errorf("out = %q", out)
	}
}

func TestDefaultLayout(t *testing.T) {
	t.Parallel()
	p := tmpFile(t, "hi")
	out, _, err := run(t, p)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	for _, want := range []string{"File:", "Size:", "Access:", "Modify:"} {
		if !strings.Contains(out, want) {
			t.Errorf("default layout missing %q in %q", want, out)
		}
	}
}

func TestUnknownSpecifierPassThrough(t *testing.T) {
	t.Parallel()
	p := tmpFile(t, "x")
	// %q is not a directive stat implements, so it is emitted verbatim.
	out, _, err := run(t, "-c", "%q", p)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "%q\n" {
		t.Errorf("out = %q", out)
	}
}

func TestFormatStatFields(t *testing.T) {
	t.Parallel()
	p := tmpFile(t, "abc")
	// %A perms string, %i inode, %h links, %u uid, %g gid, %f raw mode hex.
	out, _, err := run(t, "-c", "%A|%i|%h|%u|%g|%f", p)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	fields := strings.Split(strings.TrimRight(out, "\n"), "|")
	if len(fields) != 6 {
		t.Fatalf("got %d fields: %q", len(fields), out)
	}
	if !strings.HasPrefix(fields[0], "-rw") {
		t.Errorf("%%A = %q", fields[0])
	}
	for i, f := range fields[1:] {
		if f == "" {
			t.Errorf("field %d is empty", i+1)
		}
	}
}

func TestDereference(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	target := filepath.Join(dir, "target")
	linkPath := filepath.Join(dir, "link")
	if err := os.WriteFile(target, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, linkPath); err != nil {
		t.Fatal(err)
	}
	// Without -L, lstat sees a symbolic link; with -L it follows to the file.
	out, _, err := run(t, "-L", "-c", "%F %s", linkPath)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "regular file 5\n" {
		t.Errorf("out = %q", out)
	}
}

func TestSymlinkType(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	target := filepath.Join(dir, "t")
	linkPath := filepath.Join(dir, "l")
	if err := os.WriteFile(target, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, linkPath); err != nil {
		t.Fatal(err)
	}
	out, _, err := run(t, "-c", "%F", linkPath)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "symbolic link\n" {
		t.Errorf("out = %q", out)
	}
}

func TestMissingFile(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, "/no/such/file")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(errOut, "stat: cannot stat") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestMissingOperand(t *testing.T) {
	t.Parallel()
	_, _, err := run(t)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "missing operand") {
		t.Errorf("err = %v", err)
	}
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := stat.New()
	if c.Name() != "stat" {
		t.Errorf("Name() = %q", c.Name())
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}

func TestPrintfNoTrailingNewlineAndEscapes(t *testing.T) {
	t.Parallel()
	p := tmpFile(t, "hello")
	// --printf interprets backslash escapes and adds no trailing newline.
	out, _, err := run(t, "--printf", `%n=%s\n`, p)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != p+"=5\n" {
		t.Errorf("out = %q, want %q", out, p+"=5\n")
	}
}

func TestPrintfNoNewlineWhenNoEscape(t *testing.T) {
	t.Parallel()
	p := tmpFile(t, "ab")
	// Without an explicit \n, --printf emits nothing extra.
	out, _, err := run(t, "--printf", "%s", p)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "2" {
		t.Errorf("out = %q, want %q", out, "2")
	}
}

func TestPrintfDirectives(t *testing.T) {
	t.Parallel()
	p := tmpFile(t, "abc")
	// %a octal perms, %i inode, %Y mtime epoch. All numeric/non-empty.
	out, _, err := run(t, "--printf", "%a|%i|%Y|%b|%B", p)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	fields := strings.Split(out, "|")
	if len(fields) != 5 {
		t.Fatalf("got %d fields: %q", len(fields), out)
	}
	if fields[0] != "644" {
		t.Errorf("%%a = %q, want 644", fields[0])
	}
	for i, f := range fields {
		if f == "" {
			t.Errorf("field %d empty in %q", i, out)
		}
	}
}

func TestFormatAddsTrailingNewline(t *testing.T) {
	t.Parallel()
	p := tmpFile(t, "xy")
	// -c/--format honors backslash escapes too, plus a trailing newline.
	out, _, err := run(t, "--format", `%s`, p)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "2\n" {
		t.Errorf("out = %q, want %q", out, "2\n")
	}
}

func TestTerseLine(t *testing.T) {
	t.Parallel()
	p := tmpFile(t, "hello")
	out, _, err := run(t, "--terse", p)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.HasSuffix(out, "\n") {
		t.Errorf("terse output should end with newline: %q", out)
	}
	fields := strings.Fields(strings.TrimRight(out, "\n"))
	// name size blocks rawmode uid gid dev inode nlink major minor atime mtime ctime blksize = 15 fields.
	if len(fields) != 15 {
		t.Fatalf("terse line has %d fields, want 15: %q", len(fields), out)
	}
	if fields[0] != p {
		t.Errorf("terse name = %q, want %q", fields[0], p)
	}
	if fields[1] != "5" {
		t.Errorf("terse size = %q, want 5", fields[1])
	}
}

func TestHelpSections(t *testing.T) {
	t.Parallel()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := stat.New().Run(context.Background(), io, []string{"--help"}); err != nil {
		t.Fatalf("Run --help error = %v", err)
	}
	for _, want := range []string{"Examples:", "Exit status:"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("--help missing %q section:\n%s", want, out.String())
		}
	}
}
