package setkeycodes

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

type pair struct{ scancode, keycode int }

func stub(t *testing.T, err error) *[]pair {
	t.Helper()
	var calls []pair
	orig := setKeycodeFn
	setKeycodeFn = func(s, k int) error { calls = append(calls, pair{s, k}); return err }
	t.Cleanup(func() { setKeycodeFn = orig })
	return &calls
}

func run(t *testing.T, args ...string) error {
	t.Helper()
	io := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	return New().Run(context.Background(), io, args)
}

func TestSetsPairs(t *testing.T) {
	calls := stub(t, nil)
	if err := run(t, "e060", "122", "e061", "123"); err != nil {
		t.Fatal(err)
	}
	if len(*calls) != 2 || (*calls)[0] != (pair{0xe060, 122}) || (*calls)[1] != (pair{0xe061, 123}) {
		t.Errorf("calls = %v", *calls)
	}
}

func TestErrors(t *testing.T) {
	stub(t, nil)
	if err := run(t); err == nil {
		t.Errorf("no args should fail")
	}
	if err := run(t, "e060"); err == nil {
		t.Errorf("an odd number of args should fail")
	}
	if err := run(t, "zz", "1"); err == nil {
		t.Errorf("a bad scancode should fail")
	}
	if err := run(t, "e060", "notanumber"); err == nil {
		t.Errorf("a bad keycode should fail")
	}
	stub(t, errors.New("permission denied"))
	if err := run(t, "e060", "122"); err == nil {
		t.Errorf("an ioctl failure should fail")
	}
}
