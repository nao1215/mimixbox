package dumpkmap

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/console-tools/internal/kbd"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, args []string) (*bytes.Buffer, string, error) {
	t.Helper()
	var out, errBuf bytes.Buffer
	io := command.IO{In: strings.NewReader(""), Out: &out, Err: &errBuf}
	err := New().Run(context.Background(), io, args)
	return &out, errBuf.String(), err
}

func TestRunCapabilityError(t *testing.T) {
	t.Parallel()
	if _, _, err := run(t, nil); err == nil {
		t.Fatal("expected a capability error without a console")
	}
}

func TestRunInjectedSuccess(t *testing.T) {
	orig := readKeymapFn
	km := &kbd.Keymap{}
	km.Present[0] = true
	readKeymapFn = func() (*kbd.Keymap, error) { return km, nil }
	defer func() { readKeymapFn = orig }()

	out, _, err := run(t, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// The output must decode back to a valid keymap.
	got, err := kbd.DecodeKeymap(bytes.NewReader(out.Bytes()))
	if err != nil {
		t.Fatalf("output is not a valid keymap: %v", err)
	}
	if len(got.PresentTables()) != 1 {
		t.Errorf("present tables = %v", got.PresentTables())
	}
}

func TestRunUnexpectedArg(t *testing.T) {
	t.Parallel()
	if _, _, err := run(t, []string{"x"}); err == nil {
		t.Error("expected error for unexpected argument")
	}
}

func TestHelp(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, []string{"--help"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "Usage: dumpkmap") {
		t.Errorf("help missing usage:\n%s", out.String())
	}
}
