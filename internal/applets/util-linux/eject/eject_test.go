package eject

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

type call struct {
	device    string
	closeTray bool
}

func withStub(t *testing.T, err error) *call {
	t.Helper()
	got := &call{device: "<unset>"}
	orig := ejectFn
	ejectFn = func(device string, closeTray bool) error {
		*got = call{device, closeTray}
		return err
	}
	t.Cleanup(func() { ejectFn = orig })
	return got
}

func run(t *testing.T, args ...string) error {
	t.Helper()
	io := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	return New().Run(context.Background(), io, args)
}

func TestDefaultDeviceEjects(t *testing.T) {
	got := withStub(t, nil)
	if err := run(t); err != nil {
		t.Fatal(err)
	}
	if got.device != defaultDevice || got.closeTray {
		t.Errorf("eject call = %+v, want open of %s", *got, defaultDevice)
	}
}

func TestNamedDeviceCloseTray(t *testing.T) {
	got := withStub(t, nil)
	if err := run(t, "-t", "/dev/sr0"); err != nil {
		t.Fatal(err)
	}
	if got.device != "/dev/sr0" || !got.closeTray {
		t.Errorf("eject -t call = %+v", *got)
	}
}

func TestFailure(t *testing.T) {
	withStub(t, errors.New("no medium found"))
	if err := run(t, "/dev/sr0"); err == nil {
		t.Errorf("an ioctl failure should fail")
	}
}
