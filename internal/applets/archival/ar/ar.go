// Package ar implements the ar applet: create, list and extract Unix "common"
// (System V/GNU) ar archives, the format used by .a static libraries and Debian
// .deb packages. It supports the everyday operations: r (replace/create), t
// (list) and x (extract).
package ar

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// magic is the ar archive global header.
const magic = "!<arch>\n"

// headerSize is the fixed size of each per-member header.
const headerSize = 60

// Command is the ar applet.
type Command struct{}

// New returns an ar command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "ar" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Create, modify and extract from archives" }

// Run executes ar. The first operand is a key letter (r, t or x) optionally
// preceded by a dash, matching ar's traditional calling convention.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "{rtx}[v] ARCHIVE [MEMBER]...", stdio.Err)
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	rest := fs.Args()
	if len(rest) < 2 {
		_, _ = fmt.Fprintln(stdio.Err, "ar: usage: ar {rtx}[v] ARCHIVE [MEMBER]...")
		return command.SilentFailure()
	}

	key := strings.TrimPrefix(rest[0], "-")
	archive := rest[1]
	members := rest[2:]
	verbose := strings.ContainsRune(key, 'v')

	var runErr error
	switch {
	case strings.ContainsRune(key, 'r'):
		runErr = create(archive, members, verbose, stdio)
	case strings.ContainsRune(key, 't'):
		runErr = list(archive, stdio)
	case strings.ContainsRune(key, 'x'):
		runErr = extract(archive, members, verbose, stdio)
	default:
		_, _ = fmt.Fprintf(stdio.Err, "ar: invalid operation key '%s' (use r, t or x)\n", key)
		return command.SilentFailure()
	}
	if runErr != nil {
		_, _ = fmt.Fprintf(stdio.Err, "ar: %v\n", runErr)
		return command.SilentFailure()
	}
	return nil
}

// member is one entry parsed from an archive.
type member struct {
	name string
	mode int64
	data []byte
}

// create writes an archive containing the named files.
func create(archive string, files []string, verbose bool, stdio command.IO) error {
	if len(files) == 0 {
		return fmt.Errorf("no members specified")
	}
	out, err := os.Create(archive) //nolint:gosec // user-named file
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()

	if _, err := io.WriteString(out, magic); err != nil {
		return err
	}
	for _, f := range files {
		info, err := os.Stat(f)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(f) //nolint:gosec // archiving a user-named file
		if err != nil {
			return err
		}
		if err := writeMember(out, filepath.Base(f), int64(info.Mode().Perm()), info.ModTime().Unix(), data); err != nil {
			return err
		}
		if verbose {
			_, _ = fmt.Fprintf(stdio.Out, "a - %s\n", filepath.Base(f))
		}
	}
	return nil
}

// writeMember writes one member header and its (even-padded) data.
func writeMember(w io.Writer, name string, mode, mtime int64, data []byte) error {
	hdr := fmt.Sprintf("%-16s%-12d%-6d%-6d%-8o%-10d`\n",
		name+"/", mtime, 0, 0, mode, len(data))
	if len(hdr) != headerSize {
		hdr = fmt.Sprintf("%-16s%-12d%-6d%-6d%-8o%-10d`\n", trunc(name+"/", 16), mtime, 0, 0, mode, len(data))
	}
	if _, err := io.WriteString(w, hdr); err != nil {
		return err
	}
	if _, err := w.Write(data); err != nil {
		return err
	}
	if len(data)%2 == 1 {
		if _, err := io.WriteString(w, "\n"); err != nil {
			return err
		}
	}
	return nil
}

// trunc shortens s to at most n bytes.
func trunc(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s
}

// readArchive parses every member of an archive.
func readArchive(archive string) ([]member, error) {
	f, err := os.Open(archive) //nolint:gosec // user-named file
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	hdr := make([]byte, len(magic))
	if _, err := io.ReadFull(f, hdr); err != nil {
		return nil, fmt.Errorf("not an ar archive")
	}
	if string(hdr) != magic {
		return nil, fmt.Errorf("not an ar archive")
	}

	var members []member
	for {
		h := make([]byte, headerSize)
		_, err := io.ReadFull(f, h)
		if err == io.EOF {
			return members, nil
		}
		if err != nil {
			return nil, err
		}
		name := strings.TrimRight(string(h[0:16]), " ")
		name = strings.TrimSuffix(name, "/")
		mode, _ := strconv.ParseInt(strings.TrimSpace(string(h[40:48])), 8, 64)
		size, err := strconv.ParseInt(strings.TrimSpace(string(h[48:58])), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("corrupt member header")
		}
		data := make([]byte, size)
		if _, err := io.ReadFull(f, data); err != nil {
			return nil, err
		}
		if size%2 == 1 {
			if _, err := f.Seek(1, io.SeekCurrent); err != nil {
				return nil, err
			}
		}
		members = append(members, member{name: name, mode: mode, data: data})
	}
}

// list prints the member names in the archive.
func list(archive string, stdio command.IO) error {
	members, err := readArchive(archive)
	if err != nil {
		return err
	}
	for _, m := range members {
		_, _ = fmt.Fprintln(stdio.Out, m.name)
	}
	return nil
}

// extract writes the requested members (or all of them) to the current
// directory.
func extract(archive string, want []string, verbose bool, stdio command.IO) error {
	members, err := readArchive(archive)
	if err != nil {
		return err
	}
	wanted := map[string]bool{}
	for _, w := range want {
		wanted[w] = true
	}
	for _, m := range members {
		if len(want) > 0 && !wanted[m.name] {
			continue
		}
		mode := os.FileMode(m.mode)
		if mode == 0 {
			mode = 0o644
		}
		if err := os.WriteFile(m.name, m.data, mode); err != nil { //nolint:gosec // archive-defined mode
			return err
		}
		if verbose {
			_, _ = fmt.Fprintf(stdio.Out, "x - %s\n", m.name)
		}
	}
	return nil
}
