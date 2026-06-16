package uptime

import (
	"bytes"
	"context"
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
)

func setup(t *testing.T, up, load string, users int) {
	t.Helper()
	dir := t.TempDir()
	write := func(name, content string) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		return p
	}
	utmp := make([]byte, recordSize*users)
	for i := 0; i < users; i++ {
		binary.LittleEndian.PutUint16(utmp[i*recordSize:], userProcess)
	}
	uf := filepath.Join(dir, "utmp")
	if err := os.WriteFile(uf, utmp, 0o644); err != nil {
		t.Fatal(err)
	}

	ou, ol, oum, on := uptimePath, loadavgPath, utmpPath, now
	uptimePath = write("uptime", up)
	loadavgPath = write("loadavg", load)
	utmpPath = uf
	now = func() time.Time { return time.Date(2026, 6, 10, 21, 0, 0, 0, time.UTC) }
	t.Cleanup(func() { uptimePath, loadavgPath, utmpPath, now = ou, ol, oum, on })
}

func run(t *testing.T) string {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, nil); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	return strings.TrimRight(out.String(), "\n")
}

func TestUptimeLine(t *testing.T) {
	setup(t, "90000.0 1.0\n", "0.50 0.40 0.30 1/1 1\n", 2)
	out := run(t)
	if !strings.Contains(out, "load average: 0.50, 0.40, 0.30") {
		t.Errorf("load = %q", out)
	}
	if !strings.Contains(out, "2 users") {
		t.Errorf("users = %q", out)
	}
	if !strings.Contains(out, "up 1 day,  1:00") { // 90000s = 1 day 1:00
		t.Errorf("uptime = %q", out)
	}
}

func TestFormatUptime(t *testing.T) {
	t.Parallel()
	cases := map[time.Duration]string{
		90 * time.Minute: " 1:30",
		25 * time.Hour:   "1 day,  1:00",
		49 * time.Hour:   "2 days,  1:00",
	}
	for d, want := range cases {
		if got := formatUptime(d); got != want {
			t.Errorf("formatUptime(%v) = %q, want %q", d, got, want)
		}
	}
}

func TestHelpExitStatus(t *testing.T) {
	t.Parallel()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, []string{"--help"}); err != nil {
		t.Fatalf("Run --help error = %v", err)
	}
	if !strings.Contains(out.String(), "Exit status:") {
		t.Errorf("--help missing Exit status section:\n%s", out.String())
	}
}
