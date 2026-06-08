package xxd_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/textutils/xxd"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, stdin string, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(stdin), Out: out, Err: errBuf}
	err := xxd.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestDumpShort(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "hello\n")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	want := "00000000: 6865 6c6c 6f0a                           hello.\n"
	if out != want {
		t.Errorf("out = %q, want %q", out, want)
	}
}

func TestDumpFullLine(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "0123456789abcdefghij")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	want := "00000000: 3031 3233 3435 3637 3839 6162 6364 6566  0123456789abcdef\n" +
		"00000010: 6768 696a                                ghij\n"
	if out != want {
		t.Errorf("out = %q, want %q", out, want)
	}
}

func TestRevert(t *testing.T) {
	t.Parallel()
	dump, _, err := run(t, "hello\n")
	if err != nil {
		t.Fatalf("dump error = %v", err)
	}
	out, _, err := run(t, dump, "-r")
	if err != nil {
		t.Fatalf("revert error = %v", err)
	}
	if out != "hello\n" {
		t.Errorf("out = %q", out)
	}
}

func TestRevertFullLineRoundTrip(t *testing.T) {
	t.Parallel()
	orig := "0123456789abcdefghij"
	dump, _, err := run(t, orig)
	if err != nil {
		t.Fatalf("dump error = %v", err)
	}
	out, _, err := run(t, dump, "-r")
	if err != nil {
		t.Fatalf("revert error = %v", err)
	}
	if out != orig {
		t.Errorf("round trip = %q, want %q", out, orig)
	}
}

func TestRevertInvalid(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, "00000000: zzzz  ..\n", "-r")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(errOut, "xxd: invalid hex dump") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestMissingFile(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, "", "/no/such/file")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(errOut, "xxd:") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := xxd.New()
	if c.Name() != "xxd" {
		t.Errorf("Name() = %q", c.Name())
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}
