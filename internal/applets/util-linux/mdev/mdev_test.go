package mdev

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

type mknodCall struct {
	path         string
	isBlock      bool
	major, minor uint32
}

func fixture(t *testing.T, failMknod bool) *[]mknodCall {
	t.Helper()
	dir := t.TempDir()
	mk := func(class, name, dev string) {
		d := filepath.Join(dir, class, name)
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
		if dev != "" {
			if err := os.WriteFile(filepath.Join(d, "dev"), []byte(dev+"\n"), 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
	mk("block", "sda", "8:0")
	mk("mem", "null", "1:3")
	mk("tty", "tty0", "4:0")
	mk("net", "eth0", "") // no dev attribute -> skipped

	var calls []mknodCall
	osc, odv, omk := sysClassDir, devDir, mknodFn
	sysClassDir = dir
	devDir = filepath.Join(dir, "dev")
	mknodFn = func(path string, isBlock bool, major, minor uint32) error {
		if failMknod {
			return errors.New("permission denied")
		}
		calls = append(calls, mknodCall{path, isBlock, major, minor})
		return nil
	}
	t.Cleanup(func() { sysClassDir, devDir, mknodFn = osc, odv, omk })
	return &calls
}

func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	err := New().Run(context.Background(), io, args)
	return out.String(), err
}

func TestScanCreatesNodes(t *testing.T) {
	calls := fixture(t, false)
	out, err := run(t, "-s")
	if err != nil {
		t.Fatal(err)
	}
	if len(*calls) != 3 {
		t.Fatalf("created %d nodes, want 3: %+v", len(*calls), *calls)
	}
	sort.Slice(*calls, func(i, j int) bool { return (*calls)[i].path < (*calls)[j].path })
	byName := map[string]mknodCall{}
	for _, c := range *calls {
		byName[filepath.Base(c.path)] = c
	}
	if c := byName["sda"]; !c.isBlock || c.major != 8 || c.minor != 0 {
		t.Errorf("sda node = %+v, want block 8:0", c)
	}
	if c := byName["null"]; c.isBlock || c.major != 1 || c.minor != 3 {
		t.Errorf("null node = %+v, want char 1:3", c)
	}
	if c := byName["tty0"]; c.isBlock || c.major != 4 || c.minor != 0 {
		t.Errorf("tty0 node = %+v, want char 4:0", c)
	}
	if !strings.Contains(out, "created 3") {
		t.Errorf("summary = %q", out)
	}
}

func TestRequiresScan(t *testing.T) {
	fixture(t, false)
	if _, err := run(t); err == nil {
		t.Errorf("without -s it should fail")
	}
}

func TestMknodFailure(t *testing.T) {
	fixture(t, true)
	if _, err := run(t, "-s"); err == nil {
		t.Errorf("a mknod failure should fail")
	}
}

func TestReadDev(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	good := filepath.Join(dir, "dev")
	_ = os.WriteFile(good, []byte("259:5\n"), 0o644)
	if maj, min, ok := readDev(good); !ok || maj != 259 || min != 5 {
		t.Errorf("readDev good = %d,%d,%v", maj, min, ok)
	}
	bad := filepath.Join(dir, "bad")
	_ = os.WriteFile(bad, []byte("notanumber\n"), 0o644)
	if _, _, ok := readDev(bad); ok {
		t.Errorf("readDev should reject a malformed dev file")
	}
	if _, _, ok := readDev(filepath.Join(dir, "missing")); ok {
		t.Errorf("readDev should reject a missing file")
	}
}
