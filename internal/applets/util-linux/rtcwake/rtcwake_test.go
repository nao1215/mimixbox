package rtcwake

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

func fixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	alarm := filepath.Join(dir, "rtc0", "wakealarm")
	if err := os.MkdirAll(filepath.Dir(alarm), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(alarm, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	od, on := sysRTCDir, now
	sysRTCDir = dir
	now = func() time.Time { return time.Unix(1_000_000, 0) }
	t.Cleanup(func() { sysRTCDir, now = od, on })
	return alarm
}

func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	err := New().Run(context.Background(), io, args)
	return out.String(), err
}

func TestRelativeAlarm(t *testing.T) {
	alarm := fixture(t)
	out, err := run(t, "-m", "no", "-s", "300")
	if err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(alarm)
	if strings.TrimSpace(string(data)) != "1000300" { // 1_000_000 + 300
		t.Errorf("wakealarm = %q, want 1000300", data)
	}
	if !strings.Contains(out, "epoch 1000300") {
		t.Errorf("report = %q", out)
	}
}

func TestAbsoluteAlarm(t *testing.T) {
	alarm := fixture(t)
	if _, err := run(t, "-t", "1500000"); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(alarm)
	if strings.TrimSpace(string(data)) != "1500000" {
		t.Errorf("wakealarm = %q, want 1500000", data)
	}
}

func TestSuspendModeRejected(t *testing.T) {
	fixture(t)
	if _, err := run(t, "-m", "mem", "-s", "10"); err == nil {
		t.Errorf("suspend mode should be rejected")
	}
}

func TestErrors(t *testing.T) {
	fixture(t)
	if _, err := run(t, "-m", "no"); err == nil {
		t.Errorf("missing wake time should fail")
	}
	if _, err := run(t, "-m", "bogus", "-s", "1"); err == nil {
		t.Errorf("unknown mode should fail")
	}
}
