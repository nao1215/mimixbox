package mkswap

import (
	"bytes"
	"context"
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func makeImage(t *testing.T, pages int) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "swap.img")
	if err := os.WriteFile(p, make([]byte, pages*pageSize), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	err := New().Run(context.Background(), io, args)
	return out.String(), err
}

func TestWritesSwap(t *testing.T) {
	img := makeImage(t, 16)
	out, err := run(t, img)
	if err != nil {
		t.Fatal(err)
	}
	usable := 15 * pageSize
	if !strings.Contains(out, "version 1") {
		t.Errorf("missing version banner: %q", out)
	}
	data, _ := os.ReadFile(img)
	// Signature at pageSize-10.
	if got := string(data[pageSize-10 : pageSize]); got != signature {
		t.Errorf("signature = %q, want %q", got, signature)
	}
	// Header: version 1, last_page = 15.
	if v := binary.LittleEndian.Uint32(data[headerOffset:]); v != 1 {
		t.Errorf("version = %d, want 1", v)
	}
	if lp := binary.LittleEndian.Uint32(data[headerOffset+4:]); lp != 15 {
		t.Errorf("last_page = %d, want 15", lp)
	}
	if !strings.Contains(out, "bytes)") || usable <= 0 {
		t.Errorf("size banner wrong: %q", out)
	}
}

func TestWritesLabel(t *testing.T) {
	img := makeImage(t, 16)
	if _, err := run(t, "-L", "myswap", img); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(img)
	label := string(bytes.TrimRight(data[labelOffset:labelOffset+16], "\x00"))
	if label != "myswap" {
		t.Errorf("label = %q, want myswap", label)
	}
}

func TestTooSmall(t *testing.T) {
	img := makeImage(t, 1) // only one page
	if _, err := run(t, img); err == nil {
		t.Errorf("a one-page file should be rejected")
	}
}

func TestErrors(t *testing.T) {
	if _, err := run(t); err == nil {
		t.Errorf("missing file should fail")
	}
	if _, err := run(t, "/no/such/swapfile"); err == nil {
		t.Errorf("a missing file should fail")
	}
}
