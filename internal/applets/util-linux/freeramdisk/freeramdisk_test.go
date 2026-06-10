package freeramdisk

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func withStub(t *testing.T, err error) *string {
	t.Helper()
	got := new(string)
	*got = "<unset>"
	orig := flushFn
	flushFn = func(device string) error {
		*got = device
		return err
	}
	t.Cleanup(func() { flushFn = orig })
	return got
}

func run(t *testing.T, args ...string) error {
	t.Helper()
	io := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	return New().Run(context.Background(), io, args)
}

func TestFreesDevice(t *testing.T) {
	got := withStub(t, nil)
	if err := run(t, "/dev/ram0"); err != nil {
		t.Fatal(err)
	}
	if *got != "/dev/ram0" {
		t.Errorf("flushFn called with %q", *got)
	}
}

func TestNoDevice(t *testing.T) {
	withStub(t, nil)
	if err := run(t); err == nil {
		t.Errorf("missing device should fail")
	}
}

func TestIoctlFailure(t *testing.T) {
	withStub(t, errors.New("operation not permitted"))
	if err := run(t, "/dev/ram0"); err == nil {
		t.Errorf("an ioctl failure should fail")
	}
}
