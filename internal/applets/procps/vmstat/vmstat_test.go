package vmstat

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
	write := func(name, content string) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		return p
	}
	os, om, ov, ou := statPath, meminfoPath, vmPath, uptimePath
	meminfoPath = write("meminfo", "MemFree:        1000 kB\nBuffers:         200 kB\nCached:          300 kB\nSwapTotal:       500 kB\nSwapFree:        400 kB\n")
	statPath = write("stat", "cpu 100 0 50 800 50 0 0 0\nintr 2000 1 2 3\nctxt 4000\nprocs_running 3\nprocs_blocked 1\n")
	vmPath = write("vmstat", "pswpin 10\npswpout 20\npgpgin 400\npgpgout 200\n")
	uptimePath = write("uptime", "100.0 50.0\n")
	t.Cleanup(func() { statPath, meminfoPath, vmPath, uptimePath = os, om, ov, ou })
}

func run(t *testing.T) string {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, nil); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	return out.String()
}

func TestSnapshot(t *testing.T) {
	fixture(t)
	out := run(t)
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d:\n%s", len(lines), out)
	}
	// r=3 b=1 swpd=100 free=1000 buff=200 cache=300 si=0 so=0 bi=4 bo=2 in=20 cs=40 us=10 sy=5 id=80 wa=5 st=0
	fields := strings.Fields(lines[2])
	want := []string{"3", "1", "100", "1000", "200", "300", "0", "0", "4", "2", "20", "40", "10", "5", "80", "5", "0"}
	if len(fields) != len(want) {
		t.Fatalf("got %d columns: %q", len(fields), lines[2])
	}
	for i := range want {
		if fields[i] != want[i] {
			t.Errorf("column %d = %q, want %q (line: %q)", i, fields[i], want[i], lines[2])
		}
	}
}

func TestCPUPercents(t *testing.T) {
	t.Parallel()
	us, sy, id, wa, st := cpuPercents([]int64{100, 0, 50, 800, 50, 0, 0, 0})
	if us != 10 || sy != 5 || id != 80 || wa != 5 || st != 0 {
		t.Errorf("cpuPercents = %d %d %d %d %d, want 10 5 80 5 0", us, sy, id, wa, st)
	}
	// All-idle fallback when total is zero.
	if _, _, id, _, _ := cpuPercents(nil); id != 100 {
		t.Errorf("empty cpu idle = %d, want 100", id)
	}
}
