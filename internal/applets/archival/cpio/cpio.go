// Package cpio implements the cpio applet: copy files into and out of a cpio
// archive in the portable "new ASCII" (newc, magic 070701) format. It supports
// the three classic modes: -o (copy-out, read names from stdin and write an
// archive), -i (copy-in, extract an archive) and -t (list, with -i).
package cpio

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

const (
	newcMagic = "070701"
	trailer   = "TRAILER!!!"
)

// Command is the cpio applet.
type Command struct{}

// New returns a cpio command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "cpio" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Copy files to and from archives" }

type options struct {
	create  bool
	extract bool
	list    bool
	verbose bool
}

// Run executes cpio.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "{-o|-i|-t} [-v] [-H newc]", stdio.Err).WithHelp(command.Help{
		Description: "Copy files to and from a cpio archive in the portable \"new ASCII\" (newc) format. " +
			"Exactly one of -o (copy-out), -i (copy-in), or -t (list, with -i) selects the operation.",
		Examples: []command.Example{
			{Command: "ls | cpio -o > archive.cpio", Explain: "Create an archive from the names read on stdin."},
			{Command: "cpio -i < archive.cpio", Explain: "Extract every file from the archive read on stdin."},
			{Command: "cpio -it < archive.cpio", Explain: "List the contents of the archive."},
		},
		ExitStatus: "0  success.\n1  an error occurred (e.g. the archive could not be read or written).",
	})
	create := fs.BoolP("create", "o", false, "copy-out: read file names from stdin, write archive to stdout")
	extract := fs.BoolP("extract", "i", false, "copy-in: read archive from stdin and extract")
	list := fs.BoolP("list", "t", false, "list the contents of the archive (with -i)")
	verbose := fs.BoolP("verbose", "v", false, "list files processed")
	format := fs.StringP("format", "H", "newc", "archive format (only 'newc' is supported)")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	if *format != "newc" {
		_, _ = fmt.Fprintf(stdio.Err, "cpio: unsupported format '%s' (only newc)\n", *format)
		return command.SilentFailure()
	}

	opts := options{create: *create, extract: *extract, list: *list, verbose: *verbose}

	var runErr error
	switch {
	case opts.create:
		runErr = copyOut(stdio, opts)
	case opts.list:
		runErr = copyIn(stdio, opts, true)
	case opts.extract:
		runErr = copyIn(stdio, opts, false)
	default:
		_, _ = fmt.Fprintln(stdio.Err, "cpio: you must specify one of -o, -i or -t")
		return command.SilentFailure()
	}
	if runErr != nil {
		_, _ = fmt.Fprintf(stdio.Err, "cpio: %v\n", runErr)
		return command.SilentFailure()
	}
	return nil
}

// copyOut reads newline-separated file names from stdin and writes a newc
// archive to stdout.
func copyOut(stdio command.IO, opts options) error {
	scanner := bufio.NewScanner(stdio.In)
	ino := 1
	for scanner.Scan() {
		name := strings.TrimSpace(scanner.Text())
		if name == "" {
			continue
		}
		info, err := os.Stat(name)
		if err != nil {
			return err
		}
		var data []byte
		if info.Mode().IsRegular() {
			data, err = os.ReadFile(name) // #nosec G304 -- archiving a user-named file
			if err != nil {
				return err
			}
		}
		if err := writeEntry(stdio.Out, ino, uint32(info.Mode()), info.ModTime().Unix(), name, data); err != nil {
			return err
		}
		if opts.verbose {
			_, _ = fmt.Fprintln(stdio.Err, name)
		}
		ino++
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return writeEntry(stdio.Out, ino, 0, 0, trailer, nil)
}

// writeEntry writes one newc header, the NUL-terminated name and the file data,
// each padded to a 4-byte boundary.
func writeEntry(w io.Writer, ino int, mode uint32, mtime int64, name string, data []byte) error {
	nameBytes := append([]byte(name), 0)
	// newc size fields are 32 bits wide; a larger file cannot be stored.
	if uint64(len(data)) > math.MaxUint32 || uint64(len(nameBytes)) > math.MaxUint32 {
		return fmt.Errorf("%s: too large for the newc format", name)
	}
	fields := []uint32{
		uint32(ino), // #nosec G115 -- inode numbers only tell entries apart; newc stores 32 bits and wraps like GNU cpio
		mode, 0, 0, 1,
		uint32(mtime),     // #nosec G115 -- newc stores mtime as an unsigned 32-bit field, as GNU cpio does
		uint32(len(data)), // #nosec G115 -- checked against MaxUint32 above
		0, 0, 0, 0,
		uint32(len(nameBytes)), // #nosec G115 -- checked against MaxUint32 above
		0,
	}
	var b strings.Builder
	b.WriteString(newcMagic)
	for _, f := range fields {
		fmt.Fprintf(&b, "%08X", f)
	}
	if _, err := io.WriteString(w, b.String()); err != nil {
		return err
	}
	if _, err := w.Write(nameBytes); err != nil {
		return err
	}
	// Header (110) + name padded to 4 bytes.
	if err := pad(w, 110+len(nameBytes)); err != nil {
		return err
	}
	if _, err := w.Write(data); err != nil {
		return err
	}
	return pad(w, len(data))
}

// pad writes NUL bytes so that n becomes a multiple of 4.
func pad(w io.Writer, n int) error {
	if r := n % 4; r != 0 {
		_, err := w.Write(make([]byte, 4-r))
		return err
	}
	return nil
}

// copyIn reads a newc archive from stdin and either lists (listOnly) or
// extracts every entry.
func copyIn(stdio command.IO, opts options, listOnly bool) error {
	r := bufio.NewReader(stdio.In)
	for {
		hdr := make([]byte, 110)
		if _, err := io.ReadFull(r, hdr); err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		if string(hdr[0:6]) != newcMagic {
			return fmt.Errorf("not a newc cpio archive")
		}
		nameSize, err := hexField(hdr, 94)
		if err != nil {
			return err
		}
		fileSize, err := hexField(hdr, 54)
		if err != nil {
			return err
		}
		mode, err := hexField(hdr, 14)
		if err != nil {
			return err
		}

		// Read through a growing buffer instead of allocating the size the
		// header claims: a 110-byte header can claim 4 GiB.
		nameBuf, err := readN(r, nameSize)
		if err != nil {
			return err
		}
		name := strings.TrimRight(string(nameBuf), "\x00")
		if err := skip(r, 110+int(nameSize)); err != nil {
			return err
		}

		if name == trailer {
			return nil
		}

		data, err := readN(r, fileSize)
		if err != nil {
			return err
		}
		if err := skip(r, int(fileSize)); err != nil {
			return err
		}

		if listOnly {
			_, _ = fmt.Fprintln(stdio.Out, name)
			continue
		}
		if err := extractEntry(name, os.FileMode(mode), data); err != nil {
			return err
		}
		if opts.verbose {
			_, _ = fmt.Fprintln(stdio.Err, name)
		}
	}
}

// hexField parses the 8-char hex newc field at offset off in hdr.
func hexField(hdr []byte, off int) (uint32, error) {
	v, err := strconv.ParseUint(string(hdr[off:off+8]), 16, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid header field %q", hdr[off:off+8])
	}
	return uint32(v), nil // #nosec G115 -- ParseUint with bitSize 32 returns at most MaxUint32
}

// readN reads exactly n bytes from r, growing the buffer only as data
// actually arrives.
func readN(r io.Reader, n uint32) ([]byte, error) {
	var buf bytes.Buffer
	if _, err := io.CopyN(&buf, r, int64(n)); err != nil {
		if err == io.EOF {
			return nil, io.ErrUnexpectedEOF
		}
		return nil, err
	}
	return buf.Bytes(), nil
}

// skip discards the padding bytes that align n to a 4-byte boundary.
func skip(r *bufio.Reader, n int) error {
	if rem := n % 4; rem != 0 {
		_, err := io.CopyN(io.Discard, r, int64(4-rem))
		return err
	}
	return nil
}

// extractEntry writes one extracted entry (directory or regular file) to disk.
func extractEntry(name string, mode os.FileMode, data []byte) error {
	if mode.IsDir() {
		return os.MkdirAll(name, mode.Perm()|0o700)
	}
	if dir := filepath.Dir(name); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil { // #nosec G301 -- parent directories of extracted entries are created 0755 before umask, no wider than the 0777-minus-umask GNU cpio uses
			return err
		}
	}
	perm := mode.Perm()
	if perm == 0 {
		perm = 0o644
	}
	return os.WriteFile(name, data, perm) //nolint:gosec // archive-defined mode
}
