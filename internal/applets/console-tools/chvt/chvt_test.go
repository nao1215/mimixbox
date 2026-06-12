package chvt

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func stub(t *testing.T, err error) *int {
	t.Helper()
	got := new(int)
	*got = -1
	orig := switchFn
	switchFn = func(n int) error { *got = n; return err }
	t.Cleanup(func() { switchFn = orig })
	return got
}

func run(t *testing.T, args ...string) error {
	t.Helper()
	io := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	return New().Run(context.Background(), io, args)
}

func TestSwitchesToVT(t *testing.T) {
	got := stub(t, nil)
	if err := run(t, "2"); err != nil {
		t.Fatal(err)
	}
	if *got != 2 {
		t.Errorf("switched to %d, want 2", *got)
	}
}

func TestErrors(t *testing.T) {
	stub(t, nil)
	if err := run(t); err == nil {
		t.Errorf("missing N should fail")
	}
	if err := run(t, "foo"); err == nil {
		t.Errorf("a non-numeric N should fail")
	}
	if err := run(t, "0"); err == nil {
		t.Errorf("a zero N should fail")
	}
}

func TestSwitchFailure(t *testing.T) {
	stub(t, errors.New("operation not permitted"))
	if err := run(t, "3"); err == nil {
		t.Errorf("a switch failure should fail")
	}
}
