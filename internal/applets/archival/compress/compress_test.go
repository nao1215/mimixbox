package compress_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/archival/compress"
	"github.com/nao1215/mimixbox/internal/applets/archival/lzw"
	"github.com/nao1215/mimixbox/internal/applets/archival/uncompress"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, stdin []byte, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: bytes.NewReader(stdin), Out: out, Err: errBuf}
	err := compress.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := compress.New()
	if got := c.Name(); got != "compress" {
		t.Errorf("Name() = %q", got)
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis empty")
	}
}

func TestStdinStdout(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, []byte("hello compress hello compress"))
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	// Round-trip back through the decoder.
	var dec bytes.Buffer
	if derr := lzw.Decompress(strings.NewReader(out), &dec); derr != nil {
		t.Fatalf("decompress err = %v", derr)
	}
	if dec.String() != "hello compress hello compress" {
		t.Errorf("round trip = %q", dec.String())
	}
}

func TestFileToZ(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	f := filepath.Join(dir, "data.txt")
	content := strings.Repeat("compress me ", 100)
	if err := os.WriteFile(f, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, errOut, err := run(t, nil, f); err != nil {
		t.Fatalf("err = %v (stderr=%q)", err, errOut)
	}
	// FILE becomes FILE.Z and the original is removed.
	if _, statErr := os.Stat(f); statErr == nil {
		t.Error("original should be removed without -k")
	}
	zf := f + ".Z"
	zdata, err := os.ReadFile(zf)
	if err != nil {
		t.Fatalf("expected %s: %v", zf, err)
	}
	var dec bytes.Buffer
	if derr := lzw.Decompress(bytes.NewReader(zdata), &dec); derr != nil {
		t.Fatalf("decompress err = %v", derr)
	}
	if dec.String() != content {
		t.Error("decompressed content mismatch")
	}
}

func TestKeep(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	f := filepath.Join(dir, "k.txt")
	if err := os.WriteFile(f, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := run(t, nil, "-k", f); err != nil {
		t.Fatalf("err = %v", err)
	}
	if _, statErr := os.Stat(f); statErr != nil {
		t.Errorf("-k should keep input: %v", statErr)
	}
}

func TestStdoutFlag(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	f := filepath.Join(dir, "c.txt")
	if err := os.WriteFile(f, []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}
	out, _, err := run(t, nil, "-c", f)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if !strings.HasPrefix(out, "\x1f\x9d") {
		t.Error("-c should write a .Z stream to stdout")
	}
	if _, statErr := os.Stat(f); statErr != nil {
		t.Error("-c should keep the input")
	}
}

func TestAppletRoundTripWithUncompress(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	f := filepath.Join(dir, "rt.txt")
	content := strings.Repeat("round trip ", 50)
	if err := os.WriteFile(f, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := run(t, nil, f); err != nil {
		t.Fatalf("compress err = %v", err)
	}
	// Now decompress with the uncompress applet.
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := uncompress.New().Run(context.Background(), io, []string{f + ".Z"}); err != nil {
		t.Fatalf("uncompress err = %v", err)
	}
	got, err := os.ReadFile(f)
	if err != nil {
		t.Fatalf("expected restored file: %v", err)
	}
	if string(got) != content {
		t.Error("applet round trip mismatch")
	}
}

func TestErrors(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	t.Run("missing file", func(t *testing.T) {
		t.Parallel()
		_, errOut, err := run(t, nil, filepath.Join(dir, "nope"))
		if err == nil {
			t.Error("expected error for missing file")
		}
		if !strings.Contains(errOut, "compress:") {
			t.Errorf("stderr = %q", errOut)
		}
	})
	t.Run("already .Z", func(t *testing.T) {
		t.Parallel()
		zf := filepath.Join(dir, "already.Z")
		if err := os.WriteFile(zf, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		_, errOut, err := run(t, nil, zf)
		if err == nil {
			t.Error("expected error for .Z input")
		}
		if !strings.Contains(errOut, ".Z suffix") {
			t.Errorf("stderr = %q", errOut)
		}
	})
}

func TestHelp(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, nil, "--help")
	if err != nil {
		t.Fatalf("help err = %v", err)
	}
	if !strings.Contains(out, "Usage: compress") {
		t.Errorf("help = %q", out)
	}
	for _, want := range []string{"Examples:", "Exit status:"} {
		if !strings.Contains(out, want) {
			t.Errorf("--help missing %q section:\n%s", want, out)
		}
	}
}
