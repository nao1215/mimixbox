package lsscsi

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

// fixture builds a fake sysfs SCSI tree and returns its root.
func fixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	mk := func(hctl, typ, vendor, model, rev, blockName string) {
		dir := filepath.Join(root, hctl)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		write := func(name, val string) {
			if err := os.WriteFile(filepath.Join(dir, name), []byte(val+"\n"), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		write("type", typ)
		write("vendor", vendor)
		write("model", model)
		write("rev", rev)
		if blockName != "" {
			bdir := filepath.Join(dir, "block", blockName)
			if err := os.MkdirAll(bdir, 0o755); err != nil {
				t.Fatal(err)
			}
		}
	}
	// Out of order on purpose to exercise sorting.
	mk("2:0:0:0", "5", "HL-DT-ST", "DVD+-RW", "1.0", "")
	mk("0:0:0:0", "0", "ATA", "Samsung SSD", "2B6Q", "sda")
	// A host directory that is not a device should be ignored.
	if err := os.MkdirAll(filepath.Join(root, "host0"), 0o755); err != nil {
		t.Fatal(err)
	}
	return root
}

func withRoot(t *testing.T, root string) {
	t.Helper()
	prev := scsiDevices
	scsiDevices = root
	t.Cleanup(func() { scsiDevices = prev })
}

func run(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	var out, errBuf bytes.Buffer
	stdio := command.IO{In: strings.NewReader(""), Out: &out, Err: &errBuf}
	err := New().Run(context.Background(), stdio, args)
	return out.String(), errBuf.String(), err
}

func TestLsscsiList(t *testing.T) {
	withRoot(t, fixture(t))
	out, _, err := run(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 devices, got %d: %q", len(lines), out)
	}
	// Sorted: 0:0:0:0 before 2:0:0:0.
	if !strings.HasPrefix(lines[0], "[0:0:0:0]") {
		t.Errorf("first line not sorted: %q", lines[0])
	}
	if !strings.Contains(lines[0], "disk") || !strings.Contains(lines[0], "/dev/sda") {
		t.Errorf("disk line wrong: %q", lines[0])
	}
	if !strings.Contains(lines[1], "cd/dvd") || !strings.HasSuffix(lines[1], "-") {
		t.Errorf("cd line wrong: %q", lines[1])
	}
}

func TestLsscsiClassic(t *testing.T) {
	withRoot(t, fixture(t))
	out, _, err := run(t, "-c")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "(0x0)") {
		t.Errorf("classic type code missing: %q", out)
	}
}

func TestLsscsiMissingTree(t *testing.T) {
	withRoot(t, filepath.Join(t.TempDir(), "absent"))
	_, errOut, err := run(t)
	if err == nil {
		t.Fatal("expected error for missing sysfs tree")
	}
	if !strings.Contains(errOut, "lsscsi:") {
		t.Errorf("missing prefix: %q", errOut)
	}
}

func TestIsHCTL(t *testing.T) {
	cases := map[string]bool{
		"0:0:0:0": true,
		"6:0:0:1": true,
		"host0":   false,
		"0:0:0":   false,
		"a:0:0:0": false,
	}
	for in, want := range cases {
		if got := isHCTL(in); got != want {
			t.Errorf("isHCTL(%q)=%v want %v", in, got, want)
		}
	}
}
