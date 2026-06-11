package fdformat

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func withStub(t *testing.T, tracks int, err error) *string {
	t.Helper()
	got := new(string)
	*got = "<unset>"
	orig := formatFn
	formatFn = func(device string) (int, error) {
		*got = device
		return tracks, err
	}
	t.Cleanup(func() { formatFn = orig })
	return got
}

func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	err := New().Run(context.Background(), io, args)
	return out.String(), err
}

func TestFormatsDevice(t *testing.T) {
	got := withStub(t, 160, nil)
	out, err := run(t, "/dev/fd0")
	if err != nil {
		t.Fatal(err)
	}
	if *got != "/dev/fd0" {
		t.Errorf("formatFn called with %q", *got)
	}
	if !strings.Contains(out, "Formatting /dev/fd0") || !strings.Contains(out, "160 tracks") {
		t.Errorf("output = %q", out)
	}
}

func TestNoDevice(t *testing.T) {
	withStub(t, 0, nil)
	if _, err := run(t); err == nil {
		t.Errorf("missing device should fail")
	}
}

func TestFailure(t *testing.T) {
	withStub(t, 0, errors.New("no such device"))
	if _, err := run(t, "/dev/fd0"); err == nil {
		t.Errorf("a format failure should fail")
	}
}
