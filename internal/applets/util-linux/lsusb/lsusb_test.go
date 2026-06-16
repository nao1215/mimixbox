package lsusb

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func dev(t *testing.T, root, name string, attrs map[string]string) {
	t.Helper()
	dir := filepath.Join(root, name)
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
	dev(t, root, "usb1", map[string]string{
		"idVendor": "1d6b", "idProduct": "0002", "busnum": "1", "devnum": "1",
		"manufacturer": "Linux Foundation", "product": "2.0 root hub",
	})
	dev(t, root, "1-1", map[string]string{
		"idVendor": "046d", "idProduct": "c52b", "busnum": "1", "devnum": "5",
		"manufacturer": "Logitech", "product": "USB Receiver",
	})
	// An interface node without idVendor must be skipped.
	dev(t, root, "1-1:1.0", map[string]string{"bInterfaceClass": "03"})

	orig := usbDevices
	usbDevices = root
	t.Cleanup(func() { usbDevices = orig })

	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, nil); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	return out.String()
}

func TestListing(t *testing.T) {
	out := runFixture(t)
	if !strings.Contains(out, "Bus 001 Device 001: ID 1d6b:0002 Linux Foundation 2.0 root hub\n") {
		t.Errorf("root hub line missing: %q", out)
	}
	if !strings.Contains(out, "Bus 001 Device 005: ID 046d:c52b Logitech USB Receiver\n") {
		t.Errorf("device line missing: %q", out)
	}
	// Exactly two device lines (the interface node is skipped).
	if n := strings.Count(out, "\n"); n != 2 {
		t.Errorf("expected 2 lines, got %d: %q", n, out)
	}
}

func TestSortedByBusThenDevice(t *testing.T) {
	out := runFixture(t)
	if strings.Index(out, "Device 001") > strings.Index(out, "Device 005") {
		t.Errorf("devices not sorted: %q", out)
	}
}

func TestMissingDir(t *testing.T) {
	t.Parallel()
	orig := usbDevices
	usbDevices = "/no/such/usb/dir"
	defer func() { usbDevices = orig }()
	io := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, nil); err == nil {
		t.Errorf("missing sysfs dir should fail")
	}
}

func TestHelpExitStatus(t *testing.T) {
	t.Parallel()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, []string{"--help"}); err != nil {
		t.Fatalf("--help Run error = %v", err)
	}
	if !strings.Contains(out.String(), "Exit status:") {
		t.Errorf("--help missing exit status section = %q", out.String())
	}
}
