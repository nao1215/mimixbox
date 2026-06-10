package setpriv

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func withStubs(t *testing.T) {
	t.Helper()
	og, oeg, ogg, oeg2, oggr, onnp := getuid, geteuid, getgid, getegid, getgroups, noNewPrivs
	getuid = func() int { return 1000 }
	geteuid = func() int { return 1000 }
	getgid = func() int { return 1000 }
	getegid = func() int { return 1000 }
	getgroups = func() ([]int, error) { return []int{4, 1000}, nil }
	noNewPrivs = func() int { return 0 }
	t.Cleanup(func() {
		getuid, geteuid, getgid, getegid, getgroups, noNewPrivs = og, oeg, ogg, oeg2, oggr, onnp
	})
}

func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	err := New().Run(context.Background(), io, args)
	return out.String(), err
}

func TestDump(t *testing.T) {
	withStubs(t)
	out, err := run(t, "--dump")
	if err != nil {
		t.Fatal(err)
	}
	want := "uid: 1000\neuid: 1000\ngid: 1000\negid: 1000\nSupplementary groups: 4,1000\nno_new_privs: 0\n"
	if out != want {
		t.Errorf("dump =\n%q\nwant\n%q", out, want)
	}
}

func TestRunCommand(t *testing.T) {
	if _, err := exec.LookPath("echo"); err != nil {
		t.Skipf("echo not on PATH: %v", err)
	}
	withStubs(t)
	out, err := run(t, "--no-new-privs", "--", "echo", "hi")
	if err != nil {
		t.Fatalf("run error = %v", err)
	}
	if !strings.Contains(out, "hi") {
		t.Errorf("command output = %q", out)
	}
}

func TestMissingCommand(t *testing.T) {
	withStubs(t)
	if _, err := run(t); err == nil {
		t.Errorf("missing command should fail")
	}
}

func TestJoinInts(t *testing.T) {
	t.Parallel()
	if got := joinInts([]int{1, 2, 3}); got != "1,2,3" {
		t.Errorf("joinInts = %q", got)
	}
	if got := joinInts(nil); got != "[none]" {
		t.Errorf("joinInts(nil) = %q", got)
	}
}
