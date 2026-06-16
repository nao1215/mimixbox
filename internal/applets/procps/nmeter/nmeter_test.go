package nmeter

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
	write := func(name, content string) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		return p
	}
	os1, om, on := statPath, meminfoPath, now
	// cpu: total = 100+0+50+800+50 = 1000, idle = 800 -> busy 200/1000 = 20%.
	statPath = write("stat", "cpu 100 0 50 800 50 0 0 0\n")
	// MemTotal 8192 MiB-ish: 8388608 kB; MemAvailable 4194304 kB -> used 4194304 kB = 4096 MiB.
	meminfoPath = write("meminfo", "MemTotal:       8388608 kB\nMemAvailable:   4194304 kB\n")
	now = func() time.Time { return time.Date(2026, 6, 10, 9, 8, 7, 0, time.UTC) }
	t.Cleanup(func() { statPath, meminfoPath, now = os1, om, on })
}

func run(t *testing.T, format string) (string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	err := New().Run(context.Background(), io, []string{format})
	return strings.TrimRight(out.String(), "\n"), err
}

func TestExpand(t *testing.T) {
	fixture(t)
	out, err := run(t, "%t cpu:%c mem:%m/%M %%done")
	if err != nil {
		t.Fatal(err)
	}
	if out != "09:08:07 cpu:20% mem:4096M/8192M %done" {
		t.Errorf("nmeter = %q", out)
	}
}

func TestUnknownDirectivePassesThrough(t *testing.T) {
	fixture(t)
	out, err := run(t, "a%zb")
	if err != nil {
		t.Fatal(err)
	}
	if out != "a%zb" {
		t.Errorf("unknown directive = %q, want a%%zb", out)
	}
}

func TestNoFormat(t *testing.T) {
	fixture(t)
	io := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, nil); err == nil {
		t.Errorf("missing format should fail")
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
