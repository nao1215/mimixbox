package lzopcomp

import (
	"bytes"
	"context"
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
		{"short", []byte("hello")},
		{"text", bytes.Repeat([]byte("the quick brown fox jumps over the lazy dog\n"), 100)},
		{"binary", func() []byte {
			b := make([]byte, 9000)
			for i := range b {
				b[i] = byte(i * 7)
			}
			return b
		}()},
		{"multiblock", bytes.Repeat([]byte("A"), maxBlockSize*2+123)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var comp bytes.Buffer
			if err := compressStream(bytes.NewReader(tc.data), &comp); err != nil {
				t.Fatalf("compress: %v", err)
			}
			var out bytes.Buffer
			if err := decompressStream(bytes.NewReader(comp.Bytes()), &out); err != nil {
				t.Fatalf("decompress: %v", err)
			}
			if !bytes.Equal(out.Bytes(), tc.data) {
				t.Fatalf("round trip mismatch: got %d bytes, want %d", out.Len(), len(tc.data))
			}
		})
	}
}

func TestDecompressRejectsCorrupt(t *testing.T) {
	cases := map[string][]byte{
		"empty":        {},
		"bad magic":    []byte("not an lzo file at all....."),
		"truncated":    append(append([]byte{}, lzopMagic...), 0x12, 0x34),
		"short header": lzopMagic,
	}
	for name, in := range cases {
		t.Run(name, func(t *testing.T) {
			if err := decompressStream(bytes.NewReader(in), &bytes.Buffer{}); err == nil {
				t.Fatalf("expected error for %s", name)
			}
		})
	}
}

func TestRunStreamCompressDecompress(t *testing.T) {
	data := []byte("mimixbox lzop stream test payload\n")

	var comp bytes.Buffer
	c := NewLzop()
	if err := c.Run(context.Background(), command.IO{In: bytes.NewReader(data), Out: &comp, Err: &bytes.Buffer{}}, nil); err != nil {
		t.Fatalf("compress run: %v", err)
	}

	var out bytes.Buffer
	d := NewUnlzop()
	if err := d.Run(context.Background(), command.IO{In: bytes.NewReader(comp.Bytes()), Out: &out, Err: &bytes.Buffer{}}, nil); err != nil {
		t.Fatalf("decompress run: %v", err)
	}
	if !bytes.Equal(out.Bytes(), data) {
		t.Fatalf("round trip mismatch")
	}
}

func TestLzopcatMatchesUnlzop(t *testing.T) {
	data := []byte("lzopcat must equal unlzop -c output")
	var comp bytes.Buffer
	if err := compressStream(bytes.NewReader(data), &comp); err != nil {
		t.Fatal(err)
	}

	var catOut, unOut bytes.Buffer
	if err := NewLzopcat().Run(context.Background(), command.IO{In: bytes.NewReader(comp.Bytes()), Out: &catOut, Err: &bytes.Buffer{}}, nil); err != nil {
		t.Fatal(err)
	}
	if err := NewUnlzop().Run(context.Background(), command.IO{In: bytes.NewReader(comp.Bytes()), Out: &unOut, Err: &bytes.Buffer{}}, []string{"-c"}); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(catOut.Bytes(), unOut.Bytes()) {
		t.Fatalf("lzopcat output != unlzop -c output")
	}
}

func TestFileRoundTripAndKeep(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "data.txt")
	data := []byte(strings.Repeat("file round trip\n", 50))
	if err := os.WriteFile(src, data, 0o644); err != nil {
		t.Fatal(err)
	}

	// Compress keeping the original.
	if err := NewLzop().Run(context.Background(), command.IO{In: nil, Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}, []string{"-k", src}); err != nil {
		t.Fatalf("compress file: %v", err)
	}
	if _, err := os.Stat(src); err != nil {
		t.Fatalf("original removed despite -k: %v", err)
	}
	if _, err := os.Stat(src + ".lzo"); err != nil {
		t.Fatalf("compressed file missing: %v", err)
	}

	// Remove original, then decompress.
	if err := os.Remove(src); err != nil {
		t.Fatal(err)
	}
	if err := NewUnlzop().Run(context.Background(), command.IO{In: nil, Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}, []string{src + ".lzo"}); err != nil {
		t.Fatalf("decompress file: %v", err)
	}
	got, err := os.ReadFile(src)
	if err != nil {
		t.Fatalf("decompressed file missing: %v", err)
	}
	if !bytes.Equal(got, data) {
		t.Fatalf("file round trip mismatch")
	}
}

func TestTest(t *testing.T) {
	data := []byte("integrity test data for lzop")
	var comp bytes.Buffer
	if err := compressStream(bytes.NewReader(data), &comp); err != nil {
		t.Fatal(err)
	}
	if err := NewLzop().Run(context.Background(), command.IO{In: bytes.NewReader(comp.Bytes()), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}, []string{"-t"}); err != nil {
		t.Fatalf("-t on valid stream failed: %v", err)
	}
	corrupt := append([]byte{}, comp.Bytes()...)
	corrupt[len(corrupt)-1] ^= 0xFF
	if err := NewLzop().Run(context.Background(), command.IO{In: bytes.NewReader(corrupt), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}, []string{"-t"}); err == nil {
		t.Fatalf("-t on corrupt stream should fail")
	}
}

func TestNames(t *testing.T) {
	if NewLzop().Name() != "lzop" || NewUnlzop().Name() != "unlzop" || NewLzopcat().Name() != "lzopcat" {
		t.Fatal("unexpected applet names")
	}
}

// TestHelpSections locks in the self-describing --help output for every lzopcomp
// applet (GitHub issues #652/#653/#656 Examples, #701/#702/#704 purpose
// paragraph, plus #701/#704 Notes): each command must render a purpose
// paragraph, an Examples section, and an Exit status section.
func TestHelpSections(t *testing.T) {
	t.Parallel()
	ctors := []func() *Command{NewLzop, NewUnlzop, NewLzopcat}
	for _, ctor := range ctors {
		c := ctor()
		t.Run(c.Name(), func(t *testing.T) {
			t.Parallel()
			out := &bytes.Buffer{}
			io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
			if err := c.Run(context.Background(), io, []string{"--help"}); err != nil {
				t.Fatalf("%s --help err = %v", c.Name(), err)
			}
			help := out.String()
			for _, want := range []string{"Usage: " + c.Name(), "Examples:", "Exit status:"} {
				if !strings.Contains(help, want) {
					t.Errorf("%s --help missing %q\n%s", c.Name(), want, help)
				}
			}
			desc := help[strings.Index(help, "\n")+1:]
			if before, _, _ := strings.Cut(desc, "\nOptions:"); strings.TrimSpace(before) == "" {
				t.Errorf("%s --help has no purpose paragraph\n%s", c.Name(), help)
			}
		})
	}
}
