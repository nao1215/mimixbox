package loadkmap

import (
	"bytes"
	"context"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/console-tools/internal/kbd"
	"github.com/nao1215/mimixbox/internal/command"
)

func validKeymap(t *testing.T) []byte {
	t.Helper()
	km := &kbd.Keymap{}
	km.Present[0] = true
	var b bytes.Buffer
	if err := kbd.EncodeKeymap(&b, km); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

func run(t *testing.T, in []byte, args []string) (string, error) {
	t.Helper()
	var out, errBuf bytes.Buffer
	io := command.IO{In: bytes.NewReader(in), Out: &out, Err: &errBuf}
	err := New().Run(context.Background(), io, args)
	return errBuf.String(), err
}

// Without privilege, a valid keymap still parses but the apply step fails.
func TestRunValidButCapabilityError(t *testing.T) {
	t.Parallel()
	if _, err := run(t, validKeymap(t), nil); err == nil {
		t.Fatal("expected capability error from apply step")
	}
}

func TestRunInvalidKeymap(t *testing.T) {
	t.Parallel()
	if _, err := run(t, []byte("garbage-garbage-garbage"), nil); err == nil {
		t.Fatal("expected error for invalid keymap")
	}
}

func TestRunEmptyKeymap(t *testing.T) {
	t.Parallel()
	km := &kbd.Keymap{} // no present tables
	var b bytes.Buffer
	_ = kbd.EncodeKeymap(&b, km)
	if _, err := run(t, b.Bytes(), nil); err == nil {
		t.Fatal("expected error for empty keymap")
	}
}

func TestRunInjectedSuccess(t *testing.T) {
	orig := applyKeymapFn
	var applied *kbd.Keymap
	applyKeymapFn = func(km *kbd.Keymap) error { applied = km; return nil }
	defer func() { applyKeymapFn = orig }()

	if _, err := run(t, validKeymap(t), nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if applied == nil || len(applied.PresentTables()) != 1 {
		t.Errorf("apply received unexpected keymap: %+v", applied)
	}
}

func TestRunUnexpectedArg(t *testing.T) {
	t.Parallel()
	if _, err := run(t, validKeymap(t), []string{"file"}); err == nil {
		t.Error("expected error for unexpected argument")
	}
}
