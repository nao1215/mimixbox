package pmap

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
	pdir := filepath.Join(dir, "1234")
	if err := os.MkdirAll(pdir, 0o755); err != nil {
		t.Fatal(err)
	}
	maps := "00400000-00410000 r-xp 00000000 08:01 123 /usr/bin/cat\n" +
		"7ffd00000000-7ffd00001000 rw-p 00000000 00:00 0 [stack]\n"
	if err := os.WriteFile(filepath.Join(pdir, "maps"), []byte(maps), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pdir, "cmdline"), []byte("cat\x00file\x00"), 0o644); err != nil {
		t.Fatal(err)
	}
	orig := procDir
	procDir = dir
	t.Cleanup(func() { procDir = orig })
}

func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	err := New().Run(context.Background(), io, args)
	return out.String(), err
}

func TestReport(t *testing.T) {
	fixture(t)
	out, err := run(t, "1234")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "1234:   cat file") {
		t.Errorf("header missing: %q", out)
	}
	// 0x10000 = 64K, r-x--, cat
	if !strings.Contains(out, "0000000000400000      64K r-x-- cat") {
		t.Errorf("cat mapping wrong: %q", out)
	}
	// 0x1000 = 4K stack
	if !strings.Contains(out, "rw--- [stack]") {
		t.Errorf("stack mapping wrong: %q", out)
	}
	// total 64 + 4 = 68K
	if !strings.Contains(out, "total           68K") {
		t.Errorf("total wrong: %q", out)
	}
}

func TestParseLine(t *testing.T) {
	t.Parallel()
	start, size, perms, name, ok := parseLine("00400000-00410000 r-xp 00000000 08:01 123 /usr/bin/cat")
	if !ok || start != 0x400000 || size != 64 || perms != "r-x--" || name != "cat" {
		t.Errorf("parseLine = %x, %d, %q, %q, %v", start, size, perms, name, ok)
	}
	if _, _, _, _, ok := parseLine("garbage"); ok {
		t.Errorf("garbage should not parse")
	}
}

func TestMode(t *testing.T) {
	t.Parallel()
	cases := map[string]string{"r-xp": "r-x--", "rw-p": "rw---", "rwxs": "rwxs-", "---p": "-----"}
	for in, want := range cases {
		if got := mode(in); got != want {
			t.Errorf("mode(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestErrors(t *testing.T) {
	fixture(t)
	if _, err := run(t, "notapid"); err == nil {
		t.Errorf("invalid PID should fail")
	}
	if _, err := run(t, "9999"); err == nil {
		t.Errorf("missing process should fail")
	}
	if _, err := run(t); err == nil {
		t.Errorf("no PID should fail")
	}
}
