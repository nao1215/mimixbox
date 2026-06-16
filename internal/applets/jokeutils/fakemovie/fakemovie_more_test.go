package fakemovie_test

import (
	"image"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRunJpegOutput covers writeImage's jpeg encoding branch (the non-.png
// path) by requesting a .jpg output file.
func TestRunJpegOutput(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	in := filepath.Join(dir, "in.png")
	out := filepath.Join(dir, "out.jpg")
	newPNG(t, in, 80, 80)

	if _, errOut, err := run(t, "-o", out, in); err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}
	f, err := os.Open(out)
	if err != nil {
		t.Fatalf("jpeg output not produced: %v", err)
	}
	defer func() { _ = f.Close() }()
	_, format, err := image.DecodeConfig(f)
	if err != nil {
		t.Fatalf("output is not a decodable image: %v", err)
	}
	if format != "jpeg" {
		t.Errorf("format = %q, want jpeg", format)
	}
}

// TestRunUnwritableOutput covers writeImage's create-error branch: an output
// path inside a nonexistent directory cannot be created.
func TestRunUnwritableOutput(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	in := filepath.Join(dir, "in.png")
	newPNG(t, in, 60, 60)
	out := filepath.Join(dir, "no_such_subdir", "out.png")

	_, errOut, err := run(t, "-o", out, in)
	if err == nil {
		t.Fatal("expected error writing to a nonexistent directory")
	}
	if !strings.Contains(errOut, "fakemovie:") {
		t.Errorf("stderr = %q, want fakemovie error prefix", errOut)
	}
}

// TestRunCorruptImage covers openImage's decode-error branch: a file with a png
// extension but non-image content fails to decode.
func TestRunCorruptImage(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	in := filepath.Join(dir, "broken.png")
	if err := os.WriteFile(in, []byte("definitely not a png"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, errOut, err := run(t, in)
	if err == nil {
		t.Fatal("expected error for a corrupt image")
	}
	if !strings.Contains(errOut, "fakemovie:") {
		t.Errorf("stderr = %q, want fakemovie error prefix", errOut)
	}
}

// TestRunTwoBadFiles covers keep()'s already-recorded-error branch by failing on
// two files: both are reported and the run fails.
func TestRunTwoBadFiles(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, "/no/such/a.png", "/no/such/b.png")
	if err == nil {
		t.Fatal("expected error for two missing files")
	}
	if !strings.Contains(errOut, "a.png") || !strings.Contains(errOut, "b.png") {
		t.Errorf("stderr = %q, want both files reported", errOut)
	}
}

// TestRunDefaultRadius covers processFile's auto-radius branch (no -r) together
// with calcButtonRadius.
func TestRunDefaultRadius(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	in := filepath.Join(dir, "in.png")
	out := filepath.Join(dir, "out.png")
	newPNG(t, in, 140, 140)

	if _, errOut, err := run(t, "-o", out, in); err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}
	if _, err := os.Stat(out); err != nil {
		t.Errorf("auto-radius output not produced: %v", err)
	}
}
