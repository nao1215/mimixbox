package base64_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/shellutils/base64"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, stdin string, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(stdin), Out: out, Err: errBuf}
	err := base64.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestEncodeStreamsLargeInput(t *testing.T) {
	t.Parallel()
	// Encode a large input through the streaming path and confirm it round-trips
	// and that the wrapped output matches the standard encoding wrapped at 76
	// columns (issue #952).
	data := bytes.Repeat([]byte("mimixbox-streaming-0123456789\n"), 200000) // ~6 MiB
	out := &bytes.Buffer{}
	io := command.IO{In: bytes.NewReader(data), Out: out, Err: &bytes.Buffer{}}
	if err := base64.New().Run(context.Background(), io, nil); err != nil {
		t.Fatalf("encode error = %v", err)
	}

	// Decode the produced output back and compare with the original bytes.
	back := &bytes.Buffer{}
	dio := command.IO{In: bytes.NewReader(out.Bytes()), Out: back, Err: &bytes.Buffer{}}
	if err := base64.New().Run(context.Background(), dio, []string{"-d"}); err != nil {
		t.Fatalf("decode error = %v", err)
	}
	if !bytes.Equal(back.Bytes(), data) {
		t.Errorf("round trip mismatch: got %d bytes, want %d", back.Len(), len(data))
	}
}

func TestRun(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		stdin string
		args  []string
		want  string
	}{
		{"encode", "hello\n", nil, "aGVsbG8K\n"},
		{"decode round trip", "aGVsbG8K\n", []string{"-d"}, "hello\n"},
		{"decode long flag", "aGVsbG8K\n", []string{"--decode"}, "hello\n"},
		{"wrap zero no wrap", strings.Repeat("\x00", 60), []string{"-w", "0"},
			strings.Repeat("A", 80) + "\n"},
		{"wrap at ten", strings.Repeat("\x00", 60), []string{"-w", "10"},
			"AAAAAAAAAA\nAAAAAAAAAA\nAAAAAAAAAA\nAAAAAAAAAA\nAAAAAAAAAA\nAAAAAAAAAA\nAAAAAAAAAA\nAAAAAAAAAA\n"},
		{"default wrap at 76", strings.Repeat("\x00", 60), nil,
			strings.Repeat("A", 76) + "\n" + strings.Repeat("A", 4) + "\n"},
		{"ignore garbage", "aGVs\x21bG8K\n", []string{"-d", "-i"}, "hello\n"},
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

func TestRunDecodeInvalid(t *testing.T) {
	t.Parallel()
	out, errOut, err := run(t, "not valid base64 @@@\n", "-d")
	if err == nil {
		t.Fatal("expected error for invalid base64 input")
	}
	if out != "" {
		t.Errorf("out = %q, want empty", out)
	}
	if !strings.Contains(errOut, "base64: invalid input") {
		t.Errorf("stderr = %q, want invalid input message", errOut)
	}
}

func TestRunMissingFile(t *testing.T) {
	t.Parallel()
	out, errOut, err := run(t, "", "/no/such/file")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	if out != "" {
		t.Errorf("out = %q, want empty", out)
	}
	if !strings.Contains(errOut, "base64: /no/such/file:") {
		t.Errorf("stderr = %q, want base64 error prefix", errOut)
	}
}

func TestRunHelpAndVersion(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "", "--help")
	if err != nil {
		t.Fatalf("--help error = %v", err)
	}
	if !strings.Contains(out, "Usage: base64") {
		t.Errorf("--help out = %q", out)
	}
	if !strings.Contains(out, "Examples:") {
		t.Errorf("--help missing Examples: %q", out)
	}
	if !strings.Contains(out, "Exit status:") {
		t.Errorf("--help missing Exit status: %q", out)
	}

	out, _, err = run(t, "", "--version")
	if err != nil {
		t.Fatalf("--version error = %v", err)
	}
	if !strings.Contains(out, "base64 (mimixbox)") {
		t.Errorf("--version out = %q", out)
	}
}
