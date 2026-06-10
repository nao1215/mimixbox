package blkdiscard

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

type discard struct {
	device         string
	offset, length uint64
}

func withStub(t *testing.T, size uint64, sizeErr, discardErr error) *discard {
	t.Helper()
	got := &discard{}
	os, od := sizeFn, discardFn
	sizeFn = func(string) (uint64, error) { return size, sizeErr }
	discardFn = func(device string, offset, length uint64) error {
		*got = discard{device, offset, length}
		return discardErr
	}
	t.Cleanup(func() { sizeFn, discardFn = os, od })
	return got
}

func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	err := New().Run(context.Background(), io, args)
	return strings.TrimSpace(out.String()), err
}

func TestExplicitRange(t *testing.T) {
	got := withStub(t, 0, nil, nil)
	if _, err := run(t, "-o", "4096", "-l", "8192", "/dev/sdb"); err != nil {
		t.Fatal(err)
	}
	if *got != (discard{"/dev/sdb", 4096, 8192}) {
		t.Errorf("discard = %+v", *got)
	}
}

func TestWholeDevice(t *testing.T) {
	got := withStub(t, 1000000, nil, nil)
	if _, err := run(t, "/dev/sdb"); err != nil {
		t.Fatal(err)
	}
	if *got != (discard{"/dev/sdb", 0, 1000000}) {
		t.Errorf("whole-device discard = %+v", *got)
	}
}

func TestOffsetToEnd(t *testing.T) {
	got := withStub(t, 1000000, nil, nil)
	if _, err := run(t, "-o", "200000", "/dev/sdb"); err != nil {
		t.Fatal(err)
	}
	if got.length != 800000 {
		t.Errorf("offset-to-end length = %d, want 800000", got.length)
	}
}

func TestVerbose(t *testing.T) {
	withStub(t, 0, nil, nil)
	out, err := run(t, "-v", "-o", "0", "-l", "4096", "/dev/sdb")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "discarded 4096 bytes from offset 0") {
		t.Errorf("verbose = %q", out)
	}
}

func TestErrors(t *testing.T) {
	withStub(t, 1000, nil, nil)
	if _, err := run(t); err == nil {
		t.Errorf("no device should fail")
	}
	if _, err := run(t, "-o", "5000", "/dev/sdb"); err == nil {
		t.Errorf("offset past end should fail")
	}
	withStub(t, 0, errors.New("permission denied"), nil)
	if _, err := run(t, "/dev/sdb"); err == nil {
		t.Errorf("a size failure should fail")
	}
	withStub(t, 0, nil, errors.New("permission denied"))
	if _, err := run(t, "-l", "4096", "/dev/sdb"); err == nil {
		t.Errorf("a discard failure should fail")
	}
}
