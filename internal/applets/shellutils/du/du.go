// Package du implements the du applet: estimate file space usage for files and
// directory trees, with the common GNU options.
package du

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"sort"
	"strconv"

	"github.com/nao1215/mimixbox/internal/command"
)

// blockSize is the unit du reports in by default. GNU du historically reports
// 1024-byte blocks; MimixBox computes every entry from the file's apparent
// size rounded up to this many bytes so the result is deterministic and does
// not depend on the underlying filesystem's st_blocks.
const blockSize = 1024

// Command is the du applet.
type Command struct{}

// New returns a du command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "du" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Estimate file space usage" }

type options struct {
	summarize bool // -s: print only a total for each operand
	all       bool // -a: write a line for every file, not just directories
	human     bool // -h: print sizes in human-readable form (1K, 234M, 2G)
	bytes     bool // -b: print the apparent size in bytes
	total     bool // -c: print a grand total
}

// entry is one accumulated path/size pair produced by the size walk.
type entry struct {
	path  string // path as it should be printed
	bytes int64  // apparent size in bytes of the subtree rooted at path
	isDir bool   // whether path is a directory
}

// Run executes du.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... [FILE]...", stdio.Err).WithHelp(command.Help{
		Description: "Estimate and summarize disk space usage for each FILE, recursing into directories. " +
			"With no FILE, summarize the current directory.",
		Examples: []command.Example{
			{Command: "du", Explain: "Show usage of the current directory tree."},
			{Command: "du -sh /var/log", Explain: "Show a single human-readable total for /var/log."},
			{Command: "du -a dir", Explain: "Show usage for every file, not just directories."},
		},
		ExitStatus: "0  success.\n1  an error occurred (e.g. a file could not be read).",
	})
	summarize := fs.BoolP("summarize", "s", false, "display only a total for each argument")
	all := fs.BoolP("all", "a", false, "write counts for all files, not just directories")
	human := fs.BoolP("human-readable", "h", false, "print sizes in human readable format (e.g., 1K 234M 2G)")
	bytesFlag := fs.BoolP("bytes", "b", false, "equivalent to apparent size in bytes")
	total := fs.BoolP("total", "c", false, "produce a grand total")
	fs.BoolP("block-size-1k", "k", true, "like --block-size=1K (the default)")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	opts := options{
		summarize: *summarize,
		all:       *all,
		human:     *human,
		bytes:     *bytesFlag,
		total:     *total,
	}

	operands := fs.Args()
	if len(operands) == 0 {
		operands = []string{"."}
	}

	return run(stdio, operands, opts)
}

// run walks each operand and prints the requested report. A failed walk is
// reported on stderr but does not stop the remaining operands; the returned
// error only sets the exit code.
func run(stdio command.IO, operands []string, opts options) error {
	var grand int64
	var firstErr error

	for _, operand := range operands {
		entries, err := walk(operand)
		if err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "du: %s\n", command.FileError(operand, err))
			if firstErr == nil {
				firstErr = command.SilentFailure()
			}
			continue
		}
		printed := report(entries, opts)
		for _, e := range printed {
			writeLine(stdio.Out, e, opts)
		}
		if len(entries) > 0 {
			// The total for the operand is the size of its root entry, which
			// walk always places last.
			grand += entries[len(entries)-1].bytes
		}
	}

	if opts.total {
		writeLine(stdio.Out, entry{path: "total", bytes: grand}, opts)
	}
	return firstErr
}

// walk computes the apparent size of every file and directory under root and
// returns the accumulated entries in post-order (children before their parent,
// the root last). It is a pure function of the filesystem: it does not touch
// stdout, so it can be unit-tested directly.
func walk(root string) ([]entry, error) {
	type acc struct {
		bytes int64
		isDir bool
	}
	sizes := map[string]*acc{}
	var order []string // directory paths in the order they were entered

	walkErr := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			sizes[path] = &acc{bytes: 0, isDir: true}
			order = append(order, path)
			return nil
		}
		info, ierr := d.Info()
		if ierr != nil {
			return ierr
		}
		// Count only regular-file content: a directory's own inode size is
		// filesystem-dependent and would make the result nondeterministic.
		// Store the apparent (raw) byte size; rounding to 1K blocks happens
		// only when the block count is printed.
		size := info.Size()
		// Add this file's size to every ancestor directory up to root.
		for dir := filepath.Dir(path); ; dir = filepath.Dir(dir) {
			if a, ok := sizes[dir]; ok {
				a.bytes += size
			}
			if dir == root || filepath.Dir(dir) == dir {
				break
			}
		}
		sizes[path] = &acc{bytes: size, isDir: false}
		order = append(order, path)
		return nil
	})
	if walkErr != nil {
		return nil, walkErr
	}

	// Emit children before parents (deeper paths first), with the root last,
	// matching GNU du's default ordering.
	sort.SliceStable(order, func(i, j int) bool {
		return depth(order[i]) > depth(order[j])
	})
	entries := make([]entry, 0, len(order))
	for _, p := range order {
		a := sizes[p]
		entries = append(entries, entry{path: p, bytes: a.bytes, isDir: a.isDir})
	}
	return entries, nil
}

// report filters the accumulated entries down to the lines that should be
// printed given the options, preserving order.
func report(entries []entry, opts options) []entry {
	if len(entries) == 0 {
		return nil
	}
	if opts.summarize {
		// Only the root (always last) is printed.
		return entries[len(entries)-1:]
	}
	if opts.all {
		return entries
	}
	root := entries[len(entries)-1]
	out := make([]entry, 0, len(entries))
	for _, e := range entries {
		// Print every directory, and always print the operand itself even
		// when it is a single regular file (GNU du behaves this way).
		if e.isDir || e.path == root.path {
			out = append(out, e)
		}
	}
	return out
}

// writeLine prints a single "SIZE\tPATH" line for e.
func writeLine(w io.Writer, e entry, opts options) {
	_, _ = fmt.Fprintf(w, "%s\t%s\n", formatSize(e.bytes, opts), e.path)
}

// formatSize renders a byte count according to the active output mode: bytes
// (-b), human-readable (-h), or the default 1024-byte block count.
func formatSize(bytes int64, opts options) string {
	switch {
	case opts.bytes:
		return strconv.FormatInt(bytes, 10)
	case opts.human:
		return humanReadable(bytes)
	default:
		return strconv.FormatInt(blocks(bytes), 10)
	}
}

// blocks converts an apparent byte count to a count of 1024-byte blocks,
// rounding up (so any nonzero file occupies at least one block).
func blocks(bytes int64) int64 {
	if bytes <= 0 {
		return 0
	}
	return (bytes + blockSize - 1) / blockSize
}

// depth returns the number of path separators in p, used to order children
// before their parents.
func depth(p string) int {
	n := 0
	for _, r := range p {
		if r == filepath.Separator {
			n++
		}
	}
	return n
}

// humanReadable formats a byte count the way GNU du -h does: the largest unit
// for which the value is at least 1, with one decimal place below 10 (e.g.
// 1536 -> "1.5K", 2147483648 -> "2.0G").
func humanReadable(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return strconv.FormatInt(bytes, 10)
	}
	units := []string{"K", "M", "G", "T", "P", "E"}
	value := float64(bytes)
	var suffix string
	for _, u := range units {
		value /= unit
		suffix = u
		if value < unit {
			break
		}
	}
	if value < 10 {
		return fmt.Sprintf("%.1f%s", value, suffix)
	}
	return fmt.Sprintf("%.0f%s", value, suffix)
}
