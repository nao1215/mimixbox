package openvt

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, args []string) (string, string, error) {
	t.Helper()
	var out, errBuf bytes.Buffer
	io := command.IO{In: strings.NewReader(""), Out: &out, Err: &errBuf}
	err := New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestParseRequest(t *testing.T) {
	t.Parallel()
	r, err := ParseRequest("7", true, true, false, []string{"top", "-b"})
	if err != nil {
		t.Fatalf("ParseRequest: %v", err)
	}
	if r.VT != 7 || !r.Switch || !r.Wait {
		t.Errorf("unexpected request: %+v", r)
	}
	if len(r.Argv) != 2 || r.Argv[0] != "top" {
		t.Errorf("argv = %v", r.Argv)
	}
}

func TestParseRequestNoProgram(t *testing.T) {
	t.Parallel()
	if _, err := ParseRequest("", false, false, false, nil); err == nil {
		t.Error("expected error when no program given")
	}
}

func TestParseRequestBadVT(t *testing.T) {
	t.Parallel()
	for _, bad := range []string{"0", "64", "x"} {
		if _, err := ParseRequest(bad, false, false, false, []string{"sh"}); err == nil {
			t.Errorf("ParseRequest vt=%q expected error", bad)
		}
	}
}

func TestRunNoProgram(t *testing.T) {
	t.Parallel()
	if _, _, err := run(t, nil); err == nil {
		t.Error("expected error when no program given")
	}
}

func TestRunCapabilityError(t *testing.T) {
	t.Parallel()
	if _, _, err := run(t, []string{"--", "sh"}); err == nil {
		t.Error("expected capability error without a console")
	}
}

func TestRunBadVT(t *testing.T) {
	t.Parallel()
	if _, _, err := run(t, []string{"-c", "99", "--", "sh"}); err == nil {
		t.Error("expected error for bad VT")
	}
}

func TestRunInjectedSuccess(t *testing.T) {
	orig := runFn
	var got *Request
	runFn = func(r *Request) error { got = r; return nil }
	defer func() { runFn = orig }()

	if _, _, err := run(t, []string{"-c", "7", "-s", "--", "getty", "tty7"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil || got.VT != 7 || !got.Switch || got.Argv[0] != "getty" {
		t.Errorf("run received unexpected request: %+v", got)
	}
}

func TestHelp(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, []string{"--help"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Usage: openvt") {
		t.Errorf("help missing usage:\n%s", out)
	}
}
