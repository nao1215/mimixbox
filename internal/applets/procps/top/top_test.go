package top

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
)

func fixture(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	mkproc := func(pid, stat, statm string) {
		pdir := filepath.Join(dir, pid)
		if err := os.MkdirAll(pdir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(pdir, "stat"), []byte(stat), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(pdir, "statm"), []byte(statm), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	// resident pages: bash 250 (=1000 KiB), vim 500 (=2000 KiB).
	mkproc("100", "100 (bash) S 1 100 100 0 0", "1000 250 0 0 0 0 0")
	mkproc("200", "200 (vim) R 1 200 200 0 0", "2000 500 0 0 0 0 0")

	write := func(name, content string) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		return p
	}
	op, om, ou, ol, on := procDir, meminfoPath, uptimePath, loadavgPath, now
	procDir = dir
	meminfoPath = write("meminfo", "MemTotal:       8000 kB\nMemFree:        4000 kB\nBuffers:         500 kB\nCached:         1000 kB\n")
	uptimePath = write("uptime", "3600.0 0.0\n")
	loadavgPath = write("loadavg", "0.50 0.40 0.30 1/1 1\n")
	now = func() time.Time { return time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC) }
	t.Cleanup(func() { procDir, meminfoPath, uptimePath, loadavgPath, now = op, om, ou, ol, on })
}

func run(t *testing.T) string {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, []string{"-b", "-n", "1"}); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	return out.String()
}

func TestSnapshot(t *testing.T) {
	fixture(t)
	out := run(t)
	if !strings.Contains(out, "up  1:00,") {
		t.Errorf("uptime wrong:\n%s", out)
	}
	if !strings.Contains(out, "load average: 0.50, 0.40, 0.30") {
		t.Errorf("load wrong:\n%s", out)
	}
	if !strings.Contains(out, "Tasks: 2 total, 1 running, 1 sleeping") {
		t.Errorf("tasks wrong:\n%s", out)
	}
	if !strings.Contains(out, "MiB Mem :") || !strings.Contains(out, "buff/cache") {
		t.Errorf("mem line wrong:\n%s", out)
	}
	// vim (RES 2000, 25.0%) sorts before bash (RES 1000, 12.5%).
	if strings.Index(out, "vim") > strings.Index(out, "bash") {
		t.Errorf("not sorted by RES desc:\n%s", out)
	}
	if !strings.Contains(out, "25.0") || !strings.Contains(out, "12.5") {
		t.Errorf("%%MEM wrong:\n%s", out)
	}
}
