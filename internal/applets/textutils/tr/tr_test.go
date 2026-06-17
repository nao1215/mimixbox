package tr_test

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/textutils/tr"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, stdin string, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(stdin), Out: out, Err: errBuf}
	err := tr.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

// runBytes drives tr with raw byte input/output so non-UTF-8 data can be
// exercised without lossy string conversion.
func runBytes(t *testing.T, stdin []byte, args ...string) ([]byte, error) {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: bytes.NewReader(stdin), Out: out, Err: &bytes.Buffer{}}
	err := tr.New().Run(context.Background(), io, args)
	return out.Bytes(), err
}

// dripReader returns its data one byte per Read call, forcing tr to handle
// multi-byte UTF-8 sequences and squeeze runs that straddle read boundaries.
type dripReader struct {
	data []byte
	pos  int
}

func (d *dripReader) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	if d.pos >= len(d.data) {
		return 0, io.EOF
	}
	p[0] = d.data[d.pos]
	d.pos++
	return 1, nil
}

func TestStreamingMatchesWholeInput(t *testing.T) {
	t.Parallel()
	// Streaming the input one byte at a time must produce the same output as a
	// single read, across multi-byte UTF-8 boundaries and squeeze runs (#952).
	cases := []struct {
		in   string
		args []string
	}{
		{"あああいいうう\n", []string{"-s", "あいう"}},   // squeeze multibyte across chunks
		{"héllo wörld\n", []string{"a-z", "A-Z"}},        // translate ASCII, keep multibyte
		{"aaabbbccc\n", []string{"-s", "a-z"}},           // squeeze ASCII run
		{"abcあ123\n", []string{"-d", "[:digit:]"}},      // delete with multibyte present
		{"x\xff\xffy\n", []string{"-s", "\\377"}},        // squeeze raw bytes
	}
	for _, tc := range cases {
		out := &bytes.Buffer{}
		io1 := command.IO{In: strings.NewReader(tc.in), Out: out, Err: &bytes.Buffer{}}
		if err := tr.New().Run(context.Background(), io1, tc.args); err != nil {
			t.Fatalf("whole run error = %v", err)
		}
		drip := &bytes.Buffer{}
		io2 := command.IO{In: &dripReader{data: []byte(tc.in)}, Out: drip, Err: &bytes.Buffer{}}
		if err := tr.New().Run(context.Background(), io2, tc.args); err != nil {
			t.Fatalf("drip run error = %v", err)
		}
		if out.String() != drip.String() {
			t.Errorf("args %v: drip=%q whole=%q", tc.args, drip.String(), out.String())
		}
	}
}

func TestTranslatesRawByteMatchingOctalSet(t *testing.T) {
	// tr '\377' X on a lone 0xFF byte must translate it like GNU tr (byte
	// oriented), not corrupt it into the UTF-8 replacement character (issue
	// #953).
	got, err := runBytes(t, []byte{0xFF, '\n'}, `\377`, "X")
	if err != nil {
		t.Fatal(err)
	}
	want := []byte{'X', '\n'}
	if !bytes.Equal(got, want) {
		t.Errorf("tr output = % x, want % x", got, want)
	}
}

func TestPreservesUntranslatedBinaryBytes(t *testing.T) {
	// tr A Z must leave unrelated non-UTF-8 bytes untouched (issue #953).
	got, err := runBytes(t, []byte{0xFF, 'A', 0xFE, '\n'}, "A", "Z")
	if err != nil {
		t.Fatal(err)
	}
	want := []byte{0xFF, 'Z', 0xFE, '\n'}
	if !bytes.Equal(got, want) {
		t.Errorf("tr output = % x, want % x", got, want)
	}
}

func TestDeletesRawBinaryBytes(t *testing.T) {
	// Deleting a raw byte by its octal value must work on binary input.
	got, err := runBytes(t, []byte{0xFF, 'a', 0xFF, 'b'}, "-d", `\377`)
	if err != nil {
		t.Fatal(err)
	}
	want := []byte{'a', 'b'}
	if !bytes.Equal(got, want) {
		t.Errorf("tr output = % x, want % x", got, want)
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
		{"translate range", "hello\n", []string{"a-z", "A-Z"}, "HELLO\n"},
		{"translate literal", "hello", []string{"el", "ip"}, "hippo"},
		{"delete set", "hello world\n", []string{"-d", "lo"}, "he wrd\n"},
		{"delete range", "abc123\n", []string{"-d", "0-9"}, "abc\n"},
		{"squeeze", "aaabbbccc\n", []string{"-s", "a-z"}, "abc\n"},
		{"squeeze specific", "aaabbbccc", []string{"-s", "a"}, "abbbccc"},
		{"complement delete", "abc123\n", []string{"-cd", "0-9"}, "123"},
		{"complement translate", "abc123", []string{"-c", "0-9", "x"}, "xxx123"},
		{"digit class delete", "a1b2c3\n", []string{"-d", "[:digit:]"}, "abc\n"},
		{"upper class", "hello", []string{"[:lower:]", "[:upper:]"}, "HELLO"},
		{"space class squeeze", "a   b  c", []string{"-s", "[:space:]"}, "a b c"},
		{"newline escape", "a\nb\n", []string{"\\n", "_"}, "a_b_"},
		{"octal escape", "aXb", []string{"\\130", "_"}, "a_b"},
		{"translate pad shorter set2", "abcd", []string{"a-d", "x"}, "xxxx"},
		{"delete then squeeze", "aabbccdd", []string{"-ds", "a", "b"}, "bccdd"},
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

func TestRunMissingOperand(t *testing.T) {
	t.Parallel()
	out, errOut, err := run(t, "hello\n")
	if err == nil {
		t.Fatal("expected error for missing operand")
	}
	if out != "" {
		t.Errorf("out = %q, want empty", out)
	}
	if !strings.Contains(errOut, "tr: missing operand") {
		t.Errorf("stderr = %q, want missing operand message", errOut)
	}
}

func TestRunMissingSet2(t *testing.T) {
	t.Parallel()
	out, errOut, err := run(t, "hello\n", "a-z")
	if err == nil {
		t.Fatal("expected error for missing SET2 in translate mode")
	}
	if out != "" {
		t.Errorf("out = %q, want empty", out)
	}
	if !strings.Contains(errOut, "tr: missing operand after 'a-z'") {
		t.Errorf("stderr = %q, want missing operand after message", errOut)
	}
}

// TestRunTruncateSet1 checks --truncate-set1/-t: SET1 is cut to the length of
// SET2 before translating, so SET1 characters past SET2's length pass through
// unchanged instead of mapping to SET2's last rune.
func TestRunTruncateSet1(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		args []string
		in   string
		want string
	}{
		// SET1=abc SET2=xy: with -t, a->x, b->y, c is left unchanged.
		{"long flag", []string{"--truncate-set1", "abc", "xy"}, "abc\n", "xyc\n"},
		{"short flag", []string{"-t", "abc", "xy"}, "abc\n", "xyc\n"},
		// Without -t the default GNU padding maps c to SET2's last rune (y).
		{"no truncate pads", []string{"abc", "xy"}, "abc\n", "xyy\n"},
		// Equal-length sets are unaffected by -t.
		{"equal length", []string{"-t", "abc", "xyz"}, "abc\n", "xyz\n"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out, errOut, err := run(t, tt.in, tt.args...)
			if err != nil {
				t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
			}
			if out != tt.want {
				t.Errorf("out = %q, want %q", out, tt.want)
			}
		})
	}
}

func TestRunHelp(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "", "--help")
	if err != nil {
		t.Fatalf("--help error = %v", err)
	}
	if !strings.Contains(out, "Usage: tr") {
		t.Errorf("--help out = %q", out)
	}
	if !strings.Contains(out, "Examples:") || !strings.Contains(out, "Exit status:") {
		t.Errorf("--help missing structured sections:\n%s", out)
	}
}
