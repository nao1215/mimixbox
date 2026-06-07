package sortcmd_test

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	sortcmd "github.com/nao1215/mimixbox/internal/applets/shellutils/sort"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, stdin string, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(stdin), Out: out, Err: errBuf}
	err := sortcmd.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestRun(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		stdin string
		args  []string
		want  string
	}{
		{"lexical", "banana\napple\ncherry\n", nil, "apple\nbanana\ncherry\n"},
		{"reverse", "apple\nbanana\ncherry\n", []string{"-r"}, "cherry\nbanana\napple\n"},
		{"numeric", "10\n2\n1\n", []string{"-n"}, "1\n2\n10\n"},
		{"numeric lexical differs", "10\n2\n1\n", nil, "1\n10\n2\n"},
		{"unique", "b\na\nb\na\n", []string{"-u"}, "a\nb\n"},
		{"ignore case", "Banana\napple\nCherry\n", []string{"-f"}, "apple\nBanana\nCherry\n"},
		{"key field 2", "x 3\ny 1\nz 2\n", []string{"-k", "2"}, "y 1\nz 2\nx 3\n"},
		{"separator with key", "x,3\ny,1\nz,2\n", []string{"-t", ",", "-k", "2"}, "y,1\nz,2\nx,3\n"},
		{"ignore leading blanks", "  b\na\n", []string{"-b"}, "a\n  b\n"},
		{"numeric reverse", "1\n10\n2\n", []string{"-n", "-r"}, "10\n2\n1\n"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out, _, err := run(t, tt.stdin, tt.args...)
			if err != nil {
				t.Fatalf("Run error = %v", err)
			}
			if out != tt.want {
				t.Errorf("out = %q, want %q", out, tt.want)
			}
		})
	}
}

// TestAgainstSystemSort compares the applet output with the system sort for the
// option combinations that should match byte-for-byte across implementations.
func TestAgainstSystemSort(t *testing.T) {
	t.Parallel()
	if _, err := exec.LookPath("sort"); err != nil {
		t.Skip("system sort not available")
	}
	tests := []struct {
		name  string
		stdin string
		args  []string
	}{
		{"lexical", "banana\napple\ncherry\ndate\n", nil},
		{"reverse", "banana\napple\ncherry\n", []string{"-r"}},
		{"numeric", "10\n2\n1\n100\n3\n", []string{"-n"}},
		{"unique", "b\na\nb\nc\na\n", []string{"-u"}},
		{"key 2", "x 3\ny 1\nz 2\n", []string{"-k", "2"}},
		{"sep key", "x,3\ny,1\nz,2\n", []string{"-t", ",", "-k", "2"}},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, _, err := run(t, tt.stdin, tt.args...)
			if err != nil {
				t.Fatalf("Run error = %v", err)
			}
			cmd := exec.Command("sort", tt.args...)
			cmd.Stdin = strings.NewReader(tt.stdin)
			cmd.Env = append(os.Environ(), "LC_ALL=C")
			sysOut, cerr := cmd.Output()
			if cerr != nil {
				t.Fatalf("system sort error = %v", cerr)
			}
			if got != string(sysOut) {
				t.Errorf("applet = %q, system sort = %q", got, string(sysOut))
			}
		})
	}
}

func TestRunFiles(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	a := filepath.Join(dir, "a.txt")
	b := filepath.Join(dir, "b.txt")
	if err := os.WriteFile(a, []byte("c\na\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(b, []byte("b\nd\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	out, _, err := run(t, "", a, b)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "a\nb\nc\nd\n" {
		t.Errorf("out = %q", out)
	}
}

func TestRunOutputFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	o := filepath.Join(dir, "out.txt")
	out, _, err := run(t, "b\na\n", "-o", o)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "" {
		t.Errorf("stdout = %q, want empty", out)
	}
	data, rerr := os.ReadFile(o)
	if rerr != nil {
		t.Fatal(rerr)
	}
	if string(data) != "a\nb\n" {
		t.Errorf("file = %q", string(data))
	}
}

func TestRunCheck(t *testing.T) {
	t.Parallel()
	if _, _, err := run(t, "a\nb\nc\n", "-c"); err != nil {
		t.Errorf("sorted input -c error = %v", err)
	}
	_, errOut, err := run(t, "b\na\n", "-c")
	if err == nil {
		t.Fatal("expected error for unsorted -c input")
	}
	if !strings.Contains(errOut, "disorder") {
		t.Errorf("stderr = %q, want disorder message", errOut)
	}
}

func TestRunMissingFile(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, "", "/no/such/file")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	if !strings.Contains(errOut, "sort: /no/such/file:") {
		t.Errorf("stderr = %q, want sort error prefix", errOut)
	}
}

func TestRunHelpAndVersion(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "", "--help")
	if err != nil {
		t.Fatalf("--help error = %v", err)
	}
	if !strings.Contains(out, "Usage: sort") {
		t.Errorf("--help out = %q", out)
	}

	out, _, err = run(t, "", "--version")
	if err != nil {
		t.Fatalf("--version error = %v", err)
	}
	if !strings.Contains(out, "sort (mimixbox)") {
		t.Errorf("--version out = %q", out)
	}
}
