package tune2fs

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

// craftImage builds a minimal valid ext2 superblock image.
func craftImage(t *testing.T, withMagic bool) string {
	t.Helper()
	buf := make([]byte, superblockOffset+1024)
	sb := buf[superblockOffset:]
	binary.LittleEndian.PutUint32(sb[0:], 100)  // inodes
	binary.LittleEndian.PutUint32(sb[4:], 200)  // blocks
	binary.LittleEndian.PutUint32(sb[24:], 2)   // log_block_size -> 1024<<2 = 4096
	copy(sb[120:136], "testfs")                 // volume name
	if withMagic {
		binary.LittleEndian.PutUint16(sb[56:], extMagic)
	}
	p := filepath.Join(t.TempDir(), "fs.img")
	if err := os.WriteFile(p, buf, 0o644); err != nil {
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

func TestList(t *testing.T) {
	img := craftImage(t, true)
	out, err := run(t, "-l", img)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"Filesystem volume name:   testfs",
		"Inode count:              100",
		"Block count:              200",
		"Block size:               4096",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
}

func TestNotExt(t *testing.T) {
	img := craftImage(t, false) // no magic
	if _, err := run(t, "-l", img); err == nil {
		t.Errorf("a non-ext image should fail")
	}
}

func TestRequiresList(t *testing.T) {
	img := craftImage(t, true)
	if _, err := run(t, img); err == nil {
		t.Errorf("without -l it should fail (no mutation supported)")
	}
}

func TestErrors(t *testing.T) {
	if _, err := run(t, "-l"); err == nil {
		t.Errorf("missing file should fail")
	}
	if _, err := run(t, "-l", "/no/such/image"); err == nil {
		t.Errorf("a missing image should fail")
	}
}
