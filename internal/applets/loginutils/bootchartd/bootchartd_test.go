package bootchartd

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func fixtures(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	write := func(name, content string) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		return p
	}
	ostat, odisk, oup, oout := statPath, diskstatsPath, uptimePath, outputDir
	statPath = write("stat", "cpu  100 0 50 800 0 0 0 0 0 0\n")
	diskstatsPath = write("diskstats", "8 0 sda 100 0 200 0\n")
	uptimePath = write("uptime", "123.45 600.00\n")
	outputDir = filepath.Join(dir, "out")
	t.Cleanup(func() { statPath, diskstatsPath, uptimePath, outputDir = ostat, odisk, oup, oout })
	return outputDir
}

func run(t *testing.T, args ...string) error {
	t.Helper()
	io := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	return New().Run(context.Background(), io, args)
}

func TestRecordsSample(t *testing.T) {
	out := fixtures(t)
	if err := run(t); err != nil {
		t.Fatal(err)
	}
	stat, err := os.ReadFile(filepath.Join(out, "proc_stat.log"))
	if err != nil {
		t.Fatal(err)
	}
	// Timestamp (uptime 123.45s -> 12345 jiffies) header followed by the content.
	if !strings.HasPrefix(string(stat), "12345\n") || !strings.Contains(string(stat), "cpu  100 0 50 800") {
		t.Errorf("proc_stat.log = %q", stat)
	}
	disk, _ := os.ReadFile(filepath.Join(out, "proc_diskstats.log"))
	if !strings.Contains(string(disk), "sda 100 0 200") {
		t.Errorf("proc_diskstats.log = %q", disk)
	}
}

func TestAppendsSamples(t *testing.T) {
	out := fixtures(t)
	if err := run(t); err != nil {
		t.Fatal(err)
	}
	if err := run(t); err != nil {
		t.Fatal(err)
	}
	stat, _ := os.ReadFile(filepath.Join(out, "proc_stat.log"))
	if n := strings.Count(string(stat), "12345\n"); n != 2 {
		t.Errorf("expected 2 samples, found %d:\n%s", n, stat)
	}
}

func TestOutputFlag(t *testing.T) {
	fixtures(t)
	custom := filepath.Join(t.TempDir(), "custom")
	if err := run(t, "-o", custom); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(custom, "proc_stat.log")); err != nil {
		t.Errorf("-o directory not used: %v", err)
	}
}
