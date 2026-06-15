package conspy

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, args []string) (string, error) {
	t.Helper()
	var out, errBuf bytes.Buffer
	io := command.IO{In: strings.NewReader(""), Out: &out, Err: &errBuf}
	err := New().Run(context.Background(), io, args)
	return errBuf.String(), err
}

func TestParseOptions(t *testing.T) {
	t.Parallel()
	o, err := ParseOptions("5", true, true, false)
	if err != nil {
		t.Fatalf("ParseOptions: %v", err)
	}
	if o.VT != 5 || !o.ReadOnly || !o.NoColors || o.Quiet {
		t.Errorf("unexpected options: %+v", o)
	}
}

func TestParseOptionsDefaultVT(t *testing.T) {
	t.Parallel()
	o, err := ParseOptions("", false, false, false)
	if err != nil {
		t.Fatal(err)
	}
	if o.VT != 0 {
		t.Errorf("VT = %d, want 0", o.VT)
	}
}

func TestParseOptionsBadVT(t *testing.T) {
	t.Parallel()
	for _, bad := range []string{"0", "64", "x", "-1"} {
		if _, err := ParseOptions(bad, false, false, false); err == nil {
			t.Errorf("ParseOptions(%q) expected error", bad)
		}
	}
}

func TestRunCapabilityError(t *testing.T) {
	t.Parallel()
	if _, err := run(t, []string{"2"}); err == nil {
		t.Error("expected capability error without a console")
	}
}

func TestRunBadVT(t *testing.T) {
	t.Parallel()
	if _, err := run(t, []string{"99"}); err == nil {
		t.Error("expected error for bad VT number")
	}
}

func TestRunInjectedSuccess(t *testing.T) {
	orig := spyFn
	var got *Options
	spyFn = func(o *Options) error { got = o; return nil }
	defer func() { spyFn = orig }()

	if _, err := run(t, []string{"-d", "3"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil || got.VT != 3 || !got.ReadOnly {
		t.Errorf("spy received unexpected options: %+v", got)
	}
}

func TestRunExtraArg(t *testing.T) {
	t.Parallel()
	if _, err := run(t, []string{"2", "3"}); err == nil {
		t.Error("expected error for extra argument")
	}
}

func TestHelp(t *testing.T) {
	t.Parallel()
	var out, errBuf bytes.Buffer
	io := command.IO{In: strings.NewReader(""), Out: &out, Err: &errBuf}
	if err := New().Run(context.Background(), io, []string{"--help"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "Usage: conspy") {
		t.Errorf("help missing usage:\n%s", out.String())
	}
}
