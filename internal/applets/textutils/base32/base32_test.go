package base32_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/textutils/base32"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, stdin string, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(stdin), Out: out, Err: errBuf}
	err := base32.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestEncode(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "hello\n")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "NBSWY3DPBI======\n" {
		t.Errorf("out = %q", out)
	}
}

func TestDecode(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "NBSWY3DPBI======\n", "-d")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "hello\n" {
		t.Errorf("out = %q", out)
	}
}

func TestEncodeDecodeRoundTrip(t *testing.T) {
	t.Parallel()
	enc, _, err := run(t, "MimixBox")
	if err != nil {
		t.Fatalf("encode error = %v", err)
	}
	dec, _, err := run(t, enc, "-d")
	if err != nil {
		t.Fatalf("decode error = %v", err)
	}
	if dec != "MimixBox" {
		t.Errorf("round trip = %q", dec)
	}
}

func TestWrap(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, strings.Repeat("a", 100), "-w", "8")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	for _, line := range strings.Split(strings.TrimRight(out, "\n"), "\n") {
		if len(line) > 8 {
			t.Errorf("line too long: %q", line)
		}
	}
}

// GNU base32 ends the last line with a newline only when it wraps lines:
// -w 0 writes the encoding alone, and no input writes nothing.
func TestEncodeEndsWithANewlineOnlyWhenWrapping(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		stdin string
		args  []string
		want  string
	}{
		{"empty input writes nothing", "", nil, ""},
		{"empty input with wrap zero writes nothing", "", []string{"-w", "0"}, ""},
		{"wrap zero writes no newline", "a", []string{"-w", "0"}, "ME======"},
		{"wrapped output ends with one newline", "a", []string{"-w", "4"}, "ME==\n====\n"},
	}
	for _, tt := range tests {
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

func TestDecodeInvalid(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, "!!!not-base32!!!", "-d")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(errOut, "base32: invalid input") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestDecodeIgnoreGarbage(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "NBSW Y3DP\nBI======", "-d", "-i")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "hello\n" {
		t.Errorf("out = %q", out)
	}
}

func TestMissingFile(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, "", "/no/such/file")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(errOut, "base32:") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := base32.New()
	if c.Name() != "base32" {
		t.Errorf("Name() = %q", c.Name())
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}

func TestHelpSections(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "", "--help")
	if err != nil {
		t.Fatalf("--help error = %v", err)
	}
	if !strings.Contains(out, "Examples:") {
		t.Errorf("--help missing Examples: %q", out)
	}
	if !strings.Contains(out, "Exit status:") {
		t.Errorf("--help missing Exit status: %q", out)
	}
}
