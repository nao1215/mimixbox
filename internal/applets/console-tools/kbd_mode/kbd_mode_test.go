package kbdmode

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func stub(t *testing.T, getMode int, getErr error) *int {
	t.Helper()
	setTo := new(int)
	*setTo = -1
	og, os := getModeFn, setModeFn
	getModeFn = func() (int, error) { return getMode, getErr }
	setModeFn = func(mode int) error { *setTo = mode; return nil }
	t.Cleanup(func() { getModeFn, setModeFn = og, os })
	return setTo
}

func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	err := New().Run(context.Background(), io, args)
	return out.String(), err
}

func TestReportsMode(t *testing.T) {
	stub(t, kUnicode, nil)
	out, err := run(t)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Unicode") {
		t.Errorf("report = %q", out)
	}
}

func TestSetsMode(t *testing.T) {
	for _, c := range []struct {
		flag string
		want int
	}{{"-a", kXlate}, {"-u", kUnicode}, {"-k", kMediumRaw}, {"-s", kRaw}} {
		setTo := stub(t, kXlate, nil)
		if _, err := run(t, c.flag); err != nil {
			t.Fatalf("%s: %v", c.flag, err)
		}
		if *setTo != c.want {
			t.Errorf("%s set mode %d, want %d", c.flag, *setTo, c.want)
		}
	}
}

func TestErrors(t *testing.T) {
	stub(t, kXlate, nil)
	if _, err := run(t, "-a", "-u"); err == nil {
		t.Errorf("conflicting options should fail")
	}
	stub(t, 0, errors.New("inappropriate ioctl"))
	if _, err := run(t); err == nil {
		t.Errorf("a read failure should fail")
	}
}
