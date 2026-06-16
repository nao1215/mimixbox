package rpm2cpio_test

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/archival/rpm2cpio"
	"github.com/nao1215/mimixbox/internal/command"
)

// buildRPM assembles a minimal RPM with a gzip payload. (The signature and main
// headers are empty index headers, which is all rpm2cpio needs.)
func buildRPM(payload []byte) []byte {
	lead := make([]byte, 96)
	lead[0], lead[1], lead[2], lead[3] = 0xed, 0xab, 0xee, 0xdb
	header := func() []byte {
		intro := make([]byte, 16)
		intro[0], intro[1], intro[2], intro[3] = 0x8e, 0xad, 0xe8, 0x01
		binary.BigEndian.PutUint32(intro[8:], 0)
		binary.BigEndian.PutUint32(intro[12:], 0)
		return intro
	}
	out := append([]byte{}, lead...)
	out = append(out, header()...) // signature header (16 bytes, 8-aligned)
	out = append(out, header()...) // main header
	return append(out, payload...)
}

func gz(s string) []byte {
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	_, _ = w.Write([]byte(s))
	_ = w.Close()
	return b.Bytes()
}

func run(t *testing.T, stdin []byte, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: bytes.NewReader(stdin), Out: out, Err: errBuf}
	err := rpm2cpio.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := rpm2cpio.New()
	if got := c.Name(); got != "rpm2cpio" {
		t.Errorf("Name() = %q", got)
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis empty")
	}
}

func TestExtractFromStdin(t *testing.T) {
	t.Parallel()
	out, errOut, err := run(t, buildRPM(gz("CPIO-DATA-HERE")))
	if err != nil {
		t.Fatalf("err = %v (stderr=%q)", err, errOut)
	}
	if out != "CPIO-DATA-HERE" {
		t.Errorf("out = %q", out)
	}
}

func TestExtractFromFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	p := filepath.Join(dir, "pkg.rpm")
	if err := os.WriteFile(p, buildRPM(gz("PAYLOAD")), 0o644); err != nil {
		t.Fatal(err)
	}
	out, _, err := run(t, nil, p)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if out != "PAYLOAD" {
		t.Errorf("out = %q", out)
	}
}

func TestNotAnRPM(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, make([]byte, 200))
	if err == nil {
		t.Error("expected error for non-RPM input")
	}
	if !strings.Contains(errOut, "rpm2cpio:") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestMissingFile(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, nil, filepath.Join(t.TempDir(), "nope.rpm"))
	if err == nil {
		t.Error("expected error for missing file")
	}
	if !strings.Contains(errOut, "rpm2cpio:") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestHelp(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, nil, "--help")
	if err != nil {
		t.Fatalf("help err = %v", err)
	}
	for _, want := range []string{"Usage: rpm2cpio", "Examples:", "Exit status:"} {
		if !strings.Contains(out, want) {
			t.Errorf("help missing %q\n%s", want, out)
		}
	}
}
