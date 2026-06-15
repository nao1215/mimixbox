package loadfont

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/console-tools/internal/kbd"
	"github.com/nao1215/mimixbox/internal/command"
)

// psf1 builds a minimal valid PSF1 font (mode 0, 256 glyphs of charSize bytes).
func psf1(charSize int) []byte {
	b := []byte{0x36, 0x04, 0x00, byte(charSize)}
	return append(b, make([]byte, 256*charSize)...)
}

func run(t *testing.T, in []byte, args []string) (string, error) {
	t.Helper()
	var out, errBuf bytes.Buffer
	io := command.IO{In: bytes.NewReader(in), Out: &out, Err: &errBuf}
	err := New().Run(context.Background(), io, args)
	return errBuf.String(), err
}

func TestRunValidButCapabilityError(t *testing.T) {
	t.Parallel()
	if _, err := run(t, psf1(16), nil); err == nil {
		t.Fatal("expected capability error from apply step")
	}
}

func TestRunInvalidFont(t *testing.T) {
	t.Parallel()
	if _, err := run(t, []byte("not-a-font"), nil); err == nil {
		t.Fatal("expected error for invalid font")
	}
}

func TestRunVerbose(t *testing.T) {
	orig := applyFontFn
	applyFontFn = func(_ *kbd.Font) error { return nil }
	defer func() { applyFontFn = orig }()

	errOut, err := run(t, psf1(16), []string{"-v"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(errOut, "PSF1 font") {
		t.Errorf("verbose output missing font description: %q", errOut)
	}
}

func TestRunInjectedSuccess(t *testing.T) {
	orig := applyFontFn
	var applied *kbd.Font
	applyFontFn = func(f *kbd.Font) error { applied = f; return nil }
	defer func() { applyFontFn = orig }()

	if _, err := run(t, psf1(8), nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if applied == nil || applied.CharSize != 8 {
		t.Errorf("apply received unexpected font: %+v", applied)
	}
}

func TestRunUnexpectedArg(t *testing.T) {
	t.Parallel()
	if _, err := run(t, psf1(16), []string{"font.psf"}); err == nil {
		t.Error("expected error for unexpected argument")
	}
}
