package setconsole

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func stub(t *testing.T, err error) *string {
	t.Helper()
	got := new(string)
	*got = "<unset>"
	orig := redirectFn
	redirectFn = func(device string) error { *got = device; return err }
	t.Cleanup(func() { redirectFn = orig })
	return got
}

func run(t *testing.T, args ...string) error {
	t.Helper()
	io := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	return New().Run(context.Background(), io, args)
}

func TestRedirectsToDevice(t *testing.T) {
	got := stub(t, nil)
	if err := run(t, "/dev/ttyS0"); err != nil {
		t.Fatal(err)
	}
	if *got != "/dev/ttyS0" {
		t.Errorf("redirected to %q, want /dev/ttyS0", *got)
	}
}

func TestDefaultAndReset(t *testing.T) {
	got := stub(t, nil)
	if err := run(t); err != nil {
		t.Fatal(err)
	}
	if *got != defaultDevice {
		t.Errorf("default = %q, want %s", *got, defaultDevice)
	}
	if err := run(t, "-r"); err != nil {
		t.Fatal(err)
	}
	if *got != defaultDevice {
		t.Errorf("-r = %q, want %s", *got, defaultDevice)
	}
}

func TestFailure(t *testing.T) {
	stub(t, errors.New("permission denied"))
	if err := run(t, "/dev/ttyS0"); err == nil {
		t.Errorf("an ioctl failure should fail")
	}
}
