package lzw_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/archival/lzw"
)

func roundTrip(t *testing.T, data []byte) {
	t.Helper()
	var z bytes.Buffer
	if err := lzw.Compress(bytes.NewReader(data), &z); err != nil {
		t.Fatalf("Compress error = %v", err)
	}
	// The stream must start with the .Z magic and block-mode max-bits byte.
	b := z.Bytes()
	if len(b) < 3 || b[0] != 0x1f || b[1] != 0x9d {
		t.Fatalf("missing .Z magic: % x", b[:min(3, len(b))])
	}
	var out bytes.Buffer
	if err := lzw.Decompress(bytes.NewReader(b), &out); err != nil {
		t.Fatalf("Decompress error = %v", err)
	}
	if !bytes.Equal(out.Bytes(), data) {
		t.Errorf("round trip mismatch: got %d bytes, want %d", out.Len(), len(data))
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func TestRoundTrip(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		data []byte
	}{
		{"empty", []byte{}},
		{"single", []byte("a")},
		{"repeated", []byte(strings.Repeat("ab", 1000))},
		{"text", []byte("the quick brown fox jumps over the lazy dog\n")},
		{"all bytes", func() []byte {
			b := make([]byte, 1024)
			for i := range b {
				b[i] = byte(i % 256)
			}
			return b
		}()},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			roundTrip(t, tt.data)
		})
	}
}

func TestLargeInputGrowsWidth(t *testing.T) {
	t.Parallel()
	// A large, varied input forces the code width to grow past 9 bits and
	// eventually issue CLEAR codes, exercising the block-alignment logic.
	var b bytes.Buffer
	for i := 0; i < 20000; i++ {
		b.WriteString("line ")
		b.WriteByte(byte('a' + i%26))
		b.WriteByte('\n')
	}
	roundTrip(t, b.Bytes())
}

// failWriter returns an error after allowing n bytes through, to exercise the
// codec's write-error branches.
type failWriter struct {
	n int
}

func (f *failWriter) Write(p []byte) (int, error) {
	if f.n <= 0 {
		return 0, errFail
	}
	if len(p) > f.n {
		w := f.n
		f.n = 0
		return w, errFail
	}
	f.n -= len(p)
	return len(p), nil
}

type fixedErr struct{}

func (fixedErr) Error() string { return "boom" }

var errFail = fixedErr{}

func TestCompressWriteErrors(t *testing.T) {
	t.Parallel()
	// A large, varied input compresses to well over bufio's 4 KiB buffer, so a
	// limited writer surfaces errors at the inner block writes, not just at the
	// final flush.
	var buf bytes.Buffer
	for i := 0; i < 60000; i++ {
		buf.WriteString("entry ")
		buf.WriteByte(byte('a' + i%26))
		buf.WriteByte(byte('0' + i%10))
		buf.WriteByte('\n')
	}
	data := buf.Bytes()
	for _, budget := range []int{0, 8, 5000, 9000} {
		budget := budget
		t.Run("budget", func(t *testing.T) {
			t.Parallel()
			if err := lzw.Compress(bytes.NewReader(data), &failWriter{n: budget}); err == nil {
				t.Errorf("expected a write error at budget %d", budget)
			}
		})
	}
}

func TestDecompressWriteError(t *testing.T) {
	t.Parallel()
	var z bytes.Buffer
	if err := lzw.Compress(strings.NewReader(strings.Repeat("abc", 500)), &z); err != nil {
		t.Fatal(err)
	}
	if err := lzw.Decompress(bytes.NewReader(z.Bytes()), &failWriter{n: 4}); err == nil {
		t.Error("expected a write error during decompression")
	}
}

func TestDecompressRejectsNonZ(t *testing.T) {
	t.Parallel()
	var out bytes.Buffer
	if err := lzw.Decompress(strings.NewReader("not compressed"), &out); err == nil {
		t.Error("expected error for non-.Z input")
	}
}

func TestDecompressShortHeader(t *testing.T) {
	t.Parallel()
	var out bytes.Buffer
	if err := lzw.Decompress(strings.NewReader("\x1f"), &out); err == nil {
		t.Error("expected error for truncated header")
	}
}
