package lsmod

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

const sample = `ext4 1024000 1 - Live 0xffffffffc0000000
mbcache 16384 1 ext4, Live 0xffffffffc0010000
loop 32768 0 - Live 0xffffffffc0020000
`

func TestParseModules(t *testing.T) {
	t.Parallel()
	mods, err := parseModules(strings.NewReader(sample))
	if err != nil {
		t.Fatalf("parseModules: %v", err)
	}
	if len(mods) != 3 {
		t.Fatalf("got %d modules, want 3", len(mods))
	}
	if mods[0].name != "ext4" || mods[0].size != 1024000 || mods[0].usedBy != 1 {
		t.Errorf("ext4 parsed wrong: %+v", mods[0])
	}
	if len(mods[1].usedSet) != 1 || mods[1].usedSet[0] != "ext4" {
		t.Errorf("mbcache deps wrong: %+v", mods[1].usedSet)
	}
	if len(mods[2].usedSet) != 0 {
		t.Errorf("loop should have no deps: %+v", mods[2].usedSet)
	}
}

func TestParseMalformed(t *testing.T) {
	t.Parallel()
	if _, err := parseModules(strings.NewReader("oops\n")); err == nil {
		t.Fatal("expected error for malformed line")
	}
}

func TestRunWithFixture(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "modules")
	if err := os.WriteFile(path, []byte(sample), 0o600); err != nil {
		t.Fatal(err)
	}
	old := modulesPath
	modulesPath = path
	defer func() { modulesPath = old }()

	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, nil); err != nil {
		t.Fatalf("Run: %v", err)
	}
	s := out.String()
	if !strings.Contains(s, "Module") || !strings.Contains(s, "ext4") || !strings.Contains(s, "mbcache") {
		t.Errorf("output missing expected rows:\n%s", s)
	}
}

func TestRunMissingFile(t *testing.T) {
	old := modulesPath
	modulesPath = filepath.Join(t.TempDir(), "nope")
	defer func() { modulesPath = old }()

	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: errBuf}
	if err := New().Run(context.Background(), io, nil); err == nil {
		t.Fatal("expected error when /proc/modules is missing")
	}
	if !strings.Contains(errBuf.String(), "cannot read") {
		t.Errorf("expected diagnostic, got %q", errBuf.String())
	}
}
