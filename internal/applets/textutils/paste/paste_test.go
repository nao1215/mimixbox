package paste_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/textutils/paste"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, stdin string, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(stdin), Out: out, Err: errBuf}
	err := paste.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func writeFile(t *testing.T, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "f.txt")
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestSerial(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "a\nb\nc\n", "-s")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "a\tb\tc\n" {
		t.Errorf("out = %q", out)
	}
}

func TestSerialCustomDelimiter(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "a\nb\nc\n", "-s", "-d", ",")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "a,b,c\n" {
		t.Errorf("out = %q", out)
	}
}

func TestParallelTwoFiles(t *testing.T) {
	t.Parallel()
	f1 := writeFile(t, "1\n2\n3\n")
	f2 := writeFile(t, "a\nb\nc\n")
	out, _, err := run(t, "", f1, f2)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "1\ta\n2\tb\n3\tc\n" {
		t.Errorf("out = %q", out)
	}
}

func TestParallelUnevenLength(t *testing.T) {
	t.Parallel()
	f1 := writeFile(t, "1\n2\n")
	f2 := writeFile(t, "a\n")
	out, _, err := run(t, "", f1, f2)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "1\ta\n2\t\n" {
		t.Errorf("out = %q", out)
	}
}

func TestDelimiterEscape(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "a\nb\n", "-s", "-d", `\n`)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "a\nb\n" {
		t.Errorf("out = %q", out)
	}
}

func TestMissingFile(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, "", "/no/such/file")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(errOut, "paste: /no/such/file:") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := paste.New()
	if c.Name() != "paste" {
		t.Errorf("Name() = %q", c.Name())
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}
