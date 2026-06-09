package xzcomp

import (
	"bytes"
	"compress/gzip"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func capture(t *testing.T, c *Command, in []byte, args ...string) ([]byte, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: bytes.NewReader(in), Out: out, Err: errBuf}
	err := c.Run(context.Background(), io, args)
	return out.Bytes(), errBuf.String(), err
}

func TestRoundTrip(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name       string
		compress   func() *Command
		decompress func() *Command
	}{
		{"xz", NewXz, NewXzcat},
		{"lzma", NewLzma, NewLzcat},
	}
	payload := []byte("the quick brown fox\n")
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			comp, _, err := capture(t, tc.compress(), payload)
			if err != nil {
				t.Fatalf("compress error = %v", err)
			}
			if bytes.Equal(comp, payload) {
				t.Errorf("%s did not compress the data", tc.name)
			}
			got, _, err := capture(t, tc.decompress(), comp)
			if err != nil {
				t.Fatalf("decompress error = %v", err)
			}
			if !bytes.Equal(got, payload) {
				t.Errorf("round trip = %q, want %q", got, payload)
			}
		})
	}
}

func TestZcat(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	_, _ = zw.Write([]byte("gzip payload\n"))
	_ = zw.Close()

	got, _, err := capture(t, NewZcat(), buf.Bytes())
	if err != nil {
		t.Fatalf("zcat error = %v", err)
	}
	if string(got) != "gzip payload\n" {
		t.Errorf("zcat = %q", got)
	}
}

func TestInPlace(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "data.txt")
	if err := os.WriteFile(src, []byte("in place\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, _, err := capture(t, NewXz(), nil, src); err != nil {
		t.Fatalf("xz error = %v", err)
	}
	if _, err := os.Stat(src + ".xz"); err != nil {
		t.Errorf("expected %s.xz: %v", src, err)
	}
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Errorf("original should be removed without -k")
	}

	if _, _, err := capture(t, NewUnxz(), nil, src+".xz"); err != nil {
		t.Fatalf("unxz error = %v", err)
	}
	if got, _ := os.ReadFile(src); string(got) != "in place\n" {
		t.Errorf("decompressed file = %q", got)
	}
}

func TestSuffixError(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	f := filepath.Join(dir, "noext")
	if err := os.WriteFile(f, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := capture(t, NewUnxz(), nil, f); err == nil {
		t.Errorf("unxz on a file without .xz should fail")
	}
}

func TestHelp(t *testing.T) {
	t.Parallel()
	out, _, err := capture(t, NewXz(), nil, "--help")
	if err != nil {
		t.Fatalf("--help err = %v", err)
	}
	if !strings.Contains(string(out), "Usage: xz") {
		t.Errorf("--help = %q", out)
	}
}
