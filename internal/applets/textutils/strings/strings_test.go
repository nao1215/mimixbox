package strings_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	cmdstrings "github.com/nao1215/mimixbox/internal/applets/textutils/strings"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, stdin string, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(stdin), Out: out, Err: errBuf}
	err := cmdstrings.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestRun(t *testing.T) {
	t.Parallel()
	// "hi" (too short) then NUL then "hello" then NUL then "world".
	in := "hi\x00hello\x00world"
	out, _, err := run(t, in)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "hello\nworld\n" {
		t.Errorf("out = %q", out)
	}
}

func TestMinLength(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "ab\x00abcd", "-n", "2")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "ab\nabcd\n" {
		t.Errorf("out = %q", out)
	}
}

func TestTrailingRun(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "\x01\x02abcd")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "abcd\n" {
		t.Errorf("out = %q", out)
	}
}

func TestInvalidLength(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, "abcd", "-n", "0")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(errOut, "invalid minimum string length") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestMissingFile(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, "", "/no/such/file")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(errOut, "strings: /no/such/file:") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := cmdstrings.New()
	if c.Name() != "strings" {
		t.Errorf("Name() = %q", c.Name())
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}
