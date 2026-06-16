package tar

import (
	"archive/tar"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func testIO() (command.IO, *bytes.Buffer) {
	errBuf := &bytes.Buffer{}
	return command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: errBuf}, errBuf
}

// TestSafeJoin verifies path-traversal entries are rejected while normal names
// resolve inside the destination.
func TestSafeJoin(t *testing.T) {
	t.Parallel()
	dest := "/tmp/dest"
	tests := []struct {
		name    string
		entry   string
		wantErr bool
	}{
		{"plain", "file.txt", false},
		{"nested", "sub/file.txt", false},
		{"dot self", ".", false},
		{"escape parent", "../evil", true},
		{"deep escape", "a/../../evil", true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := safeJoin(dest, tt.entry)
			if (err != nil) != tt.wantErr {
				t.Fatalf("safeJoin(%q, %q) err = %v, wantErr %v", dest, tt.entry, err, tt.wantErr)
			}
			if !tt.wantErr && !strings.HasPrefix(got, filepath.Clean(dest)) {
				t.Errorf("safeJoin = %q, want under %q", got, dest)
			}
		})
	}
}

func TestHasPrefix(t *testing.T) {
	t.Parallel()
	if !hasPrefix("abcdef", "abc") {
		t.Error("hasPrefix(abcdef, abc) = false")
	}
	if hasPrefix("ab", "abc") {
		t.Error("hasPrefix(ab, abc) = true")
	}
	if hasPrefix("xyz", "abc") {
		t.Error("hasPrefix(xyz, abc) = true")
	}
}

// TestExtractEntryDir covers the directory entry branch.
func TestExtractEntryDir(t *testing.T) {
	t.Parallel()
	dest := t.TempDir()
	io, _ := testIO()
	hdr := &tar.Header{Name: "newdir/", Typeflag: tar.TypeDir, Mode: 0o755}
	if err := extractEntry(io, nil, dest, hdr, false); err != nil {
		t.Fatalf("extractEntry dir err = %v", err)
	}
	info, err := os.Stat(filepath.Join(dest, "newdir"))
	if err != nil || !info.IsDir() {
		t.Fatalf("directory not created: %v", err)
	}
}

// TestExtractEntryRegularVerbose covers the regular-file branch with verbose
// output enabled (so the banner branch is taken too).
func TestExtractEntryRegularVerbose(t *testing.T) {
	t.Parallel()
	dest := t.TempDir()
	io, errBuf := testIO()
	body := "hello body"
	hdr := &tar.Header{Name: "sub/file.txt", Typeflag: tar.TypeReg, Mode: 0o644, Size: int64(len(body))}
	tr := tar.NewReader(buildArchive(t, hdr, body))
	if _, err := tr.Next(); err != nil {
		t.Fatal(err)
	}
	if err := extractEntry(io, tr, dest, hdr, true); err != nil {
		t.Fatalf("extractEntry reg err = %v", err)
	}
	got, err := os.ReadFile(filepath.Join(dest, "sub", "file.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != body {
		t.Errorf("extracted = %q, want %q", got, body)
	}
	if !strings.Contains(errBuf.String(), "sub/file.txt") {
		t.Errorf("verbose stderr = %q, want entry name", errBuf.String())
	}
}

// TestExtractEntrySymlink covers the symlink branch.
func TestExtractEntrySymlink(t *testing.T) {
	t.Parallel()
	dest := t.TempDir()
	io, _ := testIO()
	hdr := &tar.Header{Name: "link", Typeflag: tar.TypeSymlink, Linkname: "target", Mode: 0o777}
	if err := extractEntry(io, nil, dest, hdr, false); err != nil {
		t.Fatalf("extractEntry symlink err = %v", err)
	}
	link, err := os.Readlink(filepath.Join(dest, "link"))
	if err != nil {
		t.Fatalf("readlink err = %v", err)
	}
	if link != "target" {
		t.Errorf("symlink target = %q, want target", link)
	}
}

// TestExtractEntryUnsupportedSkipped covers the default (skip) branch.
func TestExtractEntryUnsupportedSkipped(t *testing.T) {
	t.Parallel()
	dest := t.TempDir()
	io, _ := testIO()
	hdr := &tar.Header{Name: "fifo", Typeflag: tar.TypeFifo, Mode: 0o644}
	if err := extractEntry(io, nil, dest, hdr, false); err != nil {
		t.Fatalf("extractEntry fifo err = %v, want nil (skipped)", err)
	}
	if _, err := os.Lstat(filepath.Join(dest, "fifo")); !os.IsNotExist(err) {
		t.Errorf("fifo should have been skipped, stat err = %v", err)
	}
}

// TestExtractEntryRejectsTraversal verifies a malicious entry name is refused.
func TestExtractEntryRejectsTraversal(t *testing.T) {
	t.Parallel()
	dest := t.TempDir()
	io, _ := testIO()
	hdr := &tar.Header{Name: "../escape.txt", Typeflag: tar.TypeReg, Mode: 0o644}
	if err := extractEntry(io, nil, dest, hdr, false); err == nil {
		t.Fatal("expected extractEntry to reject a traversal entry")
	}
}

// buildArchive returns a reader over a one-entry tar archive carrying body.
func buildArchive(t *testing.T, hdr *tar.Header, body string) *bytes.Reader {
	t.Helper()
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write([]byte(body)); err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	return bytes.NewReader(buf.Bytes())
}
