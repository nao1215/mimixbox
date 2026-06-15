package adjtimex

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, args []string) (string, string, error) {
	t.Helper()
	var out, errBuf bytes.Buffer
	io := command.IO{In: strings.NewReader(""), Out: &out, Err: &errBuf}
	err := New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestParseArgs(t *testing.T) {
	t.Parallel()
	r, err := ParseArgs("10000", "5", "-3", map[string]bool{"tick": true, "frequency": true, "offset": true})
	if err != nil {
		t.Fatalf("ParseArgs: %v", err)
	}
	if r.Tick != 10000 || r.Freq != 5 || r.Offset != -3 {
		t.Errorf("unexpected request: %+v", r)
	}
	if !r.Modifies() {
		t.Error("Modifies should be true")
	}
}

func TestParseArgsNoChanges(t *testing.T) {
	t.Parallel()
	r, err := ParseArgs("", "", "", map[string]bool{})
	if err != nil {
		t.Fatal(err)
	}
	if r.Modifies() {
		t.Error("Modifies should be false with no flags")
	}
}

func TestParseArgsBad(t *testing.T) {
	t.Parallel()
	if _, err := ParseArgs("x", "", "", map[string]bool{"tick": true}); err == nil {
		t.Error("expected error for bad tick")
	}
}

func TestRunQueryCapabilityError(t *testing.T) {
	t.Parallel()
	if _, _, err := run(t, nil); err == nil {
		t.Error("expected capability error for read-only query")
	}
}

func TestRunQueryInjected(t *testing.T) {
	orig := readStatusFn
	readStatusFn = func() (*status, error) {
		return &status{tick: 10000, freq: 1, offset: 2, statusFlags: 0x40}, nil
	}
	defer func() { readStatusFn = orig }()

	out, _, err := run(t, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "tick:      10000") || !strings.Contains(out, "status:    0x40") {
		t.Errorf("query output = %q", out)
	}
}

func TestRunSetCapabilityError(t *testing.T) {
	t.Parallel()
	if _, _, err := run(t, []string{"-t", "10000"}); err == nil {
		t.Error("expected capability error for set")
	}
}

func TestRunSetInjected(t *testing.T) {
	orig := applyFn
	var got *Request
	applyFn = func(r *Request) error { got = r; return nil }
	defer func() { applyFn = orig }()

	if _, _, err := run(t, []string{"-t", "9999", "-o", "1"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil || got.Tick != 9999 || got.Offset != 1 {
		t.Errorf("apply received unexpected request: %+v", got)
	}
}

func TestRunBadValue(t *testing.T) {
	t.Parallel()
	if _, _, err := run(t, []string{"-t", "notanumber"}); err == nil {
		t.Error("expected error for bad value")
	}
}
