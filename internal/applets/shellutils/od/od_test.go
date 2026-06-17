package od_test

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/shellutils/od"
	"github.com/nao1215/mimixbox/internal/command"
)

func runStdin(t *testing.T, stdin string, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(stdin), Out: out, Err: errBuf}
	err := od.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

// dripReader returns its data one byte per Read call so the row formatter is
// forced to assemble rows across many read boundaries.
type dripReader struct {
	data []byte
	pos  int
}

func (d *dripReader) Read(p []byte) (int, error) {
	if d.pos >= len(d.data) {
		return 0, io.EOF
	}
	p[0] = d.data[d.pos]
	d.pos++
	return 1, nil
}

func TestStreamingMatchesSingleRead(t *testing.T) {
	t.Parallel()
	// Streaming the input one byte at a time must produce the same dump as a
	// single read, across row boundaries and the trailing offset line (#952).
	data := bytes.Repeat([]byte{0x00, 0x41, 0xff, '\n', 0x7f, 0x10}, 5000) // ~30 KiB
	for _, args := range [][]string{nil, {"-c"}, {"-A", "x", "-t", "x1"}} {
		whole := &bytes.Buffer{}
		io1 := command.IO{In: bytes.NewReader(data), Out: whole, Err: &bytes.Buffer{}}
		if err := od.New().Run(context.Background(), io1, args); err != nil {
			t.Fatalf("whole run error = %v", err)
		}
		drip := &bytes.Buffer{}
		io2 := command.IO{In: &dripReader{data: data}, Out: drip, Err: &bytes.Buffer{}}
		if err := od.New().Run(context.Background(), io2, args); err != nil {
			t.Fatalf("drip run error = %v", err)
		}
		if whole.String() != drip.String() {
			t.Errorf("args %v: streaming output differs from single read", args)
		}
	}
}

// Expected outputs are verified against GNU od, e.g. `printf 'ABC\n' | od -c`.
func TestRun(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		stdin string
		args  []string
		want  string
	}{
		{
			name:  "char dump",
			stdin: "ABC\n",
			args:  []string{"-c"},
			want:  "0000000   A   B   C  \\n\n0000004\n",
		},
		{
			name:  "hex address hex bytes",
			stdin: "ABC\n",
			args:  []string{"-A", "x", "-t", "x1"},
			want:  "000000 41 42 43 0a\n000004\n",
		},
		{
			name:  "octal bytes",
			stdin: "ABC\n",
			args:  []string{"-t", "o1"},
			want:  "0000000 101 102 103 012\n0000004\n",
		},
		{
			name:  "no addresses",
			stdin: "ABC\n",
			args:  []string{"-A", "n", "-t", "x1"},
			want:  " 41 42 43 0a\n",
		},
		{
			name:  "default is octal words",
			stdin: "ABC\n",
			args:  nil,
			want:  "0000000 041101 005103\n0000004\n",
		},
		{
			name:  "shortcut -b",
			stdin: "ABC\n",
			args:  []string{"-b"},
			want:  "0000000 101 102 103 012\n0000004\n",
		},
		{
			name:  "shortcut -x",
			stdin: "ABC\n",
			args:  []string{"-x"},
			want:  "0000000 4241 0a43\n0000004\n",
		},
		{
			name:  "shortcut -o",
			stdin: "ABC\n",
			args:  []string{"-o"},
			want:  "0000000 041101 005103\n0000004\n",
		},
		{
			name:  "decimal address",
			stdin: "ABC\n",
			args:  []string{"-A", "d", "-t", "x1"},
			want:  "0000000 41 42 43 0a\n0000004\n",
		},
		{
			name:  "empty input",
			stdin: "",
			args:  []string{"-c"},
			want:  "0000000\n",
		},
		{
			name:  "two lines wrap at 16 bytes",
			stdin: "Hello, World!\nThis is a test of more than sixteen bytes here.\n",
			args:  []string{"-c"},
			want: "0000000   H   e   l   l   o   ,       W   o   r   l   d   !  \\n   T   h\n" +
				"0000020   i   s       i   s       a       t   e   s   t       o   f    \n" +
				"0000040   m   o   r   e       t   h   a   n       s   i   x   t   e   e\n" +
				"0000060   n       b   y   t   e   s       h   e   r   e   .  \\n\n" +
				"0000076\n",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out, _, err := runStdin(t, tt.stdin, tt.args...)
			if err != nil {
				t.Fatalf("Run error = %v", err)
			}
			if out != tt.want {
				t.Errorf("out = %q, want %q", out, tt.want)
			}
		})
	}
}

func TestRunFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	f := filepath.Join(dir, "a.bin")
	if err := os.WriteFile(f, []byte("ABC\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	out, _, err := runStdin(t, "", "-A", "x", "-t", "x1", f)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	want := "000000 41 42 43 0a\n000004\n"
	if out != want {
		t.Errorf("out = %q, want %q", out, want)
	}
}

func TestRunMissingFile(t *testing.T) {
	t.Parallel()
	_, errOut, err := runStdin(t, "", "/no/such/file")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	if !strings.Contains(errOut, "od: /no/such/file:") {
		t.Errorf("stderr = %q, want od error prefix", errOut)
	}
}

func TestRunInvalidRadix(t *testing.T) {
	t.Parallel()
	_, errOut, err := runStdin(t, "ABC\n", "-A", "z")
	if err == nil {
		t.Fatal("expected error for invalid radix")
	}
	if !strings.Contains(errOut, "od: invalid output address radix") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestRunInvalidType(t *testing.T) {
	t.Parallel()
	_, errOut, err := runStdin(t, "ABC\n", "-t", "z9")
	if err == nil {
		t.Fatal("expected error for invalid type")
	}
	if !strings.Contains(errOut, "od: invalid type string") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestRunHelpAndVersion(t *testing.T) {
	t.Parallel()
	out, _, err := runStdin(t, "", "--help")
	if err != nil {
		t.Fatalf("--help error = %v", err)
	}
	if !strings.Contains(out, "Usage: od") {
		t.Errorf("--help out = %q", out)
	}

	out, _, err = runStdin(t, "", "--version")
	if err != nil {
		t.Fatalf("--version error = %v", err)
	}
	if !strings.Contains(out, "od (mimixbox)") {
		t.Errorf("--version out = %q", out)
	}
}

func TestHelpSections(t *testing.T) {
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := od.New().Run(context.Background(), io, []string{"--help"}); err != nil {
		t.Fatalf("Run --help error = %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "Examples:") {
		t.Errorf("--help missing Examples section:\n%s", got)
	}
	if !strings.Contains(got, "Exit status:") {
		t.Errorf("--help missing Exit status section:\n%s", got)
	}
}
