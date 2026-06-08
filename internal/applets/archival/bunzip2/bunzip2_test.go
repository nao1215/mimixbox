package bunzip2_test

import (
	"bytes"
	"context"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/archival/bunzip2"
	"github.com/nao1215/mimixbox/internal/command"
)

// bz2Hex is bzip2 -c of the literal "hello bunzip2\n". Go's standard library
// has a bzip2 decompressor but no compressor, so the fixture is precomputed.
const (
	bz2Hex  = "425a6839314159265359d77cc601000002d9800010400010001265c21020002200034201a005cd7a817a13b878bb9229c28486bbe63008"
	bz2Text = "hello bunzip2\n"
)

func bz2(t *testing.T) []byte {
	t.Helper()
	b, err := hex.DecodeString(bz2Hex)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func run(t *testing.T, stdin []byte, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: bytes.NewReader(stdin), Out: out, Err: errBuf}
	err := bunzip2.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func writeBz2(t *testing.T, path string) {
	t.Helper()
	if err := os.WriteFile(path, bz2(t), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := bunzip2.New()
	if got := c.Name(); got != "bunzip2" {
		t.Errorf("Name() = %q", got)
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis empty")
	}
}

func TestStdin(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, bz2(t))
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if out != bz2Text {
		t.Errorf("out = %q, want %q", out, bz2Text)
	}
}

func TestStdout(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "f.bz2")
	writeBz2(t, src)
	out, _, err := run(t, nil, "-c", src)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if out != bz2Text {
		t.Errorf("out = %q", out)
	}
	if _, statErr := os.Stat(src); statErr != nil {
		t.Errorf("-c should keep input: %v", statErr)
	}
}

func TestFileReplace(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "doc.bz2")
	writeBz2(t, src)
	if _, errOut, err := run(t, nil, src); err != nil {
		t.Fatalf("err = %v (stderr=%q)", err, errOut)
	}
	got, err := os.ReadFile(filepath.Join(dir, "doc"))
	if err != nil {
		t.Fatalf("expected decompressed file: %v", err)
	}
	if string(got) != bz2Text {
		t.Errorf("decompressed = %q", got)
	}
	if _, statErr := os.Stat(src); statErr == nil {
		t.Error("input .bz2 should be removed without -k")
	}
}

func TestKeep(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "k.bz2")
	writeBz2(t, src)
	if _, _, err := run(t, nil, "-k", src); err != nil {
		t.Fatalf("err = %v", err)
	}
	if _, statErr := os.Stat(src); statErr != nil {
		t.Errorf("-k should keep input: %v", statErr)
	}
}

func TestUnknownSuffix(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "noext")
	writeBz2(t, src)
	_, errOut, err := run(t, nil, src)
	if err == nil {
		t.Error("expected error for unknown suffix")
	}
	if !strings.Contains(errOut, "unknown suffix") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestExistingOutputNeedsForce(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "e.bz2")
	writeBz2(t, src)
	if err := os.WriteFile(filepath.Join(dir, "e"), []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, errOut, err := run(t, nil, src)
	if err == nil {
		t.Error("expected error when output exists without -f")
	}
	if !strings.Contains(errOut, "already exists") {
		t.Errorf("stderr = %q", errOut)
	}
	if _, _, err := run(t, nil, "-f", src); err != nil {
		t.Fatalf("-f err = %v", err)
	}
}

func TestHelp(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, nil, "--help")
	if err != nil {
		t.Fatalf("help err = %v", err)
	}
	if !strings.Contains(out, "Usage: bunzip2") {
		t.Errorf("help = %q", out)
	}
}
