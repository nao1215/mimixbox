package chattr

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func withStub(t *testing.T, cur int, getErr error) *int {
	t.Helper()
	written := new(int)
	*written = -1
	og, os := getFlags, setFlags
	getFlags = func(string) (int, error) { return cur, getErr }
	setFlags = func(_ string, flags int) error { *written = flags; return nil }
	t.Cleanup(func() { getFlags, setFlags = og, os })
	return written
}

func run(t *testing.T, args ...string) error {
	t.Helper()
	io := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	return New().Run(context.Background(), io, args)
}

func TestAddRemoveSet(t *testing.T) {
	w := withStub(t, 0, nil)
	if err := run(t, "+i", "file"); err != nil {
		t.Fatal(err)
	}
	if *w != 0x10 {
		t.Errorf("+i wrote %#x, want 0x10", *w)
	}

	w = withStub(t, 0x20, nil) // append-only set
	if err := run(t, "-a", "file"); err != nil {
		t.Fatal(err)
	}
	if *w != 0 {
		t.Errorf("-a wrote %#x, want 0", *w)
	}

	w = withStub(t, 0xff, nil)
	if err := run(t, "=e", "file"); err != nil {
		t.Fatal(err)
	}
	if *w != 0x80000 {
		t.Errorf("=e wrote %#x, want 0x80000", *w)
	}
}

func TestApply(t *testing.T) {
	t.Parallel()
	if got := apply('+', 0, 0x10); got != 0x10 {
		t.Errorf("apply(+) = %#x", got)
	}
	if got := apply('-', 0x30, 0x20); got != 0x10 {
		t.Errorf("apply(-) = %#x", got)
	}
	if got := apply('=', 0xffff, 0x80000); got != 0x80000 {
		t.Errorf("apply(=) = %#x", got)
	}
}

func TestParseAttrs(t *testing.T) {
	t.Parallel()
	bits, err := parseAttrs("ia")
	if err != nil || bits != 0x10|0x20 {
		t.Errorf("parseAttrs(ia) = %#x, %v", bits, err)
	}
	if _, err := parseAttrs("Z"); err == nil {
		t.Errorf("unknown attribute should error")
	}
}

func TestErrors(t *testing.T) {
	withStub(t, 0, nil)
	if err := run(t, "xi", "file"); err == nil {
		t.Errorf("bad mode should fail")
	}
	if err := run(t, "+Z", "file"); err == nil {
		t.Errorf("bad attribute should fail")
	}
	if err := run(t, "+i"); err == nil {
		t.Errorf("missing file should fail")
	}
	withStub(t, 0, errors.New("no such file"))
	if err := run(t, "+i", "file"); err == nil {
		t.Errorf("a get failure should fail")
	}
}
