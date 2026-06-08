package sed_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/editors/sed"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, stdin string, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(stdin), Out: out, Err: errBuf}
	err := sed.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := sed.New()
	if got := c.Name(); got != "sed" {
		t.Errorf("Name() = %q", got)
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis empty")
	}
}

func TestSubstitute(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		script string
		in     string
		want   string
	}{
		{"basic", "s/a/X/", "aaa\n", "Xaa\n"},
		{"global", "s/a/X/g", "aaa\n", "XXX\n"},
		{"nth", "s/a/X/2", "aaa\n", "aXa\n"},
		{"nth-global", "s/a/X/2g", "aaaa\n", "aXXX\n"},
		{"ignore-case", "s/a/X/gi", "aAa\n", "XXX\n"},
		{"whole-match-amp", "s/bc/[&]/", "abcd\n", "a[bc]d\n"},
		{"backref", `s/\(a\)\(b\)/\2\1/`, "ab\n", "ba\n"},
		{"alt-delim", "s|/usr|/opt|", "/usr/bin\n", "/opt/bin\n"},
		{"no-match", "s/zzz/Q/", "abc\n", "abc\n"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out, errOut, err := run(t, tt.in, tt.script)
			if err != nil {
				t.Fatalf("err = %v (stderr=%q)", err, errOut)
			}
			if out != tt.want {
				t.Errorf("script %q: out = %q, want %q", tt.script, out, tt.want)
			}
		})
	}
}

func TestPrintWithN(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "one\ntwo\nthree\n", "-n", "2p")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if out != "two\n" {
		t.Errorf("out = %q, want two", out)
	}
}

func TestDeleteLine(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "a\nb\nc\n", "2d")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if out != "a\nc\n" {
		t.Errorf("out = %q, want a\\nc", out)
	}
}

func TestDeleteRegex(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "keep\ndrop me\nkeep\n", "/drop/d")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if out != "keep\nkeep\n" {
		t.Errorf("out = %q", out)
	}
}

func TestRangeDelete(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "1\n2\n3\n4\n5\n", "2,4d")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if out != "1\n5\n" {
		t.Errorf("out = %q, want 1\\n5", out)
	}
}

func TestLastLineAddress(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "a\nb\nc\n", "-n", "$p")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if out != "c\n" {
		t.Errorf("out = %q, want c", out)
	}
}

func TestPrintFlagOnSub(t *testing.T) {
	t.Parallel()
	// -n with s///p prints only substituted lines.
	out, _, err := run(t, "cat\ndog\n", "-n", "s/cat/CAT/p")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if out != "CAT\n" {
		t.Errorf("out = %q, want CAT", out)
	}
}

func TestQuit(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "1\n2\n3\n", "2q")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if out != "1\n2\n" {
		t.Errorf("out = %q, want 1\\n2", out)
	}
}

func TestMultipleExpressions(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "abc\n", "-e", "s/a/X/", "-e", "s/c/Z/")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if out != "XbZ\n" {
		t.Errorf("out = %q, want XbZ", out)
	}
}

func TestSemicolonSeparatedCommands(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "a\nb\n", "s/a/X/;s/b/Y/")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if out != "X\nY\n" {
		t.Errorf("out = %q, want X\\nY", out)
	}
}

func TestScriptFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	script := filepath.Join(dir, "prog.sed")
	if err := os.WriteFile(script, []byte("s/foo/bar/g\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	out, _, err := run(t, "foo foo\n", "-f", script)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if out != "bar bar\n" {
		t.Errorf("out = %q", out)
	}
}

func TestInPlace(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	f := filepath.Join(dir, "data.txt")
	if err := os.WriteFile(f, []byte("hello\nworld\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, errOut, err := run(t, "", "-i", "s/world/earth/", f); err != nil {
		t.Fatalf("err = %v (stderr=%q)", err, errOut)
	}
	got, err := os.ReadFile(f)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "hello\nearth\n" {
		t.Errorf("file = %q", got)
	}
}

func TestFileOperand(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	f := filepath.Join(dir, "in.txt")
	if err := os.WriteFile(f, []byte("aaa\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	out, _, err := run(t, "", "s/a/b/g", f)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if out != "bbb\n" {
		t.Errorf("out = %q", out)
	}
}

func TestErrors(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"no script", nil, "no script"},
		{"unknown command", []string{"Z"}, "unknown command"},
		{"missing file", []string{"s/a/b/", "/no/such/file/x"}, "sed:"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, errOut, err := run(t, "", tt.args...)
			if err == nil {
				t.Errorf("expected error for %v", tt.args)
			}
			if !strings.Contains(errOut, tt.want) {
				t.Errorf("stderr = %q, want %q", errOut, tt.want)
			}
		})
	}
}

func TestHelp(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "", "--help")
	if err != nil {
		t.Fatalf("help err = %v", err)
	}
	if !strings.Contains(out, "Usage: sed") {
		t.Errorf("help = %q", out)
	}
}
