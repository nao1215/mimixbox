package fbset

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func withStub(t *testing.T, info varScreenInfo, err error) *string {
	t.Helper()
	asked := new(string)
	*asked = "<unset>"
	orig := readVarFn
	readVarFn = func(device string) (varScreenInfo, error) {
		*asked = device
		return info, err
	}
	t.Cleanup(func() { readVarFn = orig })
	return asked
}

func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	err := New().Run(context.Background(), io, args)
	return out.String(), err
}

func TestShowsMode(t *testing.T) {
	asked := withStub(t, varScreenInfo{xres: 1920, yres: 1080, bpp: 32}, nil)
	out, err := run(t)
	if err != nil {
		t.Fatal(err)
	}
	if *asked != "/dev/fb0" {
		t.Errorf("default device = %q", *asked)
	}
	for _, want := range []string{`mode "1920x1080"`, "geometry 1920 1080 1920 1080 32", "endmode"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
}

func TestCustomDevice(t *testing.T) {
	asked := withStub(t, varScreenInfo{xres: 800, yres: 600, bpp: 16}, nil)
	if _, err := run(t, "-fb", "/dev/fb1"); err != nil {
		t.Fatal(err)
	}
	if *asked != "/dev/fb1" {
		t.Errorf("device = %q, want /dev/fb1", *asked)
	}
}

func TestReadFailure(t *testing.T) {
	withStub(t, varScreenInfo{}, errors.New("no such device"))
	if _, err := run(t); err == nil {
		t.Errorf("a read failure should fail")
	}
}
