package nice

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func TestNicePrintsCurrent(t *testing.T) {
	origGet := getpriority
	getpriority = func() (int, error) { return 7, nil }
	defer func() { getpriority = origGet }()

	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, nil); err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(out.String()) != "7" {
		t.Errorf("nice = %q, want 7", out.String())
	}
}

func TestNiceAdjustsAndRuns(t *testing.T) {
	origGet, origSet := getpriority, setpriority
	getpriority = func() (int, error) { return 5, nil }
	var setTo int
	setpriority = func(n int) error { setTo = n; return nil }
	defer func() { getpriority, setpriority = origGet, origSet }()

	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	// "true" exits 0 on every supported platform.
	if err := New().Run(context.Background(), io, []string{"-n", "3", "true"}); err != nil {
		t.Fatalf("nice run error = %v", err)
	}
	if setTo != 8 {
		t.Errorf("setpriority got %d, want 8 (5+3)", setTo)
	}
}

func TestNiceCommandNotFound(t *testing.T) {
	origGet, origSet := getpriority, setpriority
	getpriority = func() (int, error) { return 0, nil }
	setpriority = func(int) error { return nil }
	defer func() { getpriority, setpriority = origGet, origSet }()

	io := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	err := New().Run(context.Background(), io, []string{"no-such-command-xyz"})
	var ee *command.ExitError
	if err == nil {
		t.Fatal("expected an error for a missing command")
	}
	if e, ok := err.(*command.ExitError); ok {
		ee = e
	}
	if ee == nil || ee.Code != 127 {
		t.Errorf("err = %v, want exit 127", err)
	}
}
