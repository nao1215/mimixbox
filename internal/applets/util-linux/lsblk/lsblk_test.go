package lsblk

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func writeAttr(t *testing.T, dir, name, val string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name), []byte(val+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func fixture(t *testing.T) {
	t.Helper()
	root := t.TempDir()
	block := filepath.Join(root, "block")

	// sda: 1 GiB disk with one 512 MiB partition.
	sda := filepath.Join(block, "sda")
	writeAttr(t, sda, "size", "2097152") // *512 = 1 GiB
	writeAttr(t, sda, "ro", "0")
	writeAttr(t, sda, "removable", "0")
	writeAttr(t, sda, "dev", "8:0")
	sda1 := filepath.Join(sda, "sda1")
	writeAttr(t, sda1, "partition", "1")
	writeAttr(t, sda1, "size", "1048576") // *512 = 512 MiB
	writeAttr(t, sda1, "ro", "0")
	writeAttr(t, sda1, "dev", "8:1")

	// ram0 and an empty loop, both hidden by default.
	ram := filepath.Join(block, "ram0")
	writeAttr(t, ram, "size", "8192")
	writeAttr(t, ram, "dev", "1:0")
	loop := filepath.Join(block, "loop0")
	writeAttr(t, loop, "size", "0")
	writeAttr(t, loop, "dev", "7:0")

	mounts := filepath.Join(root, "mounts")
	if err := os.WriteFile(mounts, []byte("/dev/sda1 /mnt ext4 rw 0 0\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	origB, origM := sysBlock, procMounts
	sysBlock, procMounts = block, mounts
	t.Cleanup(func() { sysBlock, procMounts = origB, origM })
}

func run(t *testing.T, args ...string) string {
	t.Helper()
	fixture(t)
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, args); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	return out.String()
}

func TestDiskAndPartition(t *testing.T) {
	out := run(t)
	if !strings.Contains(out, "sda") || !strings.Contains(out, "1G") || !strings.Contains(out, "disk") {
		t.Errorf("disk row = %q", out)
	}
	if !strings.Contains(out, "sda1") || !strings.Contains(out, "512M") || !strings.Contains(out, "part") {
		t.Errorf("partition row = %q", out)
	}
	if !strings.Contains(out, "/mnt") {
		t.Errorf("mountpoint missing: %q", out)
	}
	// The partition is drawn as a tree child.
	if !strings.Contains(out, "─sda1") {
		t.Errorf("tree prefix missing: %q", out)
	}
}

func TestHidesRamAndEmptyByDefault(t *testing.T) {
	out := run(t)
	if strings.Contains(out, "ram0") || strings.Contains(out, "loop0") {
		t.Errorf("ram/empty devices should be hidden: %q", out)
	}
}

func TestAllShowsEverything(t *testing.T) {
	out := run(t, "-a")
	if !strings.Contains(out, "ram0") || !strings.Contains(out, "loop0") {
		t.Errorf("-a should show all devices: %q", out)
	}
}

func TestHumanSize(t *testing.T) {
	t.Parallel()
	cases := map[int64]string{
		0:                  "0B",
		512:                "512B",
		1024:               "1K",
		1073741824:         "1G",
		382867200: "365.1M",
	}
	for in, want := range cases {
		if got := humanSize(in); got != want {
			t.Errorf("humanSize(%d) = %q, want %q", in, got, want)
		}
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
