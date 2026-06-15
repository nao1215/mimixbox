package makemime

import (
	"bytes"
	"context"
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func write(t *testing.T, name, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestSingleToStdout(t *testing.T) {
	in := write(t, "input.txt", "hello world")
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, []string{"-c", "text/plain", in}); err != nil {
		t.Fatal(err)
	}
	s := out.String()
	if !strings.Contains(s, "MIME-Version: 1.0") {
		t.Errorf("missing MIME-Version header:\n%s", s)
	}
	if !strings.Contains(s, "Content-Type: text/plain") {
		t.Errorf("missing Content-Type:\n%s", s)
	}
	wantB64 := base64.StdEncoding.EncodeToString([]byte("hello world"))
	if !strings.Contains(s, wantB64) {
		t.Errorf("body not base64-encoded; want %q in:\n%s", wantB64, s)
	}
}

func TestOutputFile(t *testing.T) {
	in := write(t, "input.txt", "data")
	outPath := filepath.Join(t.TempDir(), "out.eml")
	io := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, []string{"-o", outPath, in}); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), "MIME-Version") {
		t.Errorf("output file missing MIME content:\n%s", got)
	}
}

func TestMultipart(t *testing.T) {
	a := write(t, "a.txt", "alpha")
	b := write(t, "b.txt", "beta")
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, []string{a, b}); err != nil {
		t.Fatal(err)
	}
	s := out.String()
	if !strings.Contains(s, "multipart/mixed") {
		t.Errorf("expected multipart/mixed:\n%s", s)
	}
	if !strings.Contains(s, "boundary=") {
		t.Errorf("expected boundary parameter:\n%s", s)
	}
}

func TestNoArgs(t *testing.T) {
	io := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, nil); err == nil {
		t.Fatal("expected error with no input file")
	}
}
