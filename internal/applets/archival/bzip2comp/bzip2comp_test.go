package bzip2comp

import (
	"bytes"
	"compress/bzip2"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func TestRoundTripStream(t *testing.T) {
	cases := []struct {
		name string
		data []byte
	}{
		{"empty", []byte{}},
		{"short", []byte("hello bzip2")},
		{"text", bytes.Repeat([]byte("the quick brown fox jumps over the lazy dog\n"), 200)},
		{"binary", func() []byte {
			b := make([]byte, 10000)
			for i := range b {
				b[i] = byte(i * 13)
			}
			return b
		}()},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var comp bytes.Buffer
			if err := transform(bytes.NewReader(tc.data), &comp, false); err != nil {
				t.Fatalf("compress: %v", err)
			}
			// Decompressable by the standard library too (interop check).
			std, err := io.ReadAll(bzip2.NewReader(bytes.NewReader(comp.Bytes())))
			if err != nil {
				t.Fatalf("stdlib decompress: %v", err)
			}
			if !bytes.Equal(std, tc.data) {
				t.Fatalf("stdlib round trip mismatch")
			}

			var out bytes.Buffer
			if err := transform(bytes.NewReader(comp.Bytes()), &out, true); err != nil {
				t.Fatalf("decompress: %v", err)
			}
			if !bytes.Equal(out.Bytes(), tc.data) {
				t.Fatalf("round trip mismatch")
			}
		})
	}
}

func TestRunStreamRoundTrip(t *testing.T) {
	data := []byte("bzip2 stream round trip payload\n")
	var comp bytes.Buffer
	if err := New().Run(context.Background(), command.IO{In: bytes.NewReader(data), Out: &comp, Err: &bytes.Buffer{}}, nil); err != nil {
		t.Fatalf("compress run: %v", err)
	}
	var out bytes.Buffer
	if err := New().Run(context.Background(), command.IO{In: bytes.NewReader(comp.Bytes()), Out: &out, Err: &bytes.Buffer{}}, []string{"-d"}); err != nil {
		t.Fatalf("decompress run: %v", err)
	}
	if !bytes.Equal(out.Bytes(), data) {
		t.Fatalf("round trip mismatch")
	}
}

func TestTestOption(t *testing.T) {
	data := []byte("integrity payload")
	var comp bytes.Buffer
	if err := transform(bytes.NewReader(data), &comp, false); err != nil {
		t.Fatal(err)
	}
	if err := New().Run(context.Background(), command.IO{In: bytes.NewReader(comp.Bytes()), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}, []string{"-t"}); err != nil {
		t.Fatalf("-t on valid stream failed: %v", err)
	}
	bad := []byte("this is not bzip2 at all")
	if err := New().Run(context.Background(), command.IO{In: bytes.NewReader(bad), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}, []string{"-t"}); err == nil {
		t.Fatalf("-t on garbage should fail")
	}
}

func TestFileKeepAndRemove(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "data.txt")
	data := []byte(strings.Repeat("bzip2 file test\n", 80))
	if err := os.WriteFile(src, data, 0o644); err != nil {
		t.Fatal(err)
	}

	// Compress without -k removes the original.
	if err := New().Run(context.Background(), command.IO{In: nil, Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}, []string{src}); err != nil {
		t.Fatalf("compress: %v", err)
	}
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Fatalf("original should have been removed")
	}
	if _, err := os.Stat(src + ".bz2"); err != nil {
		t.Fatalf("compressed file missing: %v", err)
	}

	// Decompress to stdout with -c keeps the .bz2.
	var out bytes.Buffer
	if err := New().Run(context.Background(), command.IO{In: nil, Out: &out, Err: &bytes.Buffer{}}, []string{"-dc", src + ".bz2"}); err != nil {
		t.Fatalf("decompress -c: %v", err)
	}
	if !bytes.Equal(out.Bytes(), data) {
		t.Fatalf("decompressed output mismatch")
	}
	if _, err := os.Stat(src + ".bz2"); err != nil {
		t.Fatalf("compressed file should remain after -c")
	}
}

func TestOutputName(t *testing.T) {
	cases := []struct {
		in, want   string
		decompress bool
		wantErr    bool
	}{
		{"file", "file.bz2", false, false},
		{"file.bz2", "file", true, false},
		{"x.tbz2", "x.tar", true, false},
		{"x.tbz", "x.tar", true, false},
		{"file.txt", "", true, true},
	}
	for _, tc := range cases {
		got, err := outputName(tc.in, tc.decompress)
		if tc.wantErr {
			if err == nil {
				t.Errorf("outputName(%q,%v): expected error", tc.in, tc.decompress)
			}
			continue
		}
		if err != nil || got != tc.want {
			t.Errorf("outputName(%q,%v) = %q,%v; want %q", tc.in, tc.decompress, got, err, tc.want)
		}
	}
}
