package envuidgid

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func fixture(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "passwd")
	content := "root:x:0:0:root:/root:/bin/sh\nnobody:x:65534:65533:nobody:/:/usr/sbin/nologin\n"
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	orig := passwdPath
	passwdPath = p
	t.Cleanup(func() { passwdPath = orig })
}

func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	err := New().Run(context.Background(), io, args)
	return out.String(), err
}

func TestSetsUIDGID(t *testing.T) {
	fixture(t)
	out, err := run(t, "nobody", "sh", "-c", `printf '%s:%s' "$UID" "$GID"`)
	if err != nil {
		t.Fatal(err)
	}
	if out != "65534:65533" {
		t.Errorf("UID:GID = %q, want 65534:65533", out)
	}
}

func TestExitCodePropagates(t *testing.T) {
	fixture(t)
	_, err := run(t, "root", "sh", "-c", "exit 7")
	ee, ok := err.(*command.ExitError)
	if !ok || ee.Code != 7 {
		t.Errorf("err = %v, want exit 7", err)
	}
}

func TestErrors(t *testing.T) {
	fixture(t)
	if _, err := run(t, "root"); err == nil {
		t.Errorf("missing program should fail")
	}
	if _, err := run(t, "ghost", "true"); err == nil {
		t.Errorf("an unknown user should fail")
	}
}
