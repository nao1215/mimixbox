package resume

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

type fakeResolver struct {
	num string
	err error
}

func (f *fakeResolver) DevNumber(string) (string, error) { return f.num, f.err }

func withResolver(t *testing.T, r Resolver) {
	t.Helper()
	prev := resolver
	resolver = r
	t.Cleanup(func() { resolver = prev })
}

func withResumePath(t *testing.T, path string) {
	t.Helper()
	prev := resumePath
	resumePath = path
	t.Cleanup(func() { resumePath = prev })
}

func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	var out, errBuf bytes.Buffer
	stdio := command.IO{In: strings.NewReader(""), Out: &out, Err: &errBuf}
	err := New().Run(context.Background(), stdio, args)
	return out.String(), err
}

func TestResumeWritesDevNumber(t *testing.T) {
	withResolver(t, &fakeResolver{num: "8:2"})
	target := filepath.Join(t.TempDir(), "resume")
	withResumePath(t, target)

	out, err := run(t, "/dev/sda2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, rerr := os.ReadFile(target)
	if rerr != nil || string(data) != "8:2" {
		t.Errorf("resume target not written correctly: %q err=%v", data, rerr)
	}
	if !strings.Contains(out, "8:2") {
		t.Errorf("output missing dev number: %q", out)
	}
}

func TestResumeUsage(t *testing.T) {
	withResolver(t, &fakeResolver{num: "8:2"})
	if _, err := run(t); err == nil {
		t.Fatal("expected usage error with no device")
	}
}

func TestResumeResolverError(t *testing.T) {
	withResolver(t, &fakeResolver{err: errors.New("not a block device")})
	withResumePath(t, filepath.Join(t.TempDir(), "resume"))
	if _, err := run(t, "/dev/sda2"); err == nil {
		t.Fatal("expected error when resolver fails")
	}
}

func TestResumeWriteError(t *testing.T) {
	withResolver(t, &fakeResolver{num: "8:2"})
	// Point at a path whose parent does not exist so the write fails.
	withResumePath(t, filepath.Join(t.TempDir(), "absent-dir", "resume"))
	if _, err := run(t, "/dev/sda2"); err == nil {
		t.Fatal("expected error when sysfs write fails")
	}
}
