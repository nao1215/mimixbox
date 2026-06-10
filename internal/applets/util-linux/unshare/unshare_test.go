package unshare

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
	"golang.org/x/sys/unix"
)

func withStub(t *testing.T, err error) *int {
	t.Helper()
	flags := new(int)
	*flags = -1
	orig := unshareFn
	unshareFn = func(f int) error {
		*flags = f
		return err
	}
	t.Cleanup(func() { unshareFn = orig })
	return flags
}

func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	err := New().Run(context.Background(), io, args)
	return strings.TrimSpace(out.String()), err
}

func TestSingleNamespaceRunsCommand(t *testing.T) {
	flags := withStub(t, nil)
	out, err := run(t, "-u", "echo", "hello")
	if err != nil {
		t.Fatal(err)
	}
	if *flags != unix.CLONE_NEWUTS {
		t.Errorf("flags = %#x, want CLONE_NEWUTS", *flags)
	}
	if out != "hello" {
		t.Errorf("output = %q", out)
	}
}

func TestCombinedNamespaces(t *testing.T) {
	flags := withStub(t, nil)
	if _, err := run(t, "-m", "-i", "true"); err != nil {
		t.Fatal(err)
	}
	want := unix.CLONE_NEWNS | unix.CLONE_NEWIPC
	if *flags != want {
		t.Errorf("flags = %#x, want %#x", *flags, want)
	}
}

func TestRequiresNamespace(t *testing.T) {
	withStub(t, nil)
	if _, err := run(t, "echo", "x"); err == nil {
		t.Errorf("no namespace flag should fail")
	}
}

func TestUnshareFailure(t *testing.T) {
	withStub(t, errors.New("operation not permitted"))
	if _, err := run(t, "-m", "echo", "x"); err == nil {
		t.Errorf("an unshare failure should fail")
	}
}

func TestExitCodePropagates(t *testing.T) {
	withStub(t, nil)
	_, err := run(t, "-u", "sh", "-c", "exit 5")
	ee, ok := err.(*command.ExitError)
	if !ok || ee.Code != 5 {
		t.Errorf("err = %v, want exit 5", err)
	}
}
