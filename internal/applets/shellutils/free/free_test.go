package free

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

const sampleMeminfo = `MemTotal:        8000000 kB
MemFree:         5000000 kB
MemAvailable:    6000000 kB
Buffers:          100000 kB
Cached:           500000 kB
SReclaimable:     100000 kB
Shmem:             50000 kB
SwapTotal:       2000000 kB
SwapFree:        1500000 kB
`

func stubMeminfo(t *testing.T, content string) {
	t.Helper()
	orig := meminfoSource
	meminfoSource = func() (io.Reader, error) { return strings.NewReader(content), nil }
	t.Cleanup(func() { meminfoSource = orig })
}

func run(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestDefaultKibibytes(t *testing.T) {
	stubMeminfo(t, sampleMeminfo)
	out, _, err := run(t)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	// total=8000000, buff/cache=Buffers+Cached+SReclaimable=700000,
	// used=8000000-5000000-700000=2300000.
	if !strings.Contains(out, "8000000") {
		t.Errorf("missing total in %q", out)
	}
	if !strings.Contains(out, "2300000") {
		t.Errorf("missing computed used in %q", out)
	}
	if !strings.Contains(out, "Swap:") || !strings.Contains(out, "Mem:") {
		t.Errorf("missing rows in %q", out)
	}
}

func TestMebibytes(t *testing.T) {
	stubMeminfo(t, sampleMeminfo)
	out, _, err := run(t, "-m")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	// 8000000 kB / 1024 = 7812 MiB.
	if !strings.Contains(out, "7812") {
		t.Errorf("expected mebibyte total in %q", out)
	}
}

func TestBytes(t *testing.T) {
	stubMeminfo(t, sampleMeminfo)
	out, _, err := run(t, "-b")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	// 8000000 kB * 1024 = 8192000000 bytes.
	if !strings.Contains(out, "8192000000") {
		t.Errorf("expected byte total in %q", out)
	}
}

func TestHuman(t *testing.T) {
	stubMeminfo(t, sampleMeminfo)
	out, _, err := run(t, "-h")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(out, "Gi") && !strings.Contains(out, "Mi") {
		t.Errorf("expected a human-readable suffix in %q", out)
	}
}

func TestSwapUsed(t *testing.T) {
	stubMeminfo(t, sampleMeminfo)
	out, _, err := run(t)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	// swap used = SwapTotal - SwapFree = 2000000 - 1500000 = 500000.
	if !strings.Contains(out, "500000") {
		t.Errorf("missing swap used in %q", out)
	}
}

func TestHumanFormatter(t *testing.T) {
	t.Parallel()
	if got := human(512); got != "512B" {
		t.Errorf("human(512) = %q", got)
	}
	if got := human(1024); got != "1.0Ki" {
		t.Errorf("human(1024) = %q", got)
	}
	if got := human(1024 * 1024); got != "1.0Mi" {
		t.Errorf("human(1Mi) = %q", got)
	}
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := New()
	if c.Name() != "free" {
		t.Errorf("Name() = %q", c.Name())
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}
