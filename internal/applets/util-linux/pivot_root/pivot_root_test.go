package pivotroot

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

type call struct{ newRoot, putOld string }

func withStub(t *testing.T, err error) *call {
	t.Helper()
	got := &call{}
	orig := pivotFn
	pivotFn = func(newRoot, putOld string) error {
		*got = call{newRoot, putOld}
		return err
	}
	t.Cleanup(func() { pivotFn = orig })
	return got
}

func run(t *testing.T, args ...string) error {
	t.Helper()
	io := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	return New().Run(context.Background(), io, args)
}

func TestPivots(t *testing.T) {
	got := withStub(t, nil)
	if err := run(t, "/newroot", "/newroot/old"); err != nil {
		t.Fatal(err)
	}
	if *got != (call{"/newroot", "/newroot/old"}) {
		t.Errorf("pivotFn = %+v", *got)
	}
}

func TestWrongArgCount(t *testing.T) {
	withStub(t, nil)
	if err := run(t, "/newroot"); err == nil {
		t.Errorf("one argument should fail")
	}
	if err := run(t, "/a", "/b", "/c"); err == nil {
		t.Errorf("three arguments should fail")
	}
	if err := run(t); err == nil {
		t.Errorf("no arguments should fail")
	}
}

func TestSyscallFailure(t *testing.T) {
	withStub(t, errors.New("operation not permitted"))
	if err := run(t, "/a", "/b"); err == nil {
		t.Errorf("a syscall failure should fail")
	}
}
