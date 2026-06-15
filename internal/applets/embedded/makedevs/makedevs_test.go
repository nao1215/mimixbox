package makedevs

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

// fakeNodeMaker records the device nodes it was asked to create instead of
// calling mknod, so tests run unprivileged and deterministically.
type fakeNodeMaker struct {
	nodes []string
	err   error
}

func (f *fakeNodeMaker) Mknod(path string, kind byte, mode os.FileMode, major, minor uint32) error {
	if f.err != nil {
		return f.err
	}
	f.nodes = append(f.nodes, fmt.Sprintf("%c %s %o %d:%d", kind, path, mode.Perm(), major, minor))
	return nil
}

func withNodeMaker(t *testing.T, nm NodeMaker) {
	t.Helper()
	prev := nodeMaker
	nodeMaker = nm
	t.Cleanup(func() { nodeMaker = prev })
}

func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	var out, errBuf bytes.Buffer
	stdio := command.IO{In: strings.NewReader(""), Out: &out, Err: &errBuf}
	err := New().Run(context.Background(), stdio, args)
	return errBuf.String(), err
}

func writeTable(t *testing.T, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "table.txt")
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestParseRow(t *testing.T) {
	e, err := parseRow("/dev/sda b 660 0 0 8 0 0 0 0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e.typ != 'b' || e.mode.Perm() != 0o660 || e.major != 8 {
		t.Errorf("unexpected entry: %+v", e)
	}
}

func TestParseRowErrors(t *testing.T) {
	rows := []string{
		"too few fields",
		"/dev/x z 660 0 0 0 0 0 0 0", // bad type
		"/dev/x c 8z8 0 0 0 0 0 0 0", // bad mode
		"/dev/x c 660 0 0 0 0 0 0 zz", // bad number
	}
	for _, r := range rows {
		if _, err := parseRow(r); err == nil {
			t.Errorf("expected error for row %q", r)
		}
	}
}

func TestMakedevsDirsAndFiles(t *testing.T) {
	withNodeMaker(t, &fakeNodeMaker{})
	table := writeTable(t, strings.Join([]string{
		"# comment",
		"",
		"/dev d 755 0 0 0 0 0 0 0",
		"/etc/hostname f 644 0 0 0 0 0 0 0",
	}, "\n"))
	root := filepath.Join(t.TempDir(), "rootfs")
	if _, err := run(t, "-d", table, root); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fi, err := os.Stat(filepath.Join(root, "dev")); err != nil || !fi.IsDir() {
		t.Errorf("dev directory not created: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "etc/hostname")); err != nil {
		t.Errorf("file not created: %v", err)
	}
}

func TestMakedevsDeviceNodes(t *testing.T) {
	nm := &fakeNodeMaker{}
	withNodeMaker(t, nm)
	table := writeTable(t, "/dev/sda b 660 0 0 8 0 0 0 0\n")
	root := filepath.Join(t.TempDir(), "rootfs")
	if _, err := run(t, "-d", table, root); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(nm.nodes) != 1 {
		t.Fatalf("expected 1 node, got %v", nm.nodes)
	}
	if !strings.Contains(nm.nodes[0], "b ") || !strings.Contains(nm.nodes[0], "8:0") {
		t.Errorf("node wrong: %q", nm.nodes[0])
	}
}

func TestMakedevsNumberedRange(t *testing.T) {
	nm := &fakeNodeMaker{}
	withNodeMaker(t, nm)
	// 3 nodes: tty0 tty1 tty2 with minors 0,1,2 (start=0, inc=1, count=3).
	table := writeTable(t, "/dev/tty c 666 0 0 4 0 0 1 3\n")
	root := filepath.Join(t.TempDir(), "rootfs")
	if _, err := run(t, "-d", table, root); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(nm.nodes) != 3 {
		t.Fatalf("expected 3 nodes, got %v", nm.nodes)
	}
	sort.Strings(nm.nodes)
	wantMinors := []string{"4:0", "4:1", "4:2"}
	for i, want := range wantMinors {
		if !strings.Contains(nm.nodes[i], want) {
			t.Errorf("node %d missing %s: %q", i, want, nm.nodes[i])
		}
	}
}

func TestMakedevsNodeErrorPropagates(t *testing.T) {
	nm := &fakeNodeMaker{err: errors.New("operation not permitted")}
	withNodeMaker(t, nm)
	table := writeTable(t, "/dev/sda b 660 0 0 8 0 0 0 0\n")
	root := filepath.Join(t.TempDir(), "rootfs")
	if _, err := run(t, "-d", table, root); err == nil {
		t.Fatal("expected error when mknod fails")
	}
}

func TestMakedevsUsage(t *testing.T) {
	withNodeMaker(t, &fakeNodeMaker{})
	if _, err := run(t, "./rootfs"); err == nil {
		t.Fatal("expected usage error without -d")
	}
}

func TestMakedevsStdinTable(t *testing.T) {
	withNodeMaker(t, &fakeNodeMaker{})
	var out, errBuf bytes.Buffer
	root := filepath.Join(t.TempDir(), "rootfs")
	stdio := command.IO{
		In:  strings.NewReader("/var d 755 0 0 0 0 0 0 0\n"),
		Out: &out,
		Err: &errBuf,
	}
	if err := New().Run(context.Background(), stdio, []string{"-d", "-", root}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fi, err := os.Stat(filepath.Join(root, "var")); err != nil || !fi.IsDir() {
		t.Errorf("dir from stdin table not created: %v", err)
	}
}
