package w

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

func record(user, line, host string, login int32) []byte {
	b := make([]byte, recordSize)
	binary.LittleEndian.PutUint16(b[typeOffset:], userProcess)
	copy(b[lineOffset:lineOffset+fieldLen], line)
	copy(b[userOffset:userOffset+fieldLen], user)
	copy(b[hostOffset:hostOffset+hostLen], host)
	binary.LittleEndian.PutUint32(b[tvSecOffset:], uint32(login))
	return b
}

func setup(t *testing.T, recs []byte, loadavg, uptime string) {
	t.Helper()
	dir := t.TempDir()
	write := func(name, content string, data []byte) string {
		p := filepath.Join(dir, name)
		var b []byte
		if data != nil {
			b = data
		} else {
			b = []byte(content)
		}
		if err := os.WriteFile(p, b, 0o644); err != nil {
			t.Fatal(err)
		}
		return p
	}
	origU, origL, origUp, origNow := utmpPath, loadavgPath, uptimePath, now
	utmpPath = write("utmp", "", recs)
	loadavgPath = write("loadavg", loadavg, nil)
	uptimePath = write("uptime", uptime, nil)
	now = func() time.Time { return time.Date(2026, 6, 9, 20, 45, 0, 0, time.UTC) }
	t.Cleanup(func() {
		utmpPath, loadavgPath, uptimePath, now = origU, origL, origUp, origNow
	})
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

func TestHeaderAndRows(t *testing.T) {
	// Not parallel: setup() mutates package-level paths and the clock.
	recs := append(record("alice", "pts/0", "10.0.0.1", 0), record("bob", "tty1", "", 0)...)
	setup(t, recs, "0.50 0.40 0.30 1/100 999\n", "90000.0 1.0\n")

	out := run(t)
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if !strings.Contains(lines[0], "load average: 0.50, 0.40, 0.30") {
		t.Errorf("header load = %q", lines[0])
	}
	if !strings.Contains(lines[0], "2 users") {
		t.Errorf("header users = %q", lines[0])
	}
	// 90000s = 1 day, 1:00.
	if !strings.Contains(lines[0], "up 1 day,  1:00") {
		t.Errorf("header uptime = %q", lines[0])
	}
	if !strings.HasPrefix(lines[1], "USER") {
		t.Errorf("column header = %q", lines[1])
	}
	if !strings.Contains(out, "alice") || !strings.Contains(out, "pts/0") || !strings.Contains(out, "10.0.0.1") {
		t.Errorf("alice row missing: %q", out)
	}
	// bob has no host -> "-".
	if !strings.Contains(out, "bob") {
		t.Errorf("bob row missing: %q", out)
	}
}

func TestNoUsers(t *testing.T) {
	// Not parallel: setup() mutates package-level paths and the clock.
	setup(t, nil, "0.00 0.00 0.00 1/1 1\n", "60.0 1.0\n")
	out := run(t)
	if !strings.Contains(out, "0 users") {
		t.Errorf("expected 0 users, got %q", out)
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
		t.Fatalf("--help err = %v", err)
	}
	if !strings.Contains(out.String(), "Exit status:") {
		t.Errorf("--help missing Exit status section = %q", out.String())
	}
}
