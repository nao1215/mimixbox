package md5sum_test

import (
	"bytes"
	"context"
	"crypto/md5" //nolint:gosec // test exercises the md5sum applet
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/textutils/md5sum"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, stdin string, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(stdin), Out: out, Err: errBuf}
	err := md5sum.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestNewAndMeta(t *testing.T) {
	t.Parallel()
	c := md5sum.New()
	if c == nil {
		t.Fatal("New() returned nil")
	}
	if c.Name() != "md5sum" {
		t.Errorf("Name() = %q", c.Name())
	}
	if c.Synopsis() != "Calculate or Check md5sum message digest" {
		t.Errorf("Synopsis() = %q", c.Synopsis())
	}
}

func TestDigest(t *testing.T) {
	t.Parallel()
	sum := md5.Sum([]byte("abc\n")) //nolint:gosec // matches applet
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
	if !strings.HasPrefix(errOut, "md5sum: /no/such/file:") {
		t.Errorf("stderr = %q", errOut)
	}
}
