package od_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/shellutils/od"
	"github.com/nao1215/mimixbox/internal/command"
)

// TestRunFormatTypes drives the -t types whose unit renderers were previously
// uncovered: signed (d1/d2), unsigned (u1/u2) and the named-character form (a).
// Expected outputs were checked against GNU od.
func TestRunFormatTypes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		stdin string
		args  []string
		want  string
	}{
		{
			// 0xFF is -1 as int8; "A" is 65.
			name:  "signed bytes d1",
			stdin: "\xffA",
			args:  []string{"-A", "n", "-t", "d1"},
			want:  "   -1   65\n",
		},
		{
			// little-endian 0xFFFF = -1 as int16.
			name:  "signed words d2",
			stdin: "\xff\xff",
			args:  []string{"-A", "n", "-t", "d2"},
			want:  "     -1\n",
		},
		{
			name:  "unsigned bytes u1",
			stdin: "\x00\xff",
			args:  []string{"-A", "n", "-t", "u1"},
			want:  "   0 255\n",
		},
		{
			// little-endian 0xFF00 = 255.
			name:  "unsigned words u2 shortcut -d",
			stdin: "\xff\x00",
			args:  []string{"-A", "n", "-d"},
			want:  "   255\n",
		},
		{
			// -t a names control characters and prints printable bytes as is.
			name:  "named characters a",
			stdin: "\x00 A\x7f\n",
			args:  []string{"-A", "n", "-t", "a"},
			want:  " nul  sp   A del  nl\n",
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

// TestRunCharNonPrintable exercises charUnit's octal fallback for a byte that
// is neither a known C escape nor printable.
func TestRunCharNonPrintable(t *testing.T) {
	t.Parallel()
	// 0xff has no C escape and is not printable -> rendered as octal "377".
	out, _, err := runStdin(t, "\xff", "-A", "n", "-c")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != " 377\n" {
		t.Errorf("out = %q, want octal fallback", out)
	}
}

// TestRunMultipleFilesConcatenated checks that several operands are read and
// dumped as one stream, and that the address counter spans both files.
func TestRunMultipleFilesConcatenated(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	a := filepath.Join(dir, "a.bin")
	b := filepath.Join(dir, "b.bin")
	if err := os.WriteFile(a, []byte("AB"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(b, []byte("CD"), 0o600); err != nil {
		t.Fatal(err)
	}
	out, _, err := runStdin(t, "", "-A", "n", "-t", "x1", a, b)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != " 41 42 43 44\n" {
		t.Errorf("out = %q", out)
	}
}

// TestRunMultipleMissingFilesKeepsFirstError exercises keep(): two missing
// operands still produce a single (non-nil) failure and both are reported.
func TestRunMultipleMissingFilesKeepsFirstError(t *testing.T) {
	t.Parallel()
	_, errOut, err := runStdin(t, "", "/no/such/a", "/no/such/b")
	if err == nil {
		t.Fatal("expected error for missing files")
	}
	if !strings.Contains(errOut, "/no/such/a") || !strings.Contains(errOut, "/no/such/b") {
		t.Errorf("stderr = %q, want both missing files reported", errOut)
	}
}

// TestRunMixedGoodAndMissingFile checks that a readable file still dumps even
// when another operand is missing, and the exit code is still failure.
func TestRunMixedGoodAndMissingFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	good := filepath.Join(dir, "good.bin")
	if err := os.WriteFile(good, []byte("AB"), 0o600); err != nil {
		t.Fatal(err)
	}
	out, _, err := runStdin(t, "", "-A", "n", "-t", "x1", good, "/no/such/file")
	if err == nil {
		t.Fatal("expected failure because of the missing file")
	}
	if !strings.Contains(out, "41 42") {
		t.Errorf("good file should still be dumped, out = %q", out)
	}
}

// TestSynopsis covers the Synopsis accessor.
func TestSynopsis(t *testing.T) {
	t.Parallel()
	c := od.New()
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
	if c.Name() != "od" {
		t.Errorf("Name() = %q", c.Name())
	}
}

// TestRunPadAndWideValue makes sure a single trailing partial unit is rendered
// without panicking (little-endian zero-extends the missing high byte).
func TestRunPadAndWideValue(t *testing.T) {
	t.Parallel()
	// Odd number of bytes with a 2-byte type: the last word is a partial unit.
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader("A"), Out: out, Err: &bytes.Buffer{}}
	if err := od.New().Run(context.Background(), io, []string{"-A", "n", "-t", "x2"}); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	// "A" is 0x41; the trailing partial word renders the single byte as "41",
	// right-justified in the 2-byte field width of 4.
	if out.String() != "   41\n" {
		t.Errorf("out = %q", out.String())
	}
}
