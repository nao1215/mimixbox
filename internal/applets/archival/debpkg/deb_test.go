package debpkg

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

// tarEntry describes one file written into a test tarball.
type tarEntry struct {
	name     string
	mode     int64
	body     string
	typeflag byte
	linkname string
}

// buildTarGz builds a gzip-compressed tar archive from entries.
func buildTarGz(t *testing.T, entries []tarEntry) []byte {
	t.Helper()
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)
	for _, e := range entries {
		typ := e.typeflag
		if typ == 0 {
			typ = tar.TypeReg
		}
		hdr := &tar.Header{
			Name:     e.name,
			Mode:     e.mode,
			Size:     int64(len(e.body)),
			Typeflag: typ,
			Linkname: e.linkname,
		}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatal(err)
		}
		if len(e.body) > 0 {
			if _, err := tw.Write([]byte(e.body)); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// writeArMember appends one ar member to buf.
func writeArMember(buf *bytes.Buffer, name string, data []byte) {
	hdr := fmt.Sprintf("%-16s%-12d%-6d%-6d%-8o%-10d`\n", name, 0, 0, 0, 0o644, len(data))
	buf.WriteString(hdr)
	buf.Write(data)
	if len(data)%2 == 1 {
		buf.WriteByte('\n')
	}
}

// buildDeb assembles a .deb (ar archive) from a control tarball and data
// tarball.
func buildDeb(t *testing.T, control, data []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	buf.WriteString(arMagic)
	writeArMember(&buf, "debian-binary", []byte("2.0\n"))
	writeArMember(&buf, "control.tar.gz", control)
	writeArMember(&buf, "data.tar.gz", data)
	return buf.Bytes()
}

// sampleDeb returns a representative .deb fixture as bytes.
func sampleDeb(t *testing.T) []byte {
	t.Helper()
	control := buildTarGz(t, []tarEntry{
		{name: "./control", mode: 0o644, body: "Package: hello\nVersion: 1.0\nArchitecture: all\nDescription: test package\n"},
		{name: "./md5sums", mode: 0o644, body: "abc  usr/bin/hello\n"},
	})
	data := buildTarGz(t, []tarEntry{
		{name: "./usr/", mode: 0o755, typeflag: tar.TypeDir},
		{name: "./usr/bin/", mode: 0o755, typeflag: tar.TypeDir},
		{name: "./usr/bin/hello", mode: 0o755, body: "#!/bin/sh\necho hello\n"},
		{name: "./usr/share/doc/hello/README", mode: 0o644, body: "readme contents\n"},
	})
	return buildDeb(t, control, data)
}

func TestReadParsesPackage(t *testing.T) {
	pkg, err := Read(bytes.NewReader(sampleDeb(t)))
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if strings.TrimSpace(pkg.Version) != "2.0" {
		t.Fatalf("version = %q", pkg.Version)
	}
	control, err := pkg.ControlFile("control")
	if err != nil {
		t.Fatalf("ControlFile: %v", err)
	}
	if !strings.Contains(string(control), "Package: hello") {
		t.Fatalf("control missing Package field: %q", control)
	}
}

func TestReadRejectsNonDeb(t *testing.T) {
	if _, err := Read(bytes.NewReader([]byte("not an ar archive"))); err == nil {
		t.Fatal("expected error for non-deb input")
	}
}

func TestDataEntriesAndExtract(t *testing.T) {
	pkg, err := Read(bytes.NewReader(sampleDeb(t)))
	if err != nil {
		t.Fatal(err)
	}
	entries, err := pkg.DataEntries()
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name)
	}
	if !contains(names, "./usr/bin/hello") {
		t.Fatalf("expected ./usr/bin/hello, got %v", names)
	}

	dest := t.TempDir()
	if err := pkg.Extract(dest); err != nil {
		t.Fatalf("Extract: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(dest, "usr", "bin", "hello"))
	if err != nil {
		t.Fatalf("extracted file missing: %v", err)
	}
	if !strings.Contains(string(got), "echo hello") {
		t.Fatalf("extracted content wrong: %q", got)
	}
}

func TestExtractRejectsTraversal(t *testing.T) {
	control := buildTarGz(t, []tarEntry{{name: "./control", mode: 0o644, body: "Package: evil\n"}})
	data := buildTarGz(t, []tarEntry{
		{name: "../../../../tmp/evil", mode: 0o644, body: "pwned"},
	})
	deb := buildDeb(t, control, data)

	pkg, err := Read(bytes.NewReader(deb))
	if err != nil {
		t.Fatal(err)
	}
	dest := t.TempDir()
	if err := pkg.Extract(dest); err == nil {
		t.Fatal("expected traversal rejection")
	}
	// Ensure nothing escaped.
	if _, err := os.Stat(filepath.Join(filepath.Dir(dest), "..", "tmp", "evil")); err == nil {
		t.Fatal("file escaped destination")
	}
}

func TestExtractRejectsSymlinkTraversal(t *testing.T) {
	control := buildTarGz(t, []tarEntry{{name: "./control", mode: 0o644, body: "Package: evil\n"}})
	data := buildTarGz(t, []tarEntry{
		{name: "./link", mode: 0o777, typeflag: tar.TypeSymlink, linkname: "../../../../etc/passwd"},
	})
	deb := buildDeb(t, control, data)
	pkg, err := Read(bytes.NewReader(deb))
	if err != nil {
		t.Fatal(err)
	}
	if err := pkg.Extract(t.TempDir()); err == nil {
		t.Fatal("expected symlink traversal rejection")
	}
}

func TestDpkgDebContents(t *testing.T) {
	var out, errb bytes.Buffer
	dir := writeDebFile(t)
	c := NewDpkgDeb()
	if err := c.Run(context.Background(), command.IO{Out: &out, Err: &errb}, []string{"-c", dir}); err != nil {
		t.Fatalf("run: %v err=%s", err, errb.String())
	}
	if !strings.Contains(out.String(), "usr/bin/hello") {
		t.Fatalf("contents missing hello: %q", out.String())
	}
}

func TestDpkgDebInfoAndField(t *testing.T) {
	deb := writeDebFile(t)

	var info bytes.Buffer
	if err := NewDpkgDeb().Run(context.Background(), command.IO{Out: &info, Err: &bytes.Buffer{}}, []string{"-I", deb}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(info.String(), "Package: hello") {
		t.Fatalf("-I missing control: %q", info.String())
	}

	var field bytes.Buffer
	if err := NewDpkgDeb().Run(context.Background(), command.IO{Out: &field, Err: &bytes.Buffer{}}, []string{"-f", deb, "Package", "Version"}); err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(field.String()); got != "hello\n1.0" {
		t.Fatalf("-f output = %q", got)
	}
}

func TestDpkgDebExtract(t *testing.T) {
	deb := writeDebFile(t)
	dest := t.TempDir()
	if err := NewDpkgDeb().Run(context.Background(), command.IO{Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}, []string{"-x", deb, dest}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dest, "usr", "bin", "hello")); err != nil {
		t.Fatalf("extract failed: %v", err)
	}
}

func TestDpkgUnsupportedDatabaseOps(t *testing.T) {
	for _, flag := range []string{"-i", "-r", "-l", "--configure"} {
		var errb bytes.Buffer
		err := NewDpkg().Run(context.Background(), command.IO{Out: &bytes.Buffer{}, Err: &errb}, []string{flag, "whatever"})
		if err == nil {
			t.Fatalf("%s should fail", flag)
		}
		if !strings.Contains(errb.String(), "not supported") {
			t.Fatalf("%s error message = %q", flag, errb.String())
		}
	}
}

func TestDpkgExtractAndContents(t *testing.T) {
	deb := writeDebFile(t)

	var out bytes.Buffer
	if err := NewDpkg().Run(context.Background(), command.IO{Out: &out, Err: &bytes.Buffer{}}, []string{"-c", deb}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "hello") {
		t.Fatalf("dpkg -c missing hello: %q", out.String())
	}

	dest := t.TempDir()
	if err := NewDpkg().Run(context.Background(), command.IO{Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}, []string{"-x", deb, dest}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dest, "usr", "bin", "hello")); err != nil {
		t.Fatalf("dpkg -x failed: %v", err)
	}
}

// writeDebFile writes the sample .deb to a temp file and returns its path.
func writeDebFile(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "hello.deb")
	if err := os.WriteFile(path, sampleDeb(t), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func contains(s []string, want string) bool {
	for _, v := range s {
		if v == want {
			return true
		}
	}
	return false
}
