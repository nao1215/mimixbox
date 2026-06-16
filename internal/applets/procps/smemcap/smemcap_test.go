package smemcap

import (
	"archive/tar"
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func fixture(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	pdir := filepath.Join(dir, "100")
	if err := os.MkdirAll(pdir, 0o755); err != nil {
		t.Fatal(err)
	}
	for name, content := range map[string]string{
		"smaps":   "00400000-00410000 r-xp ...\n",
		"stat":    "100 (bash) S 1 100\n",
		"cmdline": "bash\x00",
	} {
		if err := os.WriteFile(filepath.Join(pdir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write := func(name, content string) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		return p
	}
	op, om, ov := procDir, meminfoPath, versionPath
	procDir = dir
	meminfoPath = write("meminfo", "MemTotal: 8000 kB\n")
	versionPath = write("version", "Linux version 6.0\n")
	t.Cleanup(func() { procDir, meminfoPath, versionPath = op, om, ov })
}

func capture(t *testing.T) map[string]string {
	t.Helper()
	out := &bytes.Buffer{}
	io2 := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io2, nil); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	files := map[string]string{}
	tr := tar.NewReader(out)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		data, _ := io.ReadAll(tr)
		files[hdr.Name] = string(data)
	}
	return files
}

func TestCapture(t *testing.T) {
	fixture(t)
	files := capture(t)
	for _, name := range []string{"meminfo", "version", "100/smaps", "100/stat", "100/cmdline"} {
		if _, ok := files[name]; !ok {
			t.Errorf("archive missing %q (have %v)", name, keys(files))
		}
	}
	if files["meminfo"] != "MemTotal: 8000 kB\n" {
		t.Errorf("meminfo content = %q", files["meminfo"])
	}
	if files["100/stat"] != "100 (bash) S 1 100\n" {
		t.Errorf("stat content = %q", files["100/stat"])
	}
}

func TestMissingFilesSkipped(t *testing.T) {
	fixture(t)
	// Remove the process smaps; the capture should still succeed without it.
	if err := os.Remove(filepath.Join(procDir, "100", "smaps")); err != nil {
		t.Fatal(err)
	}
	files := capture(t)
	if _, ok := files["100/smaps"]; ok {
		t.Errorf("missing smaps should be skipped")
	}
	if _, ok := files["100/stat"]; !ok {
		t.Errorf("stat should still be captured")
	}
}

func keys(m map[string]string) []string {
	var out []string
	for k := range m {
		out = append(out, k)
	}
	return out
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
