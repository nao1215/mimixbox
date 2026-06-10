package blkid

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

// image writes a file with magic placed at offset.
func image(t *testing.T, offset int64, magic []byte) string {
	t.Helper()
	dir := t.TempDir()
	f := filepath.Join(dir, "img")
	buf := make([]byte, offset+int64(len(magic))+16)
	copy(buf[offset:], magic)
	if err := os.WriteFile(f, buf, 0o644); err != nil {
		t.Fatal(err)
	}
	return f
}

func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	err := New().Run(context.Background(), io, args)
	return out.String(), err
}

func TestDetectTypes(t *testing.T) {
	t.Parallel()
	cases := []struct {
		typ    string
		offset int64
		magic  []byte
	}{
		{"ext2", 0x438, []byte{0x53, 0xEF}},
		{"xfs", 0, []byte("XFSB")},
		{"btrfs", 0x10040, []byte("_BHRfS_M")},
		{"ntfs", 3, []byte("NTFS    ")},
		{"squashfs", 0, []byte("hsqs")},
		{"swap", 0xFF6, []byte("SWAPSPACE2")},
	}
	for _, tc := range cases {
		t.Run(tc.typ, func(t *testing.T) {
			f := image(t, tc.offset, tc.magic)
			out, err := run(t, f)
			if err != nil {
				t.Fatalf("blkid error = %v", err)
			}
			want := f + ": TYPE=\"" + tc.typ + "\"\n"
			if out != want {
				t.Errorf("blkid = %q, want %q", out, want)
			}
		})
	}
}

func TestNothingIdentified(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	f := filepath.Join(dir, "blank")
	if err := os.WriteFile(f, make([]byte, 100000), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := run(t, f)
	if out != "" {
		t.Errorf("unexpected output: %q", out)
	}
	var ee *command.ExitError
	if e, ok := err.(*command.ExitError); ok {
		ee = e
	}
	if ee == nil || ee.Code != 2 {
		t.Errorf("err = %v, want exit 2", err)
	}
}

func TestMissingFile(t *testing.T) {
	t.Parallel()
	if _, err := run(t, "/no/such/blkid/file"); err == nil {
		t.Errorf("missing file should fail")
	}
}

func TestNoArgs(t *testing.T) {
	t.Parallel()
	if _, err := run(t); err == nil {
		t.Errorf("no file should fail")
	}
}
