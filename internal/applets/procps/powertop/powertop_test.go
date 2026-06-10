package powertop

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func fixture(t *testing.T, devices map[string]map[string]string) {
	t.Helper()
	dir := t.TempDir()
	for name, attrs := range devices {
		ddir := filepath.Join(dir, name)
		if err := os.MkdirAll(ddir, 0o755); err != nil {
			t.Fatal(err)
		}
		for k, v := range attrs {
			if err := os.WriteFile(filepath.Join(ddir, k), []byte(v+"\n"), 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
	orig := powerSupplyDir
	powerSupplyDir = dir
	t.Cleanup(func() { powerSupplyDir = orig })
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
	fixture(t, map[string]map[string]string{
		"AC":   {"type": "Mains", "online": "1"},
		"BAT0": {"type": "Battery", "capacity": "85", "status": "Discharging"},
	})
	out := run(t)
	if !strings.Contains(out, "AC (AC): online") {
		t.Errorf("AC line wrong:\n%s", out)
	}
	if !strings.Contains(out, "BAT0 (Battery): 85% (Discharging)") {
		t.Errorf("battery line wrong:\n%s", out)
	}
}

func TestOfflineAC(t *testing.T) {
	fixture(t, map[string]map[string]string{"AC": {"type": "Mains", "online": "0"}})
	if out := run(t); !strings.Contains(out, "AC (AC): offline") {
		t.Errorf("offline AC wrong:\n%s", out)
	}
}

func TestNoSupplies(t *testing.T) {
	fixture(t, map[string]map[string]string{})
	if out := run(t); !strings.Contains(out, "no power supplies found") {
		t.Errorf("empty report wrong:\n%s", out)
	}
}

func TestMissingDir(t *testing.T) {
	orig := powerSupplyDir
	powerSupplyDir = "/no/such/power"
	defer func() { powerSupplyDir = orig }()
	if out := run(t); !strings.Contains(out, "no power supplies found") {
		t.Errorf("missing dir should report none:\n%s", out)
	}
}
