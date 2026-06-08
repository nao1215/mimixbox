package install_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/shellutils/install"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := install.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := install.New()
	if got := c.Name(); got != "install" {
		t.Errorf("Name() = %q, want %q", got, "install")
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}

func TestRunCopyToFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	dst := filepath.Join(dir, "dst")
	writeFile(t, src, "hello")

	if _, errOut, err := run(t, "-m", "640", src, dst); err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}

	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "hello" {
		t.Errorf("content = %q, want %q", got, "hello")
	}
	info, err := os.Stat(dst)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o640 {
		t.Errorf("mode = %o, want 640", info.Mode().Perm())
	}
}

func TestRunDefaultMode(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	dst := filepath.Join(dir, "dst")
	writeFile(t, src, "x")

	if _, _, err := run(t, src, dst); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	info, err := os.Stat(dst)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o755 {
		t.Errorf("default mode = %o, want 755", info.Mode().Perm())
	}
}

func TestRunCopyIntoDirectory(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	destDir := filepath.Join(dir, "bin")
	if err := os.Mkdir(destDir, 0o755); err != nil {
		t.Fatal(err)
	}
	a := filepath.Join(dir, "a")
	b := filepath.Join(dir, "b")
	writeFile(t, a, "A")
	writeFile(t, b, "B")

	if _, errOut, err := run(t, a, b, destDir); err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}
	for name, want := range map[string]string{"a": "A", "b": "B"} {
		got, err := os.ReadFile(filepath.Join(destDir, name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		if string(got) != want {
			t.Errorf("%s = %q, want %q", name, got, want)
		}
	}
}

func TestRunTargetDirectory(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	destDir := filepath.Join(dir, "out")
	if err := os.Mkdir(destDir, 0o755); err != nil {
		t.Fatal(err)
	}
	src := filepath.Join(dir, "f")
	writeFile(t, src, "data")

	if _, _, err := run(t, "-t", destDir, src); err != nil {
		t.Fatalf("Run -t error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(destDir, "f")); err != nil {
		t.Errorf("expected file copied into target dir: %v", err)
	}
}

func TestRunCreateDirectories(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	target := filepath.Join(dir, "a", "b", "c")

	if _, _, err := run(t, "-d", "-m", "700", target); err != nil {
		t.Fatalf("Run -d error = %v", err)
	}
	info, err := os.Stat(target)
	if err != nil {
		t.Fatal(err)
	}
	if !info.IsDir() {
		t.Errorf("%s is not a directory", target)
	}
	if info.Mode().Perm() != 0o700 {
		t.Errorf("dir mode = %o, want 700", info.Mode().Perm())
	}
}

func TestRunCreateLeading(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	writeFile(t, src, "z")
	dst := filepath.Join(dir, "nested", "deep", "dst")

	if _, errOut, err := run(t, "-D", src, dst); err != nil {
		t.Fatalf("Run -D error = %v (stderr=%q)", err, errOut)
	}
	if _, err := os.Stat(dst); err != nil {
		t.Errorf("expected file at %s: %v", dst, err)
	}
}

func TestRunPreserveTimestamps(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	dst := filepath.Join(dir, "dst")
	writeFile(t, src, "x")

	srcInfo, err := os.Stat(src)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := run(t, "-p", src, dst); err != nil {
		t.Fatalf("Run -p error = %v", err)
	}
	dstInfo, err := os.Stat(dst)
	if err != nil {
		t.Fatal(err)
	}
	if !dstInfo.ModTime().Equal(srcInfo.ModTime()) {
		t.Errorf("mtime = %v, want %v", dstInfo.ModTime(), srcInfo.ModTime())
	}
}

func TestRunVerbose(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	dst := filepath.Join(dir, "dst")
	writeFile(t, src, "x")

	out, _, err := run(t, "-v", src, dst)
	if err != nil {
		t.Fatalf("Run -v error = %v", err)
	}
	if !strings.Contains(out, "->") {
		t.Errorf("verbose output = %q, want it to contain '->'", out)
	}
}

func TestRunErrors(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		args []string
	}{
		{"no operand", nil},
		{"missing destination", []string{"only-source"}},
		{"invalid mode", []string{"-m", "999x", "a", "b"}},
		{"directory missing operand", []string{"-d"}},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, errOut, err := run(t, tt.args...)
			if err == nil {
				t.Errorf("expected error for args %v", tt.args)
			}
			if errOut == "" {
				t.Errorf("expected stderr message for args %v", tt.args)
			}
		})
	}
}

func TestRunMissingSource(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	_, errOut, err := run(t, filepath.Join(dir, "nope"), filepath.Join(dir, "dst"))
	if err == nil {
		t.Error("expected error for missing source")
	}
	if !strings.Contains(errOut, "install:") {
		t.Errorf("stderr = %q, want install: prefix", errOut)
	}
}

func TestRunOmitDirectory(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	srcDir := filepath.Join(dir, "adir")
	if err := os.Mkdir(srcDir, 0o755); err != nil {
		t.Fatal(err)
	}
	_, errOut, err := run(t, srcDir, filepath.Join(dir, "dst"))
	if err == nil {
		t.Error("expected error when source is a directory")
	}
	if !strings.Contains(errOut, "omitting directory") {
		t.Errorf("stderr = %q, want omitting directory", errOut)
	}
}

func TestRunHelp(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "--help")
	if err != nil {
		t.Fatalf("Run --help error = %v", err)
	}
	if !strings.Contains(out, "Usage: install") {
		t.Errorf("help = %q, want usage line", out)
	}
}
