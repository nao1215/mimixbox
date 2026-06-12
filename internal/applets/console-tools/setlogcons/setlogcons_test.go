package setlogcons

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
	*got = -99
	orig := setFn
	setFn = func(n int) error { *got = n; return err }
	t.Cleanup(func() { setFn = orig })
	return got
}

func run(t *testing.T, args ...string) error {
	t.Helper()
	io := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	return New().Run(context.Background(), io, args)
}

func TestSetsLogConsole(t *testing.T) {
	got := stub(t, nil)
	if err := run(t, "12"); err != nil {
		t.Fatal(err)
	}
	if *got != 12 {
		t.Errorf("set to %d, want 12", *got)
	}
}

func TestNoArgMeansCurrent(t *testing.T) {
	got := stub(t, nil)
	if err := run(t); err != nil {
		t.Fatal(err)
	}
	if *got != 0 {
		t.Errorf("no arg should be 0 (current), got %d", *got)
	}
}

func TestErrors(t *testing.T) {
	stub(t, nil)
	if err := run(t, "foo"); err == nil {
		t.Errorf("a non-numeric N should fail")
	}
	if err := run(t, "-1"); err == nil {
		t.Errorf("a negative N should fail")
	}
	stub(t, errors.New("permission denied"))
	if err := run(t, "1"); err == nil {
		t.Errorf("an ioctl failure should fail")
	}
}
