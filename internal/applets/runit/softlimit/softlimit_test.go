package softlimit

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
	"golang.org/x/sys/unix"
)

func stub(t *testing.T, failResource int) map[int]uint64 {
	t.Helper()
	set := map[int]uint64{}
	orig := setRlimitFn
	setRlimitFn = func(resource int, value uint64) error {
		if resource == failResource {
			return errors.New("cannot set limit")
		}
		set[resource] = value
		return nil
	}
	t.Cleanup(func() { setRlimitFn = orig })
	return set
}

func run(t *testing.T, args ...string) error {
	t.Helper()
	io := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	return New().Run(context.Background(), io, args)
}

func TestSetsLimits(t *testing.T) {
	set := stub(t, -1)
	if err := run(t, "-m", "100000000", "-o", "64", "true"); err != nil {
		t.Fatal(err)
	}
	if set[unix.RLIMIT_AS] != 100000000 {
		t.Errorf("RLIMIT_AS = %d, want 100000000", set[unix.RLIMIT_AS])
	}
	if set[unix.RLIMIT_NOFILE] != 64 {
		t.Errorf("RLIMIT_NOFILE = %d, want 64", set[unix.RLIMIT_NOFILE])
	}
	// Unset flags must not set a limit.
	if _, ok := set[unix.RLIMIT_CPU]; ok {
		t.Errorf("RLIMIT_CPU should not be set")
	}
}

func TestExitCodePropagates(t *testing.T) {
	stub(t, -1)
	err := run(t, "-o", "10", "sh", "-c", "exit 5")
	ee, ok := err.(*command.ExitError)
	if !ok || ee.Code != 5 {
		t.Errorf("err = %v, want exit 5", err)
	}
}

func TestNoProgram(t *testing.T) {
	stub(t, -1)
	if err := run(t, "-o", "10"); err == nil {
		t.Errorf("a missing program should fail")
	}
}

func TestSetLimitFailure(t *testing.T) {
	stub(t, unix.RLIMIT_NOFILE)
	if err := run(t, "-o", "10", "true"); err == nil {
		t.Errorf("a failed setrlimit should fail")
	}
}
