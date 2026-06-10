package sysctl

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
	write := func(rel, val string) {
		p := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(val+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("kernel/ostype", "Linux")
	write("net/ipv4/ip_forward", "0")

	orig := sysDir
	sysDir = dir
	t.Cleanup(func() { sysDir = orig })
}

func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	err := New().Run(context.Background(), io, args)
	return out.String(), err
}

func TestRead(t *testing.T) {
	fixture(t)
	out, err := run(t, "kernel.ostype")
	if err != nil {
		t.Fatal(err)
	}
	if out != "kernel.ostype = Linux\n" {
		t.Errorf("read = %q", out)
	}
}

func TestListAll(t *testing.T) {
	fixture(t)
	out, err := run(t, "-a")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "kernel.ostype = Linux") || !strings.Contains(out, "net.ipv4.ip_forward = 0") {
		t.Errorf("-a = %q", out)
	}
}

func TestWrite(t *testing.T) {
	fixture(t)
	out, err := run(t, "net.ipv4.ip_forward=1")
	if err != nil {
		t.Fatal(err)
	}
	if out != "net.ipv4.ip_forward = 1\n" {
		t.Errorf("write output = %q", out)
	}
	data, _ := os.ReadFile(filepath.Join(sysDir, "net/ipv4/ip_forward"))
	if string(data) != "1" {
		t.Errorf("file = %q, want 1", data)
	}
}

func TestReadMissing(t *testing.T) {
	fixture(t)
	if _, err := run(t, "no.such.param"); err == nil {
		t.Errorf("missing parameter should fail")
	}
}

func TestNoArgs(t *testing.T) {
	fixture(t)
	if _, err := run(t); err == nil {
		t.Errorf("no arguments should fail")
	}
}
