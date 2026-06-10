package ps

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func fixture(t *testing.T, stats map[string]string) {
	t.Helper()
	dir := t.TempDir()
	for pid, stat := range stats {
		pdir := filepath.Join(dir, pid)
		if err := os.MkdirAll(pdir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(pdir, "stat"), []byte(stat), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	orig := procDir
	procDir = dir
	t.Cleanup(func() { procDir = orig })
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

func TestListing(t *testing.T) {
	fixture(t, map[string]string{
		// tty_nr 34816 = pts/0; utime 150 + stime 50 = 200 jiffies = 2s.
		"100": "100 (bash) S 1 100 100 34816 100 0 0 0 0 0 150 50",
		"1":   "1 (init) S 0 1 1 0 0 0 0 0 0 0 0 0",
	})
	out := run(t)
	if !strings.HasPrefix(out, "    PID TTY          TIME CMD\n") {
		t.Errorf("header wrong: %q", out)
	}
	if !strings.Contains(out, "    100 pts/0    00:00:02 bash") {
		t.Errorf("bash row wrong:\n%s", out)
	}
	if !strings.Contains(out, "      1 ?        00:00:00 init") {
		t.Errorf("init row wrong:\n%s", out)
	}
	// Sorted by PID: 1 before 100.
	if strings.Index(out, " init") > strings.Index(out, " bash") {
		t.Errorf("not sorted by PID:\n%s", out)
	}
}

func TestTTYName(t *testing.T) {
	t.Parallel()
	cases := map[int]string{0: "?", 34816: "pts/0", 34817: "pts/1", 1024: "tty0"}
	for in, want := range cases {
		if got := ttyName(in); got != want {
			t.Errorf("ttyName(%d) = %q, want %q", in, got, want)
		}
	}
}

func TestCPUTime(t *testing.T) {
	t.Parallel()
	cases := map[int64]string{0: "00:00:00", 200: "00:00:02", 100 * 65: "00:01:05", 100 * 3661: "01:01:01"}
	for in, want := range cases {
		if got := cpuTime(in); got != want {
			t.Errorf("cpuTime(%d) = %q, want %q", in, got, want)
		}
	}
}
