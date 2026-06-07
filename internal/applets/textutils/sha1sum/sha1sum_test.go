package sha1sum_test

import (
	"bytes"
	"context"
	"crypto/sha1" //nolint:gosec // test exercises the sha1sum applet
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/textutils/sha1sum"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, stdin string, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(stdin), Out: out, Err: errBuf}
	err := sha1sum.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestNewAndMeta(t *testing.T) {
	t.Parallel()
	c := sha1sum.New()
	if c == nil {
		t.Fatal("New() returned nil")
	}
	if c.Name() != "sha1sum" {
		t.Errorf("Name() = %q", c.Name())
	}
	if c.Synopsis() != "Calculate or Check secure hash 1 algorithm" {
		t.Errorf("Synopsis() = %q", c.Synopsis())
	}
}

func TestDigest(t *testing.T) {
	t.Parallel()
	sum := sha1.Sum([]byte("abc\n")) //nolint:gosec // matches applet
	wantStdin := fmt.Sprintf("%x", sum) + "  -\n"

	out, _, err := run(t, "abc\n")
	if err != nil {
		t.Fatalf("stdin Run error = %v", err)
	}
	if out != wantStdin {
		t.Errorf("stdin out = %q, want %q", out, wantStdin)
	}

	dir := t.TempDir()
	f := filepath.Join(dir, "a.txt")
	_ = os.WriteFile(f, []byte("abc\n"), 0o600)
	out, _, err = run(t, "", f)
	if err != nil {
		t.Fatalf("file Run error = %v", err)
	}
	if out != fmt.Sprintf("%x", sum)+"  "+f+"\n" {
		t.Errorf("file out = %q", out)
	}
}

func TestMissingFile(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, "", "/no/such/file")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.HasPrefix(errOut, "sha1sum: /no/such/file:") {
		t.Errorf("stderr = %q", errOut)
	}
}
