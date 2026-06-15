package setfont

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/console-tools/internal/kbd"
	"github.com/nao1215/mimixbox/internal/command"
)

func psf1(charSize int) []byte {
	b := []byte{0x36, 0x04, 0x00, byte(charSize)}
	return append(b, make([]byte, 256*charSize)...)
}

func writeFont(t *testing.T, data []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "font.psf")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func run(t *testing.T, args []string) (string, error) {
	t.Helper()
	var out, errBuf bytes.Buffer
	io := command.IO{In: strings.NewReader(""), Out: &out, Err: &errBuf}
	err := New().Run(context.Background(), io, args)
	return out.String(), err
}

func TestRunNoArg(t *testing.T) {
	t.Parallel()
	if _, err := run(t, nil); err == nil {
		t.Error("expected error when no font file given")
	}
}

func TestRunMissingFile(t *testing.T) {
	t.Parallel()
	if _, err := run(t, []string{"/no/such/font.psf"}); err == nil {
		t.Error("expected error for missing file")
	}
}

func TestRunInvalidFont(t *testing.T) {
	t.Parallel()
	path := writeFont(t, []byte("not-a-font"))
	if _, err := run(t, []string{path}); err == nil {
		t.Error("expected error for invalid font")
	}
}

func TestRunValidButCapabilityError(t *testing.T) {
	t.Parallel()
	path := writeFont(t, psf1(16))
	if _, err := run(t, []string{path}); err == nil {
		t.Fatal("expected capability error from apply step")
	}
}

func TestRunVerboseSuccess(t *testing.T) {
	orig := applyFontFn
	var applied *kbd.Font
	applyFontFn = func(f *kbd.Font) error { applied = f; return nil }
	defer func() { applyFontFn = orig }()

	path := writeFont(t, psf1(16))
	out, err := run(t, []string{"-v", path})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "PSF1 font") {
		t.Errorf("verbose output missing description: %q", out)
	}
	if applied == nil || applied.CharSize != 16 {
		t.Errorf("apply received unexpected font: %+v", applied)
	}
}

func TestRunExtraArg(t *testing.T) {
	t.Parallel()
	path := writeFont(t, psf1(16))
	if _, err := run(t, []string{path, "extra"}); err == nil {
		t.Error("expected error for extra argument")
	}
}
