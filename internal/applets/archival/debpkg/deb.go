// Package debpkg provides a small, read-only reader for Debian binary package
// (.deb) files plus the dpkg-deb and dpkg applets built on top of it. A .deb is
// an ar archive containing three members: "debian-binary" (a version string),
// "control.tar[.gz|.xz|.zst]" (package metadata) and "data.tar[.gz|.xz|...]"
// (the installed files). This package parses that container and the inner tar
// archives so the applets can list, extract and inspect a package without ever
// touching the host package database, and it does so path-safely so a hostile
// archive cannot escape the requested destination directory.
package debpkg

import (
	"archive/tar"
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ulikunitz/xz"
	"github.com/ulikunitz/xz/lzma"
)

// arMagic is the ar archive global header.
const arMagic = "!<arch>\n"

// arHeaderSize is the fixed size of each ar member header.
const arHeaderSize = 60

// ErrNotDeb reports a file that is not a Debian package.
var ErrNotDeb = errors.New("not a Debian binary package (.deb)")

// ErrUnsupportedCompression reports a control/data tarball compressed with a
// codec this reader does not bundle (currently only zstd).
var ErrUnsupportedCompression = errors.New("unsupported tarball compression")

// arMember is one raw member of the outer ar archive.
type arMember struct {
	name string
	data []byte
}

// Package is a parsed .deb file: its format version plus the raw control and
// data tarballs (still compressed) so callers decode only what they need.
type Package struct {
	Version     string // contents of the debian-binary member, e.g. "2.0\n"
	controlName string
	controlData []byte
	dataName    string
	dataData    []byte
}

// Open reads and parses the .deb file at path.
func Open(path string) (*Package, error) {
	f, err := os.Open(path) //nolint:gosec // user-named file
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	return Read(f)
}

// Read parses a .deb from r.
func Read(r io.Reader) (*Package, error) {
	members, err := readAr(r)
	if err != nil {
		return nil, err
	}

	pkg := &Package{}
	for _, m := range members {
		switch {
		case m.name == "debian-binary":
			pkg.Version = string(m.data)
		case strings.HasPrefix(m.name, "control.tar"):
			pkg.controlName = m.name
			pkg.controlData = m.data
		case strings.HasPrefix(m.name, "data.tar"):
			pkg.dataName = m.name
			pkg.dataData = m.data
		}
	}
	if pkg.Version == "" || pkg.dataData == nil {
		return nil, ErrNotDeb
	}
	return pkg, nil
}

// readAr parses every member of an ar archive into memory.
func readAr(r io.Reader) ([]arMember, error) {
	magic := make([]byte, len(arMagic))
	if _, err := io.ReadFull(r, magic); err != nil || string(magic) != arMagic {
		return nil, ErrNotDeb
	}

	var members []arMember
	for {
		h := make([]byte, arHeaderSize)
		_, err := io.ReadFull(r, h)
		if err == io.EOF {
			return members, nil
		}
		if err != nil {
			return nil, ErrNotDeb
		}
		name := strings.TrimRight(string(h[0:16]), " ")
		name = strings.TrimSuffix(name, "/")
		size, err := strconv.ParseInt(strings.TrimSpace(string(h[48:58])), 10, 64)
		if err != nil || size < 0 {
			return nil, fmt.Errorf("%w: corrupt member header", ErrNotDeb)
		}
		data := make([]byte, size)
		if _, err := io.ReadFull(r, data); err != nil {
			return nil, fmt.Errorf("%w: truncated member %q", ErrNotDeb, name)
		}
		if size%2 == 1 { // members are padded to an even boundary
			var pad [1]byte
			if _, err := io.ReadFull(r, pad[:]); err != nil && err != io.EOF {
				return nil, err
			}
		}
		members = append(members, arMember{name: name, data: data})
	}
}

// tarReader decompresses a control/data tarball (named like "data.tar.gz")
// into an *tar.Reader, choosing the codec from the member suffix.
func tarReader(name string, data []byte) (*tar.Reader, error) {
	r := bytes.NewReader(data)
	switch {
	case strings.HasSuffix(name, ".tar"):
		return tar.NewReader(r), nil
	case strings.HasSuffix(name, ".gz"):
		zr, err := gzip.NewReader(r)
		if err != nil {
			return nil, err
		}
		return tar.NewReader(zr), nil
	case strings.HasSuffix(name, ".xz"):
		zr, err := xz.NewReader(r)
		if err != nil {
			return nil, err
		}
		return tar.NewReader(zr), nil
	case strings.HasSuffix(name, ".lzma"):
		zr, err := lzma.NewReader(r)
		if err != nil {
			return nil, err
		}
		return tar.NewReader(zr), nil
	case strings.HasSuffix(name, ".bz2"):
		return tar.NewReader(bzip2.NewReader(r)), nil
	case strings.HasSuffix(name, ".zst"):
		return nil, fmt.Errorf("%w: zstd (%s)", ErrUnsupportedCompression, name)
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedCompression, name)
	}
}

// Entry describes one file in the data tarball.
type Entry struct {
	Name string
	Mode int64
	UID  int
	GID  int
	Size int64
	Type byte
}

// DataEntries returns the entries of the data tarball.
func (p *Package) DataEntries() ([]Entry, error) {
	tr, err := tarReader(p.dataName, p.dataData)
	if err != nil {
		return nil, err
	}
	var entries []Entry
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		entries = append(entries, Entry{
			Name: hdr.Name,
			Mode: hdr.Mode,
			UID:  hdr.Uid,
			GID:  hdr.Gid,
			Size: hdr.Size,
			Type: hdr.Typeflag,
		})
	}
	return entries, nil
}

// ControlFile returns the contents of a single file (e.g. "control") from the
// control tarball, or ErrNotFound if it is absent.
func (p *Package) ControlFile(name string) ([]byte, error) {
	if p.controlData == nil {
		return nil, ErrNotFound
	}
	tr, err := tarReader(p.controlName, p.controlData)
	if err != nil {
		return nil, err
	}
	want := strings.TrimPrefix(name, "./")
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		got := strings.TrimPrefix(hdr.Name, "./")
		if got == want {
			return io.ReadAll(tr) //nolint:gosec // bounded by archive size
		}
	}
	return nil, ErrNotFound
}

// ControlNames lists the file names inside the control tarball.
func (p *Package) ControlNames() ([]string, error) {
	if p.controlData == nil {
		return nil, nil
	}
	tr, err := tarReader(p.controlName, p.controlData)
	if err != nil {
		return nil, err
	}
	var names []string
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		name := strings.TrimPrefix(hdr.Name, "./")
		if name == "" || hdr.Typeflag == tar.TypeDir {
			continue
		}
		names = append(names, name)
	}
	return names, nil
}

// ErrNotFound reports a missing member.
var ErrNotFound = errors.New("not found")

// Extract writes every file in the data tarball under dest. It is path-safe:
// any entry whose cleaned path would escape dest is rejected, so a hostile
// archive cannot perform directory traversal. Symlinks and hard links that
// point outside dest are likewise rejected.
func (p *Package) Extract(dest string) error {
	absDest, err := filepath.Abs(dest)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(absDest, 0o755); err != nil {
		return err
	}

	tr, err := tarReader(p.dataName, p.dataData)
	if err != nil {
		return err
	}
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		target, err := safeJoin(absDest, hdr.Name)
		if err != nil {
			return err
		}
		if err := extractEntry(tr, hdr, target, absDest); err != nil {
			return err
		}
	}
}

// safeJoin joins base and name, rejecting any name whose cleaned form would
// escape base (directory traversal). base must already be absolute. An absolute
// name is also rejected.
func safeJoin(base, name string) (string, error) {
	if filepath.IsAbs(name) {
		return "", fmt.Errorf("unsafe absolute path in archive: %q", name)
	}
	rel := strings.TrimPrefix(name, "./")
	joined := filepath.Join(base, rel)
	if joined != base && !strings.HasPrefix(joined, base+string(os.PathSeparator)) {
		return "", fmt.Errorf("unsafe path in archive: %q", name)
	}
	return joined, nil
}

// extractEntry materializes one tar entry at target.
func extractEntry(tr *tar.Reader, hdr *tar.Header, target, absDest string) error {
	switch hdr.Typeflag {
	case tar.TypeDir:
		return os.MkdirAll(target, os.FileMode(hdr.Mode)&os.ModePerm) //nolint:gosec // archive-defined mode
	case tar.TypeReg:
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		f, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.FileMode(hdr.Mode)&os.ModePerm) //nolint:gosec // path validated by safeJoin
		if err != nil {
			return err
		}
		if _, err := io.Copy(f, tr); err != nil { //nolint:gosec // bounded by archive
			_ = f.Close()
			return err
		}
		return f.Close()
	case tar.TypeSymlink:
		// Reject symlinks whose target would escape dest when resolved.
		if _, err := safeJoin(filepath.Dir(target), hdr.Linkname); err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		_ = os.Remove(target)
		return os.Symlink(hdr.Linkname, target)
	case tar.TypeLink:
		linkTarget, err := safeJoin(absDest, hdr.Linkname)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		_ = os.Remove(target)
		return os.Link(linkTarget, target)
	default:
		// Skip devices, fifos and other special files in the first slice.
		return nil
	}
}
