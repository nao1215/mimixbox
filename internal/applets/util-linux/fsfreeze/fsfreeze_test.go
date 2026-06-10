package fsfreeze

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

type call struct {
	path   string
	freeze bool
}

func withStub(t *testing.T, err error) *call {
	t.Helper()
	got := &call{}
	got.path = "<unset>"
	orig := freezeFn
	freezeFn = func(path string, freeze bool) error {
		*got = call{path, freeze}
		return err
	}
	t.Cleanup(func() { freezeFn = orig })
	return got
}

func run(t *testing.T, args ...string) error {
	t.Helper()
	io := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	return New().Run(context.Background(), io, args)
}

func TestFreeze(t *testing.T) {
	got := withStub(t, nil)
	if err := run(t, "-f", "/mnt/data"); err != nil {
		t.Fatal(err)
	}
	if got.path != "/mnt/data" || !got.freeze {
		t.Errorf("freeze call = %+v", *got)
	}
}

func TestUnfreeze(t *testing.T) {
	got := withStub(t, nil)
	if err := run(t, "-u", "/mnt/data"); err != nil {
		t.Fatal(err)
	}
	if got.path != "/mnt/data" || got.freeze {
		t.Errorf("unfreeze call = %+v", *got)
	}
}

func TestRequiresExactlyOneMode(t *testing.T) {
	withStub(t, nil)
	if err := run(t, "/mnt"); err == nil {
		t.Errorf("no -f/-u should fail")
	}
	if err := run(t, "-f", "-u", "/mnt"); err == nil {
		t.Errorf("both -f and -u should fail")
	}
}

func TestRequiresMountpoint(t *testing.T) {
	withStub(t, nil)
	if err := run(t, "-f"); err == nil {
		t.Errorf("missing mount point should fail")
	}
}

func TestIoctlFailure(t *testing.T) {
	withStub(t, errors.New("operation not permitted"))
	if err := run(t, "-f", "/mnt"); err == nil {
		t.Errorf("an ioctl failure should fail")
	}
}
