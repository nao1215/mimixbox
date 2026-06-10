package lsattr

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func withFlags(t *testing.T, byPath map[string]int, fail bool) {
	t.Helper()
	orig := getFlags
	getFlags = func(path string) (int, error) {
		if fail {
			return 0, errors.New("operation not supported")
		}
		return byPath[path], nil
	}
	t.Cleanup(func() { getFlags = orig })
}

func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	err := New().Run(context.Background(), io, args)
	return strings.TrimSpace(out.String()), err
}

func TestDecode(t *testing.T) {
	t.Parallel()
	if got := decode(0); got != "----------------" {
		t.Errorf("decode(0) = %q", got)
	}
	if got := decode(0x10); got != "----i-----------" { // immutable, position 5
		t.Errorf("decode(immutable) = %q", got)
	}
	if got := decode(0x80000); got != "--------------e-" { // extents, position 15
		t.Errorf("decode(extents) = %q", got)
	}
	// immutable + append (0x10 | 0x20).
	if got := decode(0x30); got != "----ia----------" {
		t.Errorf("decode(immutable|append) = %q", got)
	}
}

func TestRun(t *testing.T) {
	withFlags(t, map[string]int{"file.txt": 0x10}, false)
	out, err := run(t, "file.txt")
	if err != nil {
		t.Fatal(err)
	}
	if out != "----i----------- file.txt" {
		t.Errorf("lsattr = %q", out)
	}
}

func TestReadFailure(t *testing.T) {
	withFlags(t, nil, true)
	if _, err := run(t, "file.txt"); err == nil {
		t.Errorf("a read failure should fail")
	}
}
