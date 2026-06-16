package fakemovie_test

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/jokeutils/fakemovie"
	"github.com/nao1215/mimixbox/internal/command"
)

// newPNG creates a small solid-color PNG at path with the given dimensions.
func newPNG(t *testing.T, path string, w, h int) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{R: 10, G: 20, B: 30, A: 255})
		}
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create %s: %v", path, err)
	}
	defer func() { _ = f.Close() }()
	if err := png.Encode(f, img); err != nil {
		t.Fatalf("encode %s: %v", path, err)
	}
}

func run(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: &bytes.Buffer{}, Out: out, Err: errBuf}
	err := fakemovie.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestNewAndMetadata(t *testing.T) {
	t.Parallel()
	c := fakemovie.New()
	if c == nil {
		t.Fatal("New() returned nil")
	}
	if c.Name() != "fakemovie" {
		t.Errorf("Name() = %q, want %q", c.Name(), "fakemovie")
	}
	if c.Synopsis() != "Adds a video playback button to the image" {
		t.Errorf("Synopsis() = %q", c.Synopsis())
	}
}

func TestRunProducesImage(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	in := filepath.Join(dir, "in.png")
	out := filepath.Join(dir, "out.png")
	const w, h = 140, 100
	newPNG(t, in, w, h)

	_, errOut, err := run(t, "-o", out, in)
	if err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}

	f, err := os.Open(out)
	if err != nil {
		t.Fatalf("output not produced: %v", err)
	}
	defer func() { _ = f.Close() }()

	cfg, format, err := image.DecodeConfig(f)
	if err != nil {
		t.Fatalf("output is not a decodable image: %v", err)
	}
	if format != "png" {
		t.Errorf("format = %q, want png", format)
	}
	if cfg.Width != w || cfg.Height != h {
		t.Errorf("dimensions = %dx%d, want %dx%d", cfg.Width, cfg.Height, w, h)
	}
}

func TestRunDefaultOutputName(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	in := filepath.Join(dir, "sample.png")
	newPNG(t, in, 60, 60)

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(wd) })

	if _, errOut, err := run(t, in); err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}

	// decideOutputFileName: base name without ext + "_fake" + ext.
	if _, err := os.Stat(filepath.Join(dir, "sample_fake.png")); err != nil {
		t.Errorf("default output not produced: %v", err)
	}
}

func TestRunPhubButton(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	in := filepath.Join(dir, "in.png")
	out := filepath.Join(dir, "out_phub.png")
	newPNG(t, in, 80, 80)

	if _, errOut, err := run(t, "-p", "-o", out, in); err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}
	if _, err := os.Stat(out); err != nil {
		t.Errorf("phub output not produced: %v", err)
	}
}

func TestAddButtonPure(t *testing.T) {
	t.Parallel()
	const w, h = 120, 90
	src := image.NewRGBA(image.Rect(0, 0, w, h))

	for _, phub := range []bool{false, true} {
		got := fakemovie.AddButton(src, 20, phub)
		if got == nil {
			t.Fatalf("AddButton(phub=%v) returned nil", phub)
		}
		b := got.Bounds()
		if b.Dx() != w || b.Dy() != h {
			t.Errorf("AddButton(phub=%v) bounds = %dx%d, want %dx%d", phub, b.Dx(), b.Dy(), w, h)
		}
	}
}

func TestRunMissingOperand(t *testing.T) {
	t.Parallel()
	out, errOut, err := run(t)
	if err == nil {
		t.Fatal("expected error for missing operand")
	}
	if out != "" {
		t.Errorf("out = %q, want empty", out)
	}
	if !strings.Contains(errOut, "fakemovie: missing operand") {
		t.Errorf("stderr = %q, want missing operand message", errOut)
	}
}

func TestRunMissingFile(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, "/no/such/file.png")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	if !strings.Contains(errOut, "fakemovie: /no/such/file.png:") {
		t.Errorf("stderr = %q, want fakemovie error prefix", errOut)
	}
}

func TestRunInvalidExtension(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	in := filepath.Join(dir, "notimage.txt")
	if err := os.WriteFile(in, []byte("hello"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, errOut, err := run(t, in)
	if err == nil {
		t.Fatal("expected error for invalid extension")
	}
	if !strings.Contains(errOut, "fakemovie:") {
		t.Errorf("stderr = %q, want fakemovie error prefix", errOut)
	}
}

func TestRunHelp(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "--help")
	if err != nil {
		t.Fatalf("--help error = %v", err)
	}
	if !strings.Contains(out, "Usage: fakemovie") {
		t.Errorf("--help out = %q", out)
	}
	for _, want := range []string{"Examples:", "Exit status:"} {
		if !strings.Contains(out, want) {
			t.Errorf("--help output missing %q:\n%s", want, out)
		}
	}
}
