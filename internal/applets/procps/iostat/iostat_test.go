package iostat

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
	os1, od, ou := statPath, diskstatsPath, uptimePath
	statPath = write("stat", "cpu 100 0 50 800 50 0 0 0 0 0\n")
	// sda: 4 reads, 200 sectors read; 2 writes, 100 sectors written.
	diskstatsPath = write("diskstats",
		"   8       0 sda 4 0 200 0 2 0 100 0 0 0 0\n"+
			"   1       0 ram0 0 0 0 0 0 0 0 0 0 0 0\n")
	uptimePath = write("uptime", "100.0 0.0\n")
	t.Cleanup(func() { statPath, diskstatsPath, uptimePath = os1, od, ou })
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

func TestReport(t *testing.T) {
	fixture(t)
	out := run(t)
	// avg-cpu: user 10, nice 0, system 5, iowait 5, steal 0, idle 80.
	if !strings.Contains(out, "avg-cpu:  %user") {
		t.Errorf("missing avg-cpu header:\n%s", out)
	}
	if !strings.Contains(out, "   10.00    0.00    5.00    5.00    0.00   80.00") {
		t.Errorf("avg-cpu values wrong:\n%s", out)
	}
	// sda: kB_read = 200/2 = 100, kB_wrtn = 100/2 = 50; tps = (4+2)/100 = 0.06.
	if !strings.Contains(out, "sda") || !strings.Contains(out, "       100         50") {
		t.Errorf("sda totals wrong:\n%s", out)
	}
	// ram0 has no activity and must be omitted.
	if strings.Contains(out, "ram0") {
		t.Errorf("inactive device should be omitted:\n%s", out)
	}
}

func TestCPUPercents(t *testing.T) {
	t.Parallel()
	fixture(t)
	u, n, s, w, st, id := cpuPercents()
	if u != 10 || n != 0 || s != 5 || w != 5 || st != 0 || id != 80 {
		t.Errorf("cpuPercents = %v %v %v %v %v %v", u, n, s, w, st, id)
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
