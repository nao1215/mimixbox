package awk_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/editors/awk"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, stdin string, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(stdin), Out: out, Err: errBuf}
	err := awk.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := awk.New()
	if got := c.Name(); got != "awk" {
		t.Errorf("Name() = %q", got)
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis empty")
	}
}

func TestPrintWholeLine(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "a b\nc d\n", "{print}")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if out != "a b\nc d\n" {
		t.Errorf("out = %q", out)
	}
}

func TestPrintField(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "one two three\n", "{print $2}")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if out != "two\n" {
		t.Errorf("out = %q, want two", out)
	}
}

func TestPrintMultipleFields(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "a b c\n", "{print $1, $3}")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if out != "a c\n" {
		t.Errorf("out = %q, want 'a c'", out)
	}
}

func TestFieldSeparator(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "root:x:0:0\n", "-F", ":", "{print $1}")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if out != "root\n" {
		t.Errorf("out = %q, want root", out)
	}
}

func TestRegexPattern(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "apple\nbanana\ncherry\n", "/an/{print}")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if out != "banana\n" {
		t.Errorf("out = %q, want banana", out)
	}
}

func TestRegexDefaultAction(t *testing.T) {
	t.Parallel()
	// A bare /regex/ pattern prints matching lines.
	out, _, err := run(t, "yes\nno\nyes\n", "/yes/")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if out != "yes\nyes\n" {
		t.Errorf("out = %q", out)
	}
}

func TestNRComparison(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "l1\nl2\nl3\n", "NR==2")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if out != "l2\n" {
		t.Errorf("out = %q, want l2", out)
	}
}

func TestFieldComparison(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "alice 30\nbob 25\n", "$2 > 27 {print $1}")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if out != "alice\n" {
		t.Errorf("out = %q, want alice", out)
	}
}

func TestStringComparison(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "cat\ndog\ncat\n", `$1 == "cat"`)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if out != "cat\ncat\n" {
		t.Errorf("out = %q", out)
	}
}

func TestNF(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "a b c\n", "{print NF}")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if out != "3\n" {
		t.Errorf("out = %q, want 3", out)
	}
}

func TestLastField(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "a b c d\n", "{print $NF}")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if out != "d\n" {
		t.Errorf("out = %q, want d", out)
	}
}

func TestBeginEnd(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "x\ny\nz\n", "BEGIN{print \"start\"} END{print NR}")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if out != "start\n3\n" {
		t.Errorf("out = %q, want start\\n3", out)
	}
}

func TestVarAssign(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "x\n", "-v", "name=world", "{print name}")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if out != "world\n" {
		t.Errorf("out = %q, want world", out)
	}
}

func TestStringLiteral(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "a\n", `{print "hello", $1}`)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if out != "hello a\n" {
		t.Errorf("out = %q", out)
	}
}

func TestPrintf(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "bob\n", `{printf "name=%s\n", $1}`)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if out != "name=bob\n" {
		t.Errorf("out = %q", out)
	}
}

func TestFileOperand(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	f := filepath.Join(dir, "in.txt")
	if err := os.WriteFile(f, []byte("p q\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	out, _, err := run(t, "", "{print $2}", f)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if out != "q\n" {
		t.Errorf("out = %q, want q", out)
	}
}

func TestErrors(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"no program", nil, "no program"},
		{"missing file", []string{"{print}", "/no/such/file/zz"}, "awk:"},
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
	if !strings.Contains(out, "Usage: awk") {
		t.Errorf("help = %q", out)
	}
	if !strings.Contains(out, "Examples:") {
		t.Errorf("help missing Examples: %q", out)
	}
	if !strings.Contains(out, "Exit status:") {
		t.Errorf("help missing Exit status: %q", out)
	}
}
