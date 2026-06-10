package fatattr

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func withStub(t *testing.T, cur uint32, getErr error) *uint32 {
	t.Helper()
	written := new(uint32)
	*written = 0xffff
	og, os := getAttrFn, setAttrFn
	getAttrFn = func(string) (uint32, error) { return cur, getErr }
	setAttrFn = func(_ string, attr uint32) error { *written = attr; return nil }
	t.Cleanup(func() { getAttrFn, setAttrFn = og, os })
	return written
}

func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	err := New().Run(context.Background(), io, args)
	return strings.TrimSpace(out.String()), err
}

func TestDisplay(t *testing.T) {
	withStub(t, 0x21, nil) // read-only + archive
	out, err := run(t, "file.txt")
	if err != nil {
		t.Fatal(err)
	}
	if out != "r----a file.txt" {
		t.Errorf("display = %q, want \"r----a file.txt\"", out)
	}
}

func TestAddAndRemove(t *testing.T) {
	// cur = archive (0x20); +r adds read-only, -a clears archive.
	w := withStub(t, 0x20, nil)
	if _, err := run(t, "+r", "-a", "file.txt"); err != nil {
		t.Fatal(err)
	}
	if *w != 0x01 {
		t.Errorf("set wrote %#x, want 0x01", *w)
	}
}

func TestLeadingRemove(t *testing.T) {
	// A leading "-h" must be treated as a remove operand, not a flag.
	w := withStub(t, 0x02, nil) // hidden set
	if _, err := run(t, "-h", "file.txt"); err != nil {
		t.Fatal(err)
	}
	if *w != 0x00 {
		t.Errorf("set wrote %#x, want 0x00", *w)
	}
}

func TestDecode(t *testing.T) {
	t.Parallel()
	if got := decode(0); got != "------" {
		t.Errorf("decode(0) = %q", got)
	}
	if got := decode(0x02); got != "-h----" {
		t.Errorf("decode(hidden) = %q", got)
	}
	if got := decode(0x21); got != "r----a" {
		t.Errorf("decode(ro|archive) = %q", got)
	}
}

func TestParseAttrs(t *testing.T) {
	t.Parallel()
	bits, err := parseAttrs("rh")
	if err != nil || bits != 0x03 {
		t.Errorf("parseAttrs(rh) = %#x, %v", bits, err)
	}
	if _, err := parseAttrs("Z"); err == nil {
		t.Errorf("unknown attribute should error")
	}
}

func TestErrors(t *testing.T) {
	withStub(t, 0, nil)
	if _, err := run(t); err == nil {
		t.Errorf("no file should fail")
	}
	if _, err := run(t, "+Z", "file"); err == nil {
		t.Errorf("bad attribute should fail")
	}
	withStub(t, 0, errors.New("inappropriate ioctl"))
	if _, err := run(t, "file"); err == nil {
		t.Errorf("a get failure should fail")
	}
}
