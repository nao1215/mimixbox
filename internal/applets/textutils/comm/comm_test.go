package comm_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/textutils/comm"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := comm.New().Run(context.Background(), io, args)
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

func TestThreeColumns(t *testing.T) {
	t.Parallel()
	a := writeFile(t, "apple\nbanana\ncherry\n")
	b := writeFile(t, "banana\ncherry\ndate\n")
	out, _, err := run(t, a, b)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	want := "apple\n\t\tbanana\n\t\tcherry\n\tdate\n"
	if out != want {
		t.Errorf("out = %q, want %q", out, want)
	}
}

func TestSuppressColumns(t *testing.T) {
	t.Parallel()
	a := writeFile(t, "apple\nbanana\ncherry\n")
	b := writeFile(t, "banana\ncherry\ndate\n")
	// -1 -2 leaves only the common column, with no indentation.
	out, _, err := run(t, "-1", "-2", a, b)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "banana\ncherry\n" {
		t.Errorf("out = %q", out)
	}
}

func TestOnlyUniqueToFirst(t *testing.T) {
	t.Parallel()
	a := writeFile(t, "apple\nbanana\n")
	b := writeFile(t, "banana\n")
	out, _, err := run(t, "-2", "-3", a, b)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "apple\n" {
		t.Errorf("out = %q", out)
	}
}

func TestWrongOperandCount(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, "onlyone")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(errOut, "two file operands") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestMissingFile(t *testing.T) {
	t.Parallel()
	b := writeFile(t, "x\n")
	_, errOut, err := run(t, "/no/such/file", b)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(errOut, "comm: /no/such/file:") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := comm.New()
	if c.Name() != "comm" {
		t.Errorf("Name() = %q", c.Name())
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}
