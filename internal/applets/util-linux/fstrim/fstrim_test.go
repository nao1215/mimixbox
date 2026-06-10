package fstrim

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func withStub(t *testing.T, trimmed uint64, err error) *string {
	t.Helper()
	asked := new(string)
	*asked = "<unset>"
	orig := trimFn
	trimFn = func(path string) (uint64, error) {
		*asked = path
		return trimmed, err
	}
	t.Cleanup(func() { trimFn = orig })
	return asked
}

func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	err := New().Run(context.Background(), io, args)
	return strings.TrimSpace(out.String()), err
}

func TestVerboseReportsBytes(t *testing.T) {
	asked := withStub(t, 1024000, nil)
	out, err := run(t, "-v", "/home")
	if err != nil {
		t.Fatal(err)
	}
	if *asked != "/home" {
		t.Errorf("trimFn called with %q", *asked)
	}
	if out != "/home: 1024000 bytes were trimmed" {
		t.Errorf("verbose output = %q", out)
	}
}

func TestQuietByDefault(t *testing.T) {
	withStub(t, 1024000, nil)
	out, err := run(t, "/")
	if err != nil {
		t.Fatal(err)
	}
	if out != "" {
		t.Errorf("non-verbose should print nothing, got %q", out)
	}
}

func TestNoMountpoint(t *testing.T) {
	withStub(t, 0, nil)
	if _, err := run(t); err == nil {
		t.Errorf("a missing mount point should fail")
	}
}

func TestIoctlFailure(t *testing.T) {
	withStub(t, 0, errors.New("operation not permitted"))
	if _, err := run(t, "-v", "/"); err == nil {
		t.Errorf("an ioctl failure should fail")
	}
}
