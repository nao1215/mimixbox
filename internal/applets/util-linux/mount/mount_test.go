package mount

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func fixture(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	f := filepath.Join(dir, "mounts")
	content := "/dev/sda1 / ext4 rw,relatime 0 0\n" +
		"proc /proc proc rw,nosuid 0 0\n" +
		"tmpfs /my\\040mount tmpfs rw 0 0\n"
	if err := os.WriteFile(f, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	orig := mountsPath
	mountsPath = f
	t.Cleanup(func() { mountsPath = orig })
}

func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	err := New().Run(context.Background(), io, args)
	return out.String(), err
}

func TestList(t *testing.T) {
	fixture(t)
	out, err := run(t)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "/dev/sda1 on / type ext4 (rw,relatime)") {
		t.Errorf("ext4 line wrong:\n%s", out)
	}
	// The octal-escaped space must be decoded.
	if !strings.Contains(out, "tmpfs on /my mount type tmpfs (rw)") {
		t.Errorf("escape decoding wrong:\n%s", out)
	}
}

func TestTypeFilter(t *testing.T) {
	fixture(t)
	out, err := run(t, "-t", "ext4")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "ext4") || strings.Contains(out, "proc") || strings.Contains(out, "tmpfs") {
		t.Errorf("-t ext4 wrong:\n%s", out)
	}
}

func TestMountOperandUnsupported(t *testing.T) {
	fixture(t)
	if _, err := run(t, "/dev/sda1", "/mnt"); err == nil {
		t.Errorf("requesting a mount should fail deterministically")
	}
}

func TestMissingTable(t *testing.T) {
	orig := mountsPath
	mountsPath = "/no/such/mounts"
	defer func() { mountsPath = orig }()
	if _, err := run(t); err == nil {
		t.Errorf("a missing mount table should fail")
	}
}

func TestUnescape(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		`/my\040mount`:   "/my mount",
		`/tab\011here`:   "/tab\there",
		`/plain`:         "/plain",
		`/back\134slash`: `/back\slash`,  // \134 octal = backslash
		`/lone\bslash`:   `/lone\bslash`, // a non-octal escape is left as-is
	}
	for in, want := range cases {
		if got := unescape(in); got != want {
			t.Errorf("unescape(%q) = %q, want %q", in, got, want)
		}
	}
}
