// Package df implements the df applet: report file system disk space usage for
// the file system that contains each FILE operand (or the current directory's
// file system when no operand is given).
package df

import (
	"context"
	"fmt"
	"io"
	"syscall"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the df applet.
type Command struct{}

// New returns a df command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "df" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Report file system disk space usage" }

// statfsResult is the subset of syscall.Statfs_t that df cares about. Wrapping
// it lets a test inject a fake without touching the real filesystem.
type statfsResult struct {
	Bsize  int64  // fundamental block size
	Blocks uint64 // total data blocks
	Bfree  uint64 // free blocks in filesystem
	Bavail uint64 // free blocks available to unprivileged users
	Files  uint64 // total file nodes (inodes)
	Ffree  uint64 // free file nodes (inodes)
}

// statfs resolves path to its file system statistics. It is a package var so a
// test can replace it with a deterministic fake.
var statfs = func(path string) (statfsResult, error) {
	var s syscall.Statfs_t
	if err := syscall.Statfs(path, &s); err != nil {
		return statfsResult{}, err
	}
	return statfsResult{
		Bsize:  int64(s.Bsize),
		Blocks: s.Blocks,
		Bfree:  s.Bfree,
		Bavail: s.Bavail,
		Files:  s.Files,
		Ffree:  s.Ffree,
	}, nil
}

type options struct {
	human  bool
	inodes bool
}

// usage is the computed disk usage of a file system, in bytes.
type usage struct {
	total  uint64
	used   uint64
	avail  uint64
	usePct int
}

// inodeUsage is the computed inode usage of a file system.
type inodeUsage struct {
	files  uint64
	used   uint64
	free   uint64
	usePct int
}

// computeUsage turns a Statfs result into byte-based disk usage figures. It is a
// pure function so it can be unit-tested without a real filesystem.
func computeUsage(s statfsResult) usage {
	bsize := uint64(s.Bsize)
	total := s.Blocks * bsize
	free := s.Bfree * bsize
	avail := s.Bavail * bsize
	used := total - free
	return usage{
		total:  total,
		used:   used,
		avail:  avail,
		usePct: percent(used, used+avail),
	}
}

// computeInodeUsage turns a Statfs result into inode usage figures.
func computeInodeUsage(s statfsResult) inodeUsage {
	used := s.Files - s.Ffree
	return inodeUsage{
		files:  s.Files,
		used:   used,
		free:   s.Ffree,
		usePct: percent(used, s.Files),
	}
}

// percent computes the rounded-up "Use%" the way GNU df does: used out of the
// usable total (used+available), rounded toward the next whole percent.
func percent(used, total uint64) int {
	if total == 0 {
		return 0
	}
	// Round up: ceil(used*100/total).
	return int((used*100 + total - 1) / total)
}

// humanReadable formats n bytes using powers of 1024 (e.g. 1024 -> "1.0K",
// 1048576 -> "1.0M"). It is a pure function so it can be unit-tested directly.
func humanReadable(n uint64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d", n)
	}
	units := []string{"K", "M", "G", "T", "P", "E"}
	value := float64(n)
	i := -1
	for value >= unit && i < len(units)-1 {
		value /= unit
		i++
	}
	return fmt.Sprintf("%.1f%s", value, units[i])
}

// formatSize renders a byte count either human-readable or in 1K blocks.
func formatSize(bytes uint64, human bool) string {
	if human {
		return humanReadable(bytes)
	}
	// 1K blocks, rounded up so a non-empty filesystem never shows 0.
	return fmt.Sprintf("%d", (bytes+1023)/1024)
}

// Run executes df.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... [FILE]...", stdio.Err)
	human := fs.BoolP("human-readable", "h", false, "print sizes in powers of 1024 (e.g., 1023M)")
	_ = fs.BoolP("kilobytes", "k", false, "use 1024-byte (1K) blocks (default)")
	inodes := fs.BoolP("inodes", "i", false, "list inode information instead of block usage")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	opts := options{human: *human, inodes: *inodes}

	operands := fs.Args()
	if len(operands) == 0 {
		operands = []string{"."}
	}

	writeHeader(stdio.Out, opts)

	var firstErr error
	for _, path := range operands {
		s, serr := statfs(path)
		if serr != nil {
			_, _ = fmt.Fprintf(stdio.Err, "df: %s\n", command.FileError(path, serr))
			firstErr = keep(firstErr)
			continue
		}
		writeRow(stdio.Out, path, s, opts)
	}
	return firstErr
}

// writeHeader prints the column header appropriate for the selected mode.
func writeHeader(w io.Writer, opts options) {
	if opts.inodes {
		_, _ = fmt.Fprintf(w, "%-20s %10s %10s %10s %4s %s\n",
			"Filesystem", "Inodes", "IUsed", "IFree", "IUse%", "Mounted on")
		return
	}
	size := "1K-blocks"
	if opts.human {
		size = "Size"
	}
	_, _ = fmt.Fprintf(w, "%-20s %10s %10s %10s %4s %s\n",
		"Filesystem", size, "Used", "Available", "Use%", "Mounted on")
}

// writeRow prints one data row for path using the file system statistics in s.
func writeRow(w io.Writer, path string, s statfsResult, opts options) {
	_, _ = io.WriteString(w, formatRow(path, s, opts))
}

// formatRow renders a single df row as a string. Kept separate from writeRow so
// it is straightforward to assert on in a unit test.
func formatRow(path string, s statfsResult, opts options) string {
	if opts.inodes {
		u := computeInodeUsage(s)
		return fmt.Sprintf("%-20s %10d %10d %10d %3d%% %s\n",
			path, u.files, u.used, u.free, u.usePct, path)
	}
	u := computeUsage(s)
	return fmt.Sprintf("%-20s %10s %10s %10s %3d%% %s\n",
		path,
		formatSize(u.total, opts.human),
		formatSize(u.used, opts.human),
		formatSize(u.avail, opts.human),
		u.usePct,
		path)
}

// keep preserves the first error so the exit code reflects a failure while the
// message has already been printed to stderr.
func keep(existing error) error {
	if existing != nil {
		return existing
	}
	return command.SilentFailure()
}
