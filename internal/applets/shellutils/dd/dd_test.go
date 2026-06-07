package dd_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/shellutils/dd"
	"github.com/nao1215/mimixbox/internal/command"
)

// run executes dd with the given stdin string and arguments, returning the
// captured stdout, stderr, and the error.
func run(t *testing.T, stdin string, args ...string) (string, string, error) {
	t.Helper()
	var out, errBuf bytes.Buffer
	io := command.IO{
		In:  strings.NewReader(stdin),
		Out: &out,
		Err: &errBuf,
	}
	err := dd.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestNew(t *testing.T) {
	c := dd.New()
	if c == nil {
		t.Fatal("New() returned nil")
	}
	if c.Name() != "dd" {
		t.Errorf("Name() = %q, want %q", c.Name(), "dd")
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}

func TestRun_StdinToStdout(t *testing.T) {
	input := "hello, world\nsecond line\n"
	out, errOut, err := run(t, input)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if out != input {
		t.Errorf("stdout = %q, want %q", out, input)
	}
	if !strings.Contains(errOut, "records in") || !strings.Contains(errOut, "records out") {
		t.Errorf("stderr missing records summary: %q", errOut)
	}
}

func TestRun_BsCount(t *testing.T) {
	input := "abcdefghij"
	out, errOut, err := run(t, input, "bs=1", "count=5")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if out != "abcde" {
		t.Errorf("stdout = %q, want %q", out, "abcde")
	}
	if !strings.Contains(errOut, "5+0 records in") {
		t.Errorf("stderr = %q, want to contain %q", errOut, "5+0 records in")
	}
}

func TestRun_FileToFile(t *testing.T) {
	dir := t.TempDir()
	inPath := filepath.Join(dir, "in.txt")
	outPath := filepath.Join(dir, "out.txt")
	content := []byte("the quick brown fox\njumps over the lazy dog\n")
	if err := os.WriteFile(inPath, content, 0o644); err != nil {
		t.Fatal(err)
	}

	_, _, err := run(t, "", "if="+inPath, "of="+outPath)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	got, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, content) {
		t.Errorf("output file = %q, want %q", got, content)
	}
}

func TestRun_ConvUcase(t *testing.T) {
	out, _, err := run(t, "Hello World 123", "conv=ucase")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if out != "HELLO WORLD 123" {
		t.Errorf("stdout = %q, want %q", out, "HELLO WORLD 123")
	}
}

func TestRun_ConvLcase(t *testing.T) {
	out, _, err := run(t, "Hello World 123", "conv=lcase")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if out != "hello world 123" {
		t.Errorf("stdout = %q, want %q", out, "hello world 123")
	}
}

func TestRun_StatusNoneSuppressesSummary(t *testing.T) {
	out, errOut, err := run(t, "data", "status=none")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if out != "data" {
		t.Errorf("stdout = %q, want %q", out, "data")
	}
	if errOut != "" {
		t.Errorf("stderr = %q, want empty (status=none)", errOut)
	}
}

func TestRun_InvalidOperand(t *testing.T) {
	_, _, err := run(t, "data", "bogus")
	if err == nil {
		t.Fatal("Run() error = nil, want non-nil for invalid operand")
	}
}

func TestParseSize(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in      string
		want    int64
		wantErr bool
	}{
		{"512", 512, false},
		{"1k", 1024, false},
		{"1K", 1024, false},
		{"2b", 1024, false},
		{"1M", 1024 * 1024, false},
		{"1c", 1, false},
		{"2w", 4, false},
		{"1G", 1024 * 1024 * 1024, false},
		{"0", 0, false},
		{"", 0, true},
		{"abc", 0, true},
		{"-5", 0, true},
	}
	for _, tt := range tests {
		got, err := dd.ParseSize(tt.in)
		if tt.wantErr {
			if err == nil {
				t.Errorf("ParseSize(%q) error = nil, want error", tt.in)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseSize(%q) error = %v", tt.in, err)
			continue
		}
		if got != tt.want {
			t.Errorf("ParseSize(%q) = %d, want %d", tt.in, got, tt.want)
		}
	}
}
