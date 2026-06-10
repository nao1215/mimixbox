package lspci

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func dev(t *testing.T, root, slot string, attrs map[string]string) {
	t.Helper()
	dir := filepath.Join(root, slot)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	for k, v := range attrs {
		if err := os.WriteFile(filepath.Join(dir, k), []byte(v+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func runFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	dev(t, root, "0000:00:1f.2", map[string]string{
		"class": "0x010601", "vendor": "0x8086", "device": "0x1234", "revision": "0x00",
	})
	dev(t, root, "0001:02:00.0", map[string]string{
		"class": "0x030000", "vendor": "0x10de", "device": "0x1c82", "revision": "0x05",
	})
	orig := pciDevices
	pciDevices = root
	t.Cleanup(func() { pciDevices = orig })

	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, nil); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	return out.String()
}

func TestNumericListing(t *testing.T) {
	out := runFixture(t)
	// Domain 0000 is dropped; class trimmed to 4 digits; no rev when 00.
	if !strings.Contains(out, "00:1f.2 0106: 8086:1234\n") {
		t.Errorf("first device line missing: %q", out)
	}
	// Non-default domain kept; revision shown.
	if !strings.Contains(out, "0001:02:00.0 0300: 10de:1c82 (rev 05)\n") {
		t.Errorf("second device line missing: %q", out)
	}
}

func TestSorted(t *testing.T) {
	out := runFixture(t)
	a := strings.Index(out, "00:1f.2")
	b := strings.Index(out, "0001:02:00.0")
	if a < 0 || b < 0 || a > b {
		t.Errorf("devices not sorted by slot: %q", out)
	}
}

func TestMissingDir(t *testing.T) {
	t.Parallel()
	orig := pciDevices
	pciDevices = "/no/such/pci/dir"
	defer func() { pciDevices = orig }()
	io := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, nil); err == nil {
		t.Errorf("missing sysfs dir should fail")
	}
}

func TestDisplaySlot(t *testing.T) {
	t.Parallel()
	if got := displaySlot("0000:00:00.0"); got != "00:00.0" {
		t.Errorf("displaySlot = %q", got)
	}
	if got := displaySlot("0001:00:00.0"); got != "0001:00:00.0" {
		t.Errorf("non-zero domain should be kept: %q", got)
	}
}
