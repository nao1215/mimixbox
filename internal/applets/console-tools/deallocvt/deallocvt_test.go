package deallocvt

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
	orig := deallocFn
	deallocFn = func(n int) error { *got = n; return err }
	t.Cleanup(func() { deallocFn = orig })
	return got
}

func run(t *testing.T, args ...string) error {
	t.Helper()
	io := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	return New().Run(context.Background(), io, args)
}

func TestDeallocatesN(t *testing.T) {
	got := stub(t, nil)
	if err := run(t, "3"); err != nil {
		t.Fatal(err)
	}
	if *got != 3 {
		t.Errorf("deallocated %d, want 3", *got)
	}
}

func TestNoArgMeansAll(t *testing.T) {
	got := stub(t, nil)
	if err := run(t); err != nil {
		t.Fatal(err)
	}
	if *got != 0 {
		t.Errorf("no arg should deallocate all (0), got %d", *got)
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
	stub(t, errors.New("device busy"))
	if err := run(t, "2"); err == nil {
		t.Errorf("an ioctl failure should fail")
	}
}
