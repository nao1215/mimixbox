package setuidgid

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

type runCall struct {
	uid, gid int
	prog     string
	args     []string
}

func setup(t *testing.T) *runCall {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "passwd")
	content := "root:x:0:0:root:/root:/bin/sh\nnobody:x:65534:65534:nobody:/:/usr/sbin/nologin\n"
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	call := &runCall{uid: -1}
	op, orf := passwdPath, runFn
	passwdPath = p
	runFn = func(_ context.Context, _ command.IO, uid, gid int, prog string, args []string) error {
		*call = runCall{uid, gid, prog, args}
		return nil
	}
	t.Cleanup(func() { passwdPath, runFn = op, orf })
	return call
}

func run(t *testing.T, args ...string) error {
	t.Helper()
	io := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	return New().Run(context.Background(), io, args)
}

func TestRunsAsUser(t *testing.T) {
	call := setup(t)
	if err := run(t, "nobody", "mydaemon", "--flag", "arg"); err != nil {
		t.Fatal(err)
	}
	if call.uid != 65534 || call.gid != 65534 {
		t.Errorf("uid/gid = %d/%d, want 65534/65534", call.uid, call.gid)
	}
	if call.prog != "mydaemon" || len(call.args) != 2 || call.args[0] != "--flag" {
		t.Errorf("prog/args = %q %v", call.prog, call.args)
	}
}

func TestErrors(t *testing.T) {
	setup(t)
	if err := run(t, "nobody"); err == nil {
		t.Errorf("missing program should fail")
	}
	if err := run(t, "ghost", "prog"); err == nil {
		t.Errorf("an unknown user should fail")
	}
	if err := run(t); err == nil {
		t.Errorf("no args should fail")
	}
}
