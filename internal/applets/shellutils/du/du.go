// Package du implements the du applet: estimate file space usage for files and
// directory trees, with the common GNU options.
package du

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"path"
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
	summarize     bool     // -s: print only a total for each operand
	all           bool     // -a: write a line for every file, not just directories
	human         bool     // -h: print sizes in human-readable form (1K, 234M, 2G)
	bytes         bool     // -b: print the apparent size in bytes
	total         bool     // -c: print a grand total
	apparentSize  bool     // --apparent-size: report exact apparent byte sizes, not block counts
	oneFileSystem bool     // -x/--one-file-system: do not cross filesystem boundaries
	maxDepth      int      // --max-depth=N: print totals only for dirs at depth <= N (-1 = unlimited)
	exclude       []string // --exclude=PATTERN: skip entries whose base name matches the glob
}

// entry is one accumulated path/size pair produced by the size walk.
type entry struct {
	path  string // path as it should be printed
	bytes int64  // apparent size in bytes of the subtree rooted at path
	isDir bool   // whether path is a directory
	depth int    // depth relative to the operand root (root = 0)
}

// NOTE: by default du reports the apparent total rounded up to whole 1K
// blocks (legacy MimixBox behaviour). --apparent-size (and -b) report the
// exact byte total instead. The walk only needs the apparent byte size.

// Run executes du.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... [FILE]...", stdio.Err).WithHelp(command.Help{
		Description: "Estimate and summarize disk space usage for each FILE, recursing into directories. " +
			"With no FILE, summarize the current directory.",
		Examples: []command.Example{
			{Command: "du", Explain: "Show usage of the current directory tree."},
			{Command: "du -sh /var/log", Explain: "Show a single human-readable total for /var/log."},
			{Command: "du -a dir", Explain: "Show usage for every file, not just directories."},
			{Command: "du --max-depth=1 dir", Explain: "Show totals for dir and its immediate subdirectories only."},
			{Command: "du --exclude='*.tmp' dir", Explain: "Skip entries whose base name matches the glob."},
		},
		ExitStatus: "0  success.\n1  an error occurred (e.g. a file could not be read).",
	})
	summarize := fs.BoolP("summarize", "s", false, "display only a total for each argument")
	all := fs.BoolP("all", "a", false, "write counts for all files, not just directories")
	human := fs.BoolP("human-readable", "h", false, "print sizes in human readable format (e.g., 1K 234M 2G)")
	bytesFlag := fs.BoolP("bytes", "b", false, "equivalent to apparent size in bytes")
	total := fs.BoolP("total", "c", false, "produce a grand total")
	fs.BoolP("block-size-1k", "k", true, "like --block-size=1K (the default)")
	apparentSize := fs.Bool("apparent-size", false, "print apparent (exact byte) sizes, rather than 1K block counts")
	oneFileSystem := fs.BoolP("one-file-system", "x", false, "skip directories on different file systems")
	maxDepth := fs.Int("max-depth", -1, "print the total for a directory only if it is N or fewer levels below the argument")
	exclude := fs.StringArray("exclude", nil, "exclude files that match PATTERN (glob on the base name)")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	opts := options{
		summarize:     *summarize,
		all:           *all,
		human:         *human,
		bytes:         *bytesFlag,
		total:         *total,
		apparentSize:  *apparentSize,
		oneFileSystem: *oneFileSystem,
		maxDepth:      *maxDepth,
		exclude:       *exclude,
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
		entries, err := walk(operand, opts)
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

// walk computes the size of every file and directory under root and returns the
// accumulated entries in post-order (children before their parent, the root
// last). It is a pure function of the filesystem: it does not touch stdout, so
// it can be unit-tested directly.
//
// The options influence the traversal: --one-file-system prunes subdirectories
// on a different device than root, and --exclude skips matching entries
// entirely (their sizes are not counted toward any ancestor).
func walk(root string, opts options) ([]entry, error) {
	type acc struct {
		bytes int64
		isDir bool
		depth int
	}
	sizes := map[string]*acc{}
	var order []string // paths in the order they were entered

	// rootDepth is the separator count of the operand root, so a child's depth
	// relative to the operand is depth(path) - rootDepth.
	rootDepth := depth(root)

	// rootDev is the device id of the operand root, used only when
	// --one-file-system is requested.
	var rootDev uint64
	if opts.oneFileSystem {
		dev, derr := deviceOf(root)
		if derr != nil {
			return nil, derr
		}
		rootDev = dev
	}

	walkErr := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// --exclude: skip any entry whose base name matches a pattern. The root
		// itself is never excluded (GNU du applies the filter to descendants).
		if p != root && excluded(p, opts.exclude) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// --one-file-system: prune subdirectories on a different device than the
		// operand root. The root directory is always counted.
		if opts.oneFileSystem && d.IsDir() && p != root {
			if dev, derr := deviceOf(p); derr == nil && dev != rootDev {
				return filepath.SkipDir
			}
		}

		if d.IsDir() {
			sizes[p] = &acc{isDir: true, depth: depth(p) - rootDepth}
			order = append(order, p)
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
		for dir := filepath.Dir(p); ; dir = filepath.Dir(dir) {
			if a, ok := sizes[dir]; ok {
				a.bytes += size
			}
			if dir == root || filepath.Dir(dir) == dir {
				break
			}
		}
		sizes[p] = &acc{bytes: size, isDir: false, depth: depth(p) - rootDepth}
		order = append(order, p)
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
		entries = append(entries, entry{path: p, bytes: a.bytes, isDir: a.isDir, depth: a.depth})
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
	root := entries[len(entries)-1]
	out := make([]entry, 0, len(entries))
	for _, e := range entries {
		// --max-depth: omit entries deeper than N levels below the operand,
		// but always keep the operand root itself.
		if opts.maxDepth >= 0 && e.depth > opts.maxDepth && e.path != root.path {
			continue
		}
		if opts.all {
			out = append(out, e)
			continue
		}
		// Default: print every directory, and always print the operand itself
		// even when it is a single regular file (GNU du behaves this way).
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

// formatSize renders a byte count according to the active output mode. -b and
// --apparent-size print the exact apparent byte count; -h prints a
// human-readable value; the default prints a count of 1024-byte blocks.
func formatSize(bytes int64, opts options) string {
	switch {
	case opts.bytes || opts.apparentSize:
		return strconv.FormatInt(bytes, 10)
	case opts.human:
		return humanReadable(bytes)
	default:
		return strconv.FormatInt(blocks(bytes), 10)
	}
}

// excluded reports whether the base name of p matches any of the glob patterns.
func excluded(p string, patterns []string) bool {
	if len(patterns) == 0 {
		return false
	}
	base := filepath.Base(p)
	for _, pat := range patterns {
		if ok, err := path.Match(pat, base); err == nil && ok {
			return true
		}
	}
	return false
}

// blocks converts a byte count to a count of 1024-byte blocks, rounding up (so
// any nonzero file occupies at least one block).
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
