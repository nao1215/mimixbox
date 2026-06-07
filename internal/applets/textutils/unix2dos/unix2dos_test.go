package unix2dos_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/textutils/unix2dos"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := unix2dos.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestRunConvertsLFToCRLF(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	file := filepath.Join(dir, "1.txt")
	if err := os.WriteFile(file, []byte("abc\ndef\nghi\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	out, errOut, err := run(t, file)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if errOut != "" {
		t.Errorf("stderr = %q, want empty", errOut)
	}
	want := "unix2dos: converting file " + file + " to DOS format...\n"
	if out != want {
		t.Errorf("stdout = %q, want %q", out, want)
	}

	got, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "abc\r\ndef\r\nghi\r\n" {
		t.Errorf("file = %q, want CRLF line endings", string(got))
	}
}

func TestRunCRLFIsIdempotent(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	file := filepath.Join(dir, "1.txt")
	if err := os.WriteFile(file, []byte("abc\r\ndef\r\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, _, err := run(t, file); err != nil {
		t.Fatalf("Run error = %v", err)
	}

	got, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "abc\r\ndef\r\n" {
		t.Errorf("file = %q, want CRLF preserved (no doubled CR)", string(got))
	}
}

func TestRunDirectory(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	out, errOut, err := run(t, dir)
	if err == nil {
		t.Fatal("expected error for directory operand")
	}
	if out != "" {
		t.Errorf("stdout = %q, want empty", out)
	}
	want := "unix2dos: skip " + dir + ": not regular file\n"
	if errOut != want {
		t.Errorf("stderr = %q, want %q", errOut, want)
	}
}

func TestRunMissingFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	missing := filepath.Join(dir, "nope.txt")

	out, errOut, err := run(t, missing)
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	if out != "" {
		t.Errorf("stdout = %q, want empty", out)
	}
	want := "unix2dos: skip " + missing + ": not regular file\n"
	if errOut != want {
		t.Errorf("stderr = %q, want %q", errOut, want)
	}
}

func TestRunFileAndDirectory(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	file := filepath.Join(dir, "1.txt")
	if err := os.WriteFile(file, []byte("abc\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	subdir := filepath.Join(dir, "sub")
	if err := os.Mkdir(subdir, 0o700); err != nil {
		t.Fatal(err)
	}

	out, errOut, err := run(t, file, subdir)
	if err == nil {
		t.Fatal("expected error because one operand is a directory")
	}
	if out != "unix2dos: converting file "+file+" to DOS format...\n" {
		t.Errorf("stdout = %q", out)
	}
	if errOut != "unix2dos: skip "+subdir+": not regular file\n" {
		t.Errorf("stderr = %q", errOut)
	}

	got, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "abc\r\n" {
		t.Errorf("file = %q, want CRLF", string(got))
	}
}
