package addgroup

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func fixture(t *testing.T, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "group")
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	orig := groupPath
	groupPath = p
	t.Cleanup(func() { groupPath = orig })
	return p
}

func run(t *testing.T, args ...string) error {
	t.Helper()
	io := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	return New().Run(context.Background(), io, args)
}

func lineFor(t *testing.T, path, name string) string {
	t.Helper()
	data, _ := os.ReadFile(path)
	for _, l := range strings.Split(strings.TrimRight(string(data), "\n"), "\n") {
		if strings.HasPrefix(l, name+":") {
			return l
		}
	}
	return ""
}

func TestAddAutoGID(t *testing.T) {
	p := fixture(t, "root:x:0:\nstaff:x:50:\n")
	if err := run(t, "developers"); err != nil {
		t.Fatal(err)
	}
	if got := lineFor(t, p, "developers"); got != "developers:x:1000:" {
		t.Errorf("added line = %q, want developers:x:1000:", got)
	}
}

func TestAddSpecificGID(t *testing.T) {
	p := fixture(t, "root:x:0:\n")
	if err := run(t, "--gid", "1500", "staff"); err != nil {
		t.Fatal(err)
	}
	if got := lineFor(t, p, "staff"); got != "staff:x:1500:" {
		t.Errorf("added line = %q", got)
	}
}

func TestDuplicateNameAndGID(t *testing.T) {
	fixture(t, "root:x:0:\nstaff:x:1500:\n")
	if err := run(t, "staff"); err == nil {
		t.Errorf("an existing group name should fail")
	}
	if err := run(t, "--gid", "1500", "other"); err == nil {
		t.Errorf("an in-use GID should fail")
	}
}

func TestNoTempLeft(t *testing.T) {
	p := fixture(t, "root:x:0:\n")
	if err := run(t, "newgrp"); err != nil {
		t.Fatal(err)
	}
	entries, _ := os.ReadDir(filepath.Dir(p))
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".addgroup-") {
			t.Errorf("temp file left behind: %s", e.Name())
		}
	}
}

func TestNoName(t *testing.T) {
	fixture(t, "root:x:0:\n")
	if err := run(t); err == nil {
		t.Errorf("missing group name should fail")
	}
}
