package mpstat

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func fixture(t *testing.T, content string) {
	t.Helper()
	dir := t.TempDir()
	f := filepath.Join(dir, "stat")
	if err := os.WriteFile(f, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	orig := statPath
	statPath = f
	t.Cleanup(func() { statPath = orig })
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

func TestRows(t *testing.T) {
	fixture(t, "cpu 100 0 50 800 50 0 0 0 0 0\ncpu0 50 0 25 400 25 0 0 0 0 0\nintr 1\n")
	out := run(t)
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 3 { // header, all, cpu0
		t.Fatalf("expected 3 lines, got %d:\n%s", len(lines), out)
	}
	allFields := strings.Fields(lines[1])
	// all: usr 10.00 nice 0 sys 5 iowait 5 irq 0 soft 0 steal 0 guest 0 gnice 0 idle 80
	want := []string{"all", "10.00", "0.00", "5.00", "5.00", "0.00", "0.00", "0.00", "0.00", "0.00", "80.00"}
	for i := range want {
		if allFields[i] != want[i] {
			t.Errorf("all column %d = %q, want %q", i, allFields[i], want[i])
		}
	}
	if !strings.HasPrefix(lines[2], "0 ") {
		t.Errorf("per-cpu row label wrong: %q", lines[2])
	}
}

func TestPercentages(t *testing.T) {
	t.Parallel()
	p := percentages([]int64{100, 0, 50, 800, 50, 0, 0, 0, 0, 0})
	if p[0] != 10 || p[2] != 5 || p[3] != 5 || p[9] != 80 {
		t.Errorf("percentages = %v", p)
	}
	// guest time is subtracted from user.
	g := percentages([]int64{100, 0, 0, 900, 0, 0, 0, 0, 40, 0})
	if g[0] != 6 || g[7] != 4 { // usr = (100-40)/1000*100 = 6, guest = 40/1000*100 = 4
		t.Errorf("guest split: usr=%v guest=%v", g[0], g[7])
	}
	if percentages(nil)[9] != 100 {
		t.Errorf("empty should be 100%% idle")
	}
}

func TestMissingStat(t *testing.T) {
	t.Parallel()
	orig := statPath
	statPath = "/no/such/stat"
	defer func() { statPath = orig }()
	io := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, nil); err == nil {
		t.Errorf("missing stat should fail")
	}
}
