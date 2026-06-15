// Package makedevs implements the makedevs applet: build a device tree under a
// root directory from a textual device table.
package makedevs

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the makedevs applet.
type Command struct{}

// New returns a makedevs command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "makedevs" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Create a device tree from a table" }

// nodeMaker creates a device node (mknod). It is injected so unit tests can
// exercise the planner without privilege and assert on the planned nodes.
var nodeMaker NodeMaker = osNodeMaker{}

// NodeMaker abstracts creation of device-special files so tests stay hermetic
// and unprivileged.
type NodeMaker interface {
	// Mknod creates a device node at path of the given kind ('c' or 'b'),
	// permission mode, and major/minor numbers.
	Mknod(path string, kind byte, mode os.FileMode, major, minor uint32) error
}

// entry is one parsed device-table row.
type entry struct {
	path  string
	typ   byte // f=file d=dir c=char b=block p=fifo
	mode  os.FileMode
	uid   int
	gid   int
	major uint32
	minor uint32
	start uint32
	inc   uint32
	count uint32
}

// Run executes makedevs.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "-d TABLE ROOTDIR", stdio.Err).WithHelp(command.Help{
		Description: "Create a tree of directories, regular files, named pipes, and device nodes under ROOTDIR " +
			"from a device TABLE. Each table row is: 'name type mode uid gid major minor start inc count'. " +
			"type is d (directory), f (file), p (fifo), c (character device), or b (block device). When " +
			"count is non-zero the row expands to numbered nodes name0..name{count-1}. Creating device " +
			"nodes (c/b) needs privilege: without it makedevs fails with a documented error instead of a " +
			"silent skip. Directories, files, and fifos are always created.",
		Examples: []command.Example{
			{Command: "makedevs -d device_table.txt ./rootfs", Explain: "Populate ./rootfs from the table."},
		},
		ExitStatus: "0  the whole table was applied.\n1  the table was malformed or a node could not be created.",
		Notes: []string{
			"A leading '#' marks a comment line in the table; blank lines are ignored.",
			"Device-node creation is capability-gated; run with the privilege to mknod for c/b rows.",
		},
	})
	table := fs.StringP("table", "d", "", "device table file ('-' for standard input)")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	rest := fs.Args()
	if *table == "" || len(rest) != 1 {
		return command.Failuref("usage: makedevs -d TABLE ROOTDIR")
	}
	root := rest[0]

	r, closeFn, err := openTable(stdio, *table)
	if err != nil {
		return command.Failuref("%s", command.FileError(*table, err))
	}
	defer closeFn()

	entries, err := parseTable(r)
	if err != nil {
		return command.Failuref("%v", err)
	}

	if err := os.MkdirAll(root, 0o755); err != nil {
		return command.Failuref("%s", command.FileError(root, err))
	}
	for _, e := range entries {
		if err := apply(root, e); err != nil {
			return command.Failuref("%s: %v", e.path, err)
		}
	}
	return nil
}

// openTable resolves the -d operand to a reader, honoring "-" for stdin.
func openTable(stdio command.IO, name string) (io.Reader, func(), error) {
	if name == "-" {
		return stdio.In, func() {}, nil
	}
	f, err := os.Open(name) //nolint:gosec // user-named table file
	if err != nil {
		return nil, func() {}, err
	}
	return f, func() { _ = f.Close() }, nil
}

// parseTable parses a device table into entries, validating each row.
func parseTable(r io.Reader) ([]entry, error) {
	var entries []entry
	sc := bufio.NewScanner(r)
	line := 0
	for sc.Scan() {
		line++
		text := strings.TrimSpace(sc.Text())
		if text == "" || strings.HasPrefix(text, "#") {
			continue
		}
		e, err := parseRow(text)
		if err != nil {
			return nil, fmt.Errorf("line %d: %v", line, err)
		}
		entries = append(entries, e)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}

// parseRow parses one device-table row.
func parseRow(text string) (entry, error) {
	fields := strings.Fields(text)
	if len(fields) < 10 {
		return entry{}, fmt.Errorf("expected 10 fields, got %d", len(fields))
	}
	if len(fields[1]) != 1 || !strings.ContainsAny(fields[1], "fdcbp") {
		return entry{}, fmt.Errorf("invalid type %q (want f, d, c, b, or p)", fields[1])
	}
	mode, err := strconv.ParseUint(fields[2], 8, 32)
	if err != nil {
		return entry{}, fmt.Errorf("invalid mode %q", fields[2])
	}
	nums := make([]int64, 7)
	for i, f := range fields[3:10] {
		n, err := parseNum(f)
		if err != nil {
			return entry{}, fmt.Errorf("invalid number %q", f)
		}
		nums[i] = n
	}
	return entry{
		path:  fields[0],
		typ:   fields[1][0],
		mode:  os.FileMode(mode),
		uid:   int(nums[0]),
		gid:   int(nums[1]),
		major: uint32(nums[2]),
		minor: uint32(nums[3]),
		start: uint32(nums[4]),
		inc:   uint32(nums[5]),
		count: uint32(nums[6]),
	}, nil
}

// parseNum parses a table number; "-" means zero (the BusyBox convention for an
// unused field).
func parseNum(s string) (int64, error) {
	if s == "-" {
		return 0, nil
	}
	return strconv.ParseInt(s, 0, 64)
}

// apply realizes one entry under root, expanding numbered device rows.
func apply(root string, e entry) error {
	target := filepath.Join(root, e.path)
	switch e.typ {
	case 'd':
		return os.MkdirAll(target, e.mode.Perm())
	case 'f':
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY, e.mode.Perm()) //nolint:gosec // table-driven path under root
		if err != nil {
			return err
		}
		return f.Close()
	case 'p':
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return nodeMaker.Mknod(target, 'p', e.mode.Perm(), 0, 0)
	case 'c', 'b':
		return applyNodes(root, e)
	}
	return fmt.Errorf("unsupported type %q", string(e.typ))
}

// applyNodes creates one device node, or a numbered range when count > 0.
func applyNodes(root string, e entry) error {
	if err := os.MkdirAll(filepath.Join(root, filepath.Dir(e.path)), 0o755); err != nil {
		return err
	}
	if e.count == 0 {
		return nodeMaker.Mknod(filepath.Join(root, e.path), e.typ, e.mode.Perm(), e.major, e.minor)
	}
	for i := uint32(0); i < e.count; i++ {
		name := fmt.Sprintf("%s%d", e.path, e.start+i)
		minor := e.minor + i*max1(e.inc)
		if err := nodeMaker.Mknod(filepath.Join(root, name), e.typ, e.mode.Perm(), e.major, minor); err != nil {
			return err
		}
	}
	return nil
}

// max1 returns inc, treating 0 as 1 so numbered minors advance by at least one.
func max1(inc uint32) uint32 {
	if inc == 0 {
		return 1
	}
	return inc
}
