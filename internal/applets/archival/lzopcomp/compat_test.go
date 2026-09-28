package lzopcomp

import (
	"bytes"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// compatInputs covers the shapes an LZO1X codec treats differently: no input,
// inputs shorter than the minimum match, pure literals (incompressible, stored
// verbatim by the container), long runs (long M3/M4 matches with zero-extended
// lengths), text (short and mid-distance matches) and a stream that spans
// several 256 KiB container blocks.
func compatInputs(t *testing.T) map[string][]byte {
	t.Helper()
	rng := rand.New(rand.NewSource(1))
	random := make([]byte, 100_000)
	rng.Read(random)

	var mixed bytes.Buffer
	for mixed.Len() < maxBlockSize*2+4096 {
		switch rng.Intn(3) {
		case 0:
			chunk := make([]byte, rng.Intn(2000))
			rng.Read(chunk)
			mixed.Write(chunk)
		case 1:
			mixed.Write(bytes.Repeat([]byte{byte(rng.Intn(256))}, rng.Intn(70_000)))
		default:
			mixed.Write(bytes.Repeat([]byte("mimixbox lzop compat "), rng.Intn(500)))
		}
	}

	return map[string][]byte{
		"empty":          {},
		"one byte":       {'x'},
		"three bytes":    []byte("abc"),
		"random":         random,
		"zeros 1MiB":     make([]byte, 1<<20),
		"text":           bytes.Repeat([]byte("the quick brown fox jumps over the lazy dog\n"), 5000),
		"mixed 3 blocks": mixed.Bytes(),
	}
}

// requireLzop returns the path of the upstream lzop, or skips when it is not
// installed. CI installs it for the unit tests (see unit_test.yml) so these
// comparisons run on every pull request.
func requireLzop(t *testing.T) string {
	t.Helper()
	path, err := exec.LookPath("lzop")
	if err != nil {
		t.Skip("upstream lzop is not installed")
	}
	return path
}

func TestUpstreamLzopDecompressesWhatLzopWrites(t *testing.T) {
	t.Parallel()
	lzop := requireLzop(t)
	for name, data := range compatInputs(t) {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			var comp bytes.Buffer
			if err := compressStream(bytes.NewReader(data), &comp); err != nil {
				t.Fatalf("compress: %v", err)
			}
			file := filepath.Join(t.TempDir(), "in.lzo")
			if err := os.WriteFile(file, comp.Bytes(), 0o600); err != nil {
				t.Fatal(err)
			}
			var stderr bytes.Buffer
			cmd := exec.Command(lzop, "-d", "-c", file)
			cmd.Stderr = &stderr
			got, err := cmd.Output()
			if err != nil {
				t.Fatalf("upstream lzop -d rejected our stream: %v: %s", err, stderr.String())
			}
			if !bytes.Equal(got, data) {
				t.Fatalf("upstream lzop decoded %d bytes, want %d identical bytes", len(got), len(data))
			}
		})
	}
}

func TestLzopDecompressesWhatUpstreamLzopWritesAtEveryLevel(t *testing.T) {
	t.Parallel()
	lzop := requireLzop(t)
	// -1 and -3 select LZO1X-1; -7 to -9 select LZO1X-999, whose encoder emits
	// the full range of match kinds the decoder has to accept.
	levels := []string{"-1", "-3", "-7", "-9"}
	for name, data := range compatInputs(t) {
		for _, level := range levels {
			t.Run(name+" "+level, func(t *testing.T) {
				t.Parallel()
				file := filepath.Join(t.TempDir(), "in")
				if err := os.WriteFile(file, data, 0o600); err != nil {
					t.Fatal(err)
				}
				var stderr bytes.Buffer
				cmd := exec.Command(lzop, level, "-c", file)
				cmd.Stderr = &stderr
				comp, err := cmd.Output()
				if err != nil {
					t.Fatalf("upstream lzop %s: %v: %s", level, err, stderr.String())
				}
				var out bytes.Buffer
				if err := decompressStream(bytes.NewReader(comp), &out); err != nil {
					t.Fatalf("decompress upstream %s stream: %v", level, err)
				}
				if !bytes.Equal(out.Bytes(), data) {
					t.Fatalf("decoded %d bytes, want %d identical bytes", out.Len(), len(data))
				}
			})
		}
	}
}

func TestCompressThenDecompressReturnsTheInputForRandomShapes(t *testing.T) {
	t.Parallel()
	rng := rand.New(rand.NewSource(2))
	for i := range 300 {
		// Draw from a small alphabet part of the time so matches of every
		// length and distance appear, not only literals.
		data := make([]byte, rng.Intn(3*maxBlockSize/2))
		alphabet := 1 + rng.Intn(256)
		for j := range data {
			data[j] = byte(rng.Intn(alphabet))
		}
		var comp, out bytes.Buffer
		if err := compressStream(bytes.NewReader(data), &comp); err != nil {
			t.Fatalf("case %d: compress: %v", i, err)
		}
		if err := decompressStream(bytes.NewReader(comp.Bytes()), &out); err != nil {
			t.Fatalf("case %d (len %d, alphabet %d): decompress: %v", i, len(data), alphabet, err)
		}
		if !bytes.Equal(out.Bytes(), data) {
			t.Fatalf("case %d (len %d, alphabet %d): round trip mismatch", i, len(data), alphabet)
		}
	}
}

func FuzzCompressThenDecompress(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte("a"))
	f.Add(bytes.Repeat([]byte("ab"), 1000))
	f.Fuzz(func(t *testing.T, data []byte) {
		var comp, out bytes.Buffer
		if err := compressStream(bytes.NewReader(data), &comp); err != nil {
			t.Fatalf("compress: %v", err)
		}
		if err := decompressStream(bytes.NewReader(comp.Bytes()), &out); err != nil {
			t.Fatalf("decompress: %v", err)
		}
		if !bytes.Equal(out.Bytes(), data) {
			t.Fatalf("round trip mismatch")
		}
	})
}

// FuzzDecompress feeds arbitrary bytes to the reader. A hostile .lzo file may
// only produce an error, never a panic; whatever it accepts must decode to data
// that survives another round trip.
func FuzzDecompress(f *testing.F) {
	for _, seed := range [][]byte{
		bytes.Repeat([]byte("seed "), 200),
		make([]byte, 5000),
	} {
		var comp bytes.Buffer
		if err := compressStream(bytes.NewReader(seed), &comp); err != nil {
			f.Fatal(err)
		}
		f.Add(comp.Bytes())
	}
	f.Add(lzopMagic)
	f.Fuzz(func(t *testing.T, in []byte) {
		var out bytes.Buffer
		if err := decompressStream(bytes.NewReader(in), &out); err != nil {
			return
		}
		var comp, again bytes.Buffer
		if err := compressStream(bytes.NewReader(out.Bytes()), &comp); err != nil {
			t.Fatalf("compress accepted output: %v", err)
		}
		if err := decompressStream(bytes.NewReader(comp.Bytes()), &again); err != nil {
			t.Fatalf("decompress re-compressed output: %v", err)
		}
		if !bytes.Equal(again.Bytes(), out.Bytes()) {
			t.Fatalf("accepted stream does not round trip")
		}
	})
}
