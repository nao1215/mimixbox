package svlogd

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, stdin string, args ...string) error {
	t.Helper()
	io := command.IO{In: strings.NewReader(stdin), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	return New().Run(context.Background(), io, args)
}

func TestLogsToCurrent(t *testing.T) {
	dir := t.TempDir()
	if err := run(t, "line one\nline two\n", dir); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(filepath.Join(dir, "current"))
	if string(data) != "line one\nline two\n" {
		t.Errorf("current = %q", data)
	}
}

func TestTimestampPrefix(t *testing.T) {
	on := now
	now = func() time.Time { return time.Unix(1000, 0) }
	defer func() { now = on }()
	dir := t.TempDir()
	if err := run(t, "msg\n", "-t", dir); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(filepath.Join(dir, "current"))
	if !strings.HasSuffix(string(data), " msg\n") || len(data) < len("msg\n")+10 {
		t.Errorf("timestamped line = %q", data)
	}
}

func TestRotation(t *testing.T) {
	om, on := maxSize, now
	maxSize = 12 // rotate after a few bytes
	tick := int64(0)
	now = func() time.Time { tick++; return time.Unix(tick, 0) }
	defer func() { maxSize, now = om, on }()

	dir := t.TempDir()
	if err := run(t, "aaaa\nbbbb\ncccc\ndddd\n", dir); err != nil {
		t.Fatal(err)
	}
	entries, _ := os.ReadDir(dir)
	var rotated int
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "@") {
			rotated++
		}
	}
	if rotated == 0 {
		t.Errorf("expected at least one rotated file, got entries %v", names(entries))
	}
	// current must still exist and hold the most recent line(s).
	if _, err := os.Stat(filepath.Join(dir, "current")); err != nil {
		t.Errorf("current missing after rotation: %v", err)
	}
}

func names(entries []os.DirEntry) []string {
	var out []string
	for _, e := range entries {
		out = append(out, e.Name())
	}
	return out
}

func TestNoDir(t *testing.T) {
	if err := run(t, "x\n"); err == nil {
		t.Errorf("a missing directory should fail")
	}
}
