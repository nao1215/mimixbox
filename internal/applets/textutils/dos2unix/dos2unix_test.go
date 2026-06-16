package dos2unix_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/textutils/dos2unix"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := dos2unix.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestRunConvertsCRLFToLFInPlace(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	file := filepath.Join(dir, "1.txt")
	if err := os.WriteFile(file, []byte("abc\r\ndef\r\nghi\r\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	out, errOut, err := run(t, file)
	if err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}
	if errOut != "" {
		t.Errorf("stderr = %q, want empty", errOut)
	}

	wantOut := "dos2unix: converting file " + file + " to Unix format...\n"
	if out != wantOut {
		t.Errorf("stdout = %q, want %q", out, wantOut)
	}

	got, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	want := "abc\ndef\nghi\n"
	if string(got) != want {
		t.Errorf("file content = %q, want %q", string(got), want)
	}
}

func TestRunMultipleFiles(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	f1 := filepath.Join(dir, "1.txt")
	f2 := filepath.Join(dir, "2.txt")
	for _, f := range []string{f1, f2} {
		if err := os.WriteFile(f, []byte("x\r\ny\r\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	out, _, err := run(t, f1, f2)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	wantOut := "dos2unix: converting file " + f1 + " to Unix format...\n" +
		"dos2unix: converting file " + f2 + " to Unix format...\n"
	if out != wantOut {
		t.Errorf("stdout = %q, want %q", out, wantOut)
	}
	for _, f := range []string{f1, f2} {
		got, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != "x\ny\n" {
			t.Errorf("file %s content = %q, want %q", f, string(got), "x\ny\n")
		}
	}
}

func TestRunDirectoryIsError(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	out, errOut, err := run(t, dir)
	if err == nil {
		t.Fatal("expected error for directory operand")
	}
	if out != "" {
		t.Errorf("stdout = %q, want empty", out)
	}
	wantErr := "dos2unix: skip " + dir + ": not regular file\n"
	if errOut != wantErr {
		t.Errorf("stderr = %q, want %q", errOut, wantErr)
	}
}

func TestRunMissingFileIsError(t *testing.T) {
	t.Parallel()
	missing := filepath.Join(t.TempDir(), "no_such_file.txt")

	out, errOut, err := run(t, missing)
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	if out != "" {
		t.Errorf("stdout = %q, want empty", out)
	}
	wantErr := "dos2unix: skip " + missing + ": not regular file\n"
	if errOut != wantErr {
		t.Errorf("stderr = %q, want %q", errOut, wantErr)
	}
}

func TestRunMixedFilesAndDirectory(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	f1 := filepath.Join(dir, "1.txt")
	f3 := filepath.Join(dir, "3.txt")
	subdir := filepath.Join(dir, "adir")
	if err := os.WriteFile(f1, []byte("a\r\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(f3, []byte("c\r\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(subdir, 0o700); err != nil {
		t.Fatal(err)
	}

	out, errOut, err := run(t, f1, subdir, f3)
	if err == nil {
		t.Fatal("expected error because one operand is a directory")
	}
	wantOut := "dos2unix: converting file " + f1 + " to Unix format...\n" +
		"dos2unix: converting file " + f3 + " to Unix format...\n"
	if out != wantOut {
		t.Errorf("stdout = %q, want %q", out, wantOut)
	}
	wantErr := "dos2unix: skip " + subdir + ": not regular file\n"
	if errOut != wantErr {
		t.Errorf("stderr = %q, want %q", errOut, wantErr)
	}

	// The valid files were still converted in place.
	got1, _ := os.ReadFile(f1)
	if string(got1) != "a\n" {
		t.Errorf("f1 content = %q, want %q", string(got1), "a\n")
	}
	got3, _ := os.ReadFile(f3)
	if string(got3) != "c\n" {
		t.Errorf("f3 content = %q, want %q", string(got3), "c\n")
	}
}

func TestRunPreservesFileMode(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	file := filepath.Join(dir, "m.txt")
	if err := os.WriteFile(file, []byte("a\r\nb\r\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := run(t, file); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	info, err := os.Stat(file)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("mode = %o, want 600 (in-place conversion must keep the mode)", info.Mode().Perm())
	}
	got, _ := os.ReadFile(file) //nolint:gosec // test-written file
	if string(got) != "a\nb\n" {
		t.Errorf("content = %q", got)
	}
}

func TestHelpSections(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "--help")
	if err != nil {
		t.Fatalf("--help error = %v", err)
	}
	for _, want := range []string{"Examples:", "Exit status:"} {
		if !strings.Contains(out, want) {
			t.Errorf("--help output missing %q:\n%s", want, out)
		}
	}
}
