package chpst

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
	"golang.org/x/sys/unix"
)

func setup(t *testing.T) (*procSpec, map[int]uint64) {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "passwd")
	if err := os.WriteFile(p, []byte("nobody:x:65534:65534:nobody:/:/bin/false\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	spec := &procSpec{}
	limits := map[int]uint64{}
	op, osr, orf := passwdPath, setRlimitFn, runFn
	passwdPath = p
	setRlimitFn = func(resource int, value uint64) error { limits[resource] = value; return nil }
	runFn = func(_ context.Context, _ command.IO, s procSpec) error { *spec = s; return nil }
	t.Cleanup(func() { passwdPath, setRlimitFn, runFn = op, osr, orf })
	return spec, limits
}

func run(t *testing.T, args ...string) error {
	t.Helper()
	io := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	return New().Run(context.Background(), io, args)
}

func TestUserSetsCreds(t *testing.T) {
	spec, _ := setup(t)
	if err := run(t, "-u", "nobody", "mydaemon"); err != nil {
		t.Fatal(err)
	}
	if !spec.setCreds || spec.uid != 65534 || spec.gid != 65534 {
		t.Errorf("creds = %+v", *spec)
	}
	if spec.prog != "mydaemon" {
		t.Errorf("prog = %q", spec.prog)
	}
}

func TestUserGroupOverride(t *testing.T) {
	spec, _ := setup(t)
	if err := run(t, "-u", "nobody:100", "p"); err != nil {
		t.Fatal(err)
	}
	if spec.uid != 65534 || spec.gid != 100 {
		t.Errorf("uid/gid = %d/%d, want 65534/100", spec.uid, spec.gid)
	}
}

func TestEnvdir(t *testing.T) {
	spec, _ := setup(t)
	d := t.TempDir()
	_ = os.WriteFile(filepath.Join(d, "FOO"), []byte("bar\n"), 0o644)
	if err := run(t, "-e", d, "p"); err != nil {
		t.Fatal(err)
	}
	var has bool
	for _, kv := range spec.env {
		if kv == "FOO=bar" {
			has = true
		}
	}
	if !has {
		t.Errorf("env not loaded from -e dir")
	}
}

func TestLimitsAndNice(t *testing.T) {
	spec, limits := setup(t)
	if err := run(t, "-o", "64", "-n", "5", "p"); err != nil {
		t.Fatal(err)
	}
	if limits[unix.RLIMIT_NOFILE] != 64 {
		t.Errorf("NOFILE limit = %d, want 64", limits[unix.RLIMIT_NOFILE])
	}
	if !spec.setNice || spec.nice != 5 {
		t.Errorf("nice = %+v", *spec)
	}
}

func TestErrors(t *testing.T) {
	setup(t)
	if err := run(t, "-u", "nobody"); err == nil {
		t.Errorf("missing program should fail")
	}
	if err := run(t, "-u", "ghost", "p"); err == nil {
		t.Errorf("an unknown user should fail")
	}
	if err := run(t); err == nil {
		t.Errorf("no args should fail")
	}
}
