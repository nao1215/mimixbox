// Package df implements the df applet: report file system disk space usage for
// the file system that contains each FILE operand (or the current directory's
// file system when no operand is given).
package df

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
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
	Type   int64  // filesystem type magic number (Linux f_type)
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
		Type:   int64(s.Type),
	}, nil
}

// mountEntry describes one mounted filesystem from the system mount table.
type mountEntry struct {
	source string // device or remote source (e.g. /dev/sda1, tmpfs)
	target string // mount point (e.g. /, /home)
	fstype string // filesystem type (e.g. ext4, tmpfs)
}

// readMounts returns the system mount table. It is a package var so a test can
// replace it with a deterministic fake (the injectable mount source seam).
var readMounts = func() ([]mountEntry, error) {
	f, err := os.Open("/proc/self/mounts")
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	var entries []mountEntry
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) < 3 {
			continue
		}
		entries = append(entries, mountEntry{
			source: unescapeMount(fields[0]),
			target: unescapeMount(fields[1]),
			fstype: fields[2],
		})
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}

// unescapeMount decodes the octal escapes (\040 for space, etc.) that the
// kernel writes into /proc/self/mounts for whitespace in paths.
func unescapeMount(s string) string {
	if !strings.Contains(s, `\`) {
		return s
	}
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+3 < len(s) {
			if n, err := strconv.ParseUint(s[i+1:i+4], 8, 8); err == nil {
				b.WriteByte(byte(n))
				i += 3
				continue
			}
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

type options struct {
	human     bool
	inodes    bool
	all       bool
	total     bool
	types     []string // --type filters; empty means no filter
	output    []string // --output column list; empty means classic layout
	blockSize int64    // bytes per block when scaling sizes (0 == 1K default)
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

// scaleSize renders a byte count for the column-based output. When human is set
// it uses powers-of-1024; otherwise it scales by blockSize (0 means 1K blocks),
// rounding up so a non-empty filesystem never reports 0.
func scaleSize(bytes uint64, human bool, blockSize int64) string {
	if human {
		return humanReadable(bytes)
	}
	bs := uint64(1024)
	if blockSize > 0 {
		bs = uint64(blockSize)
	}
	return fmt.Sprintf("%d", (bytes+bs-1)/bs)
}

// parseSize parses a size like "1024", "K", "1M", "512" into a byte count. A
// bare suffix (e.g. "K") means one unit. Mirrors the ls --block-size parser.
func parseSize(spec string) (int64, error) {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return 0, fmt.Errorf("empty size")
	}
	i := 0
	for i < len(spec) && spec[i] >= '0' && spec[i] <= '9' {
		i++
	}
	numPart := spec[:i]
	unitPart := strings.ToUpper(spec[i:])

	var base int64
	switch unitPart {
	case "":
		base = 1
	case "K", "KIB":
		base = 1024
	case "M", "MIB":
		base = 1024 * 1024
	case "G", "GIB":
		base = 1024 * 1024 * 1024
	case "T", "TIB":
		base = 1024 * 1024 * 1024 * 1024
	case "KB":
		base = 1000
	case "MB":
		base = 1000 * 1000
	case "GB":
		base = 1000 * 1000 * 1000
	default:
		return 0, fmt.Errorf("unknown unit %q", unitPart)
	}

	if numPart == "" {
		return base, nil
	}
	n, err := strconv.ParseInt(numPart, 10, 64)
	if err != nil {
		return 0, err
	}
	if n <= 0 {
		return 0, fmt.Errorf("non-positive size")
	}
	return n * base, nil
}

// fsTypeName maps a Linux f_type magic number to a human-readable filesystem
// name. Unknown values are rendered as a hex magic so a row is still labeled.
func fsTypeName(magic int64) string {
	if name, ok := fsMagic[magic]; ok {
		return name
	}
	return fmt.Sprintf("0x%x", uint64(magic))
}

// fsMagic is a small table of common Linux filesystem magic numbers. df uses it
// only as a fallback when the mount table does not provide a type name.
var fsMagic = map[int64]string{
	0xef53:     "ext",
	0x58465342: "xfs",
	0x9123683e: "btrfs",
	0x01021994: "tmpfs",
	0x6969:     "nfs",
	0xff534d42: "cifs",
	0x4d44:     "msdos",
	0x65735546: "fuse",
	0x794c7630: "overlay",
	0x62656572: "sysfs",
	0x9fa0:     "proc",
	0x1cd1:     "devpts",
	0x858458f6: "ramfs",
	0x73717368: "squashfs",
}

// fsEntry is one filesystem to be reported: its identity (source/fstype/target)
// joined with its space and inode statistics.
type fsEntry struct {
	source  string
	fstype  string
	target  string
	stat    statfsResult
	operand string // non-empty in operand mode; the FILE the user named
}

// Run executes df.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... [FILE]...", stdio.Err).WithHelp(command.Help{
		Description: "Report file system disk space usage for the file system that contains each FILE, " +
			"or for the current directory's file system when no FILE is given.",
		Examples: []command.Example{
			{Command: "df", Explain: "Show disk space usage for the current file system."},
			{Command: "df -h /", Explain: "Show usage for the root file system in human-readable form."},
			{Command: "df -i .", Explain: "Show inode usage instead of block usage."},
			{Command: "df --output=source,fstype,size,used,avail,pcent,target", Explain: "Select and order output columns."},
			{Command: "df --total -t ext4", Explain: "Show only ext4 filesystems and append a grand-total row."},
		},
		ExitStatus: "0  success.\n1  an error occurred (e.g. a file could not be stat'd).",
	})
	human := fs.BoolP("human-readable", "h", false, "print sizes in powers of 1024 (e.g., 1023M)")
	_ = fs.BoolP("kilobytes", "k", false, "use 1024-byte (1K) blocks (default)")
	inodes := fs.BoolP("inodes", "i", false, "list inode information instead of block usage")
	all := fs.BoolP("all", "a", false, "include pseudo, duplicate and inaccessible file systems")
	total := fs.Bool("total", false, "elide all entries insignificant to available space, and produce a grand total")
	types := fs.StringArrayP("type", "t", nil, "limit listing to file systems of type TYPE")
	blockSize := fs.String("block-size", "", "scale sizes by SIZE before printing them (e.g. 1M)")
	output := fs.String("output", "", "use the output format defined by FIELD_LIST, or print all fields if omitted")
	// --output may appear with no value (print default set). pflag's String
	// flag requires a value, so make it optional via NoOptDefVal.
	fs.Lookup("output").NoOptDefVal = defaultOutputAll

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	opts := options{human: *human, inodes: *inodes, all: *all, total: *total, types: *types}

	if strings.TrimSpace(*blockSize) != "" {
		bs, perr := parseSize(*blockSize)
		if perr != nil {
			_, _ = fmt.Fprintf(stdio.Err, "df: invalid --block-size argument '%s'\n", *blockSize)
			return command.SilentFailure()
		}
		opts.blockSize = bs
	}

	if strings.TrimSpace(*output) != "" {
		cols, cerr := parseOutput(*output)
		if cerr != nil {
			_, _ = fmt.Fprintf(stdio.Err, "df: %v\n", cerr)
			return command.SilentFailure()
		}
		opts.output = cols
	}

	operands := fs.Args()

	entries, firstErr := collectEntries(stdio.Err, operands, opts)
	entries = filterByType(entries, opts.types)

	if len(opts.output) > 0 {
		renderColumns(stdio.Out, entries, opts)
	} else {
		renderClassic(stdio.Out, entries, opts)
	}

	return firstErr
}

// collectEntries resolves the filesystems to report. With operands it reports
// the filesystem containing each operand (classic behavior). With no operands
// it lists every mounted filesystem from the mount table, hiding pseudo and
// zero-size filesystems unless --all is given.
func collectEntries(errw io.Writer, operands []string, opts options) ([]fsEntry, error) {
	if len(operands) > 0 {
		return collectFromOperands(errw, operands)
	}
	// With no operands, list the whole mount table only when a GNU option that
	// needs it is active; otherwise preserve the classic "current directory"
	// behavior so default `df` output is unchanged.
	if opts.all || opts.total || len(opts.types) > 0 || len(opts.output) > 0 {
		return collectFromMounts(errw, opts)
	}
	return collectFromOperands(errw, nil)
}

// collectFromOperands builds an entry for each FILE operand, preserving the
// historical default ("."), per-operand error reporting, and the classic
// behavior of labeling both the Filesystem and "Mounted on" columns with the
// operand path. The real device source and type (used by --output and --type)
// are resolved from the mount table when available, falling back to the f_type
// magic when not.
func collectFromOperands(errw io.Writer, operands []string) ([]fsEntry, error) {
	if len(operands) == 0 {
		operands = []string{"."}
	}
	mounts, _ := readMounts() // best-effort; source/fstype default if missing

	var entries []fsEntry
	var firstErr error
	for _, path := range operands {
		s, serr := statfs(path)
		if serr != nil {
			_, _ = fmt.Fprintf(errw, "df: %s\n", command.FileError(path, serr))
			firstErr = keep(firstErr)
			continue
		}
		src, fstype, _ := identify(path, mounts)
		if fstype == "" {
			fstype = fsTypeName(s.Type)
		}
		// Classic columns key off the operand path; source/fstype are carried
		// for --output and --type without disturbing the default layout.
		entries = append(entries, fsEntry{source: src, fstype: fstype, target: path, stat: s, operand: path})
	}
	return entries, firstErr
}

// collectFromMounts lists every mounted filesystem, statfs-ing each. Pseudo and
// zero-size filesystems are hidden unless --all is set, matching GNU df.
func collectFromMounts(errw io.Writer, opts options) ([]fsEntry, error) {
	mounts, err := readMounts()
	if err != nil {
		// Fall back to the current directory's filesystem.
		s, serr := statfs(".")
		if serr != nil {
			_, _ = fmt.Fprintf(errw, "df: %s\n", command.FileError(".", serr))
			return nil, keep(nil)
		}
		return []fsEntry{{source: "-", fstype: fsTypeName(s.Type), target: ".", stat: s}}, nil
	}

	var entries []fsEntry
	for _, m := range mounts {
		s, serr := statfs(m.target)
		if serr != nil {
			continue
		}
		if !opts.all && s.Blocks == 0 {
			continue // pseudo / zero-size filesystem
		}
		entries = append(entries, fsEntry{source: m.source, fstype: m.fstype, target: m.target, stat: s})
	}
	return entries, nil
}

// identify finds the mount entry whose target is the longest prefix of path,
// returning its source, fstype and target. Falls back to the path itself.
func identify(path string, mounts []mountEntry) (source, fstype, target string) {
	source, target = "-", path
	best := -1
	for _, m := range mounts {
		if m.target == path || strings.HasPrefix(path, strings.TrimRight(m.target, "/")+"/") || m.target == "/" {
			if len(m.target) > best {
				best = len(m.target)
				source, fstype, target = m.source, m.fstype, m.target
			}
		}
	}
	if best < 0 {
		return "-", "", path
	}
	return source, fstype, target
}

// filterByType keeps only entries whose fstype matches one of types. An empty
// types slice means no filtering.
func filterByType(entries []fsEntry, types []string) []fsEntry {
	if len(types) == 0 {
		return entries
	}
	want := make(map[string]struct{}, len(types))
	for _, t := range types {
		want[t] = struct{}{}
	}
	out := entries[:0]
	for _, e := range entries {
		if _, ok := want[e.fstype]; ok {
			out = append(out, e)
		}
	}
	return out
}

// renderClassic prints the historical fixed-column layout (block or inode
// mode), optionally followed by a grand-total row when --total is set.
func renderClassic(w io.Writer, entries []fsEntry, opts options) {
	writeHeader(w, opts)
	for _, e := range entries {
		_, _ = io.WriteString(w, formatRow(e, opts))
	}
	if opts.total {
		_, _ = io.WriteString(w, formatTotalRow(entries, opts))
	}
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
	} else if opts.blockSize > 0 {
		size = fmt.Sprintf("%d-blocks", opts.blockSize)
	}
	_, _ = fmt.Fprintf(w, "%-20s %10s %10s %10s %4s %s\n",
		"Filesystem", size, "Used", "Available", "Use%", "Mounted on")
}

// formatRow renders a single classic df row as a string.
func formatRow(e fsEntry, opts options) string {
	// In operand mode the classic layout labels the Filesystem column with the
	// operand path (historical behavior). In mount-table mode it uses the
	// device source.
	name := e.source
	if e.operand != "" {
		name = e.operand
	}
	if name == "" || name == "-" {
		name = e.target
	}
	if opts.inodes {
		u := computeInodeUsage(e.stat)
		return fmt.Sprintf("%-20s %10d %10d %10d %3d%% %s\n",
			name, u.files, u.used, u.free, u.usePct, e.target)
	}
	u := computeUsage(e.stat)
	return fmt.Sprintf("%-20s %10s %10s %10s %3d%% %s\n",
		name,
		scaleSize(u.total, opts.human, opts.blockSize),
		scaleSize(u.used, opts.human, opts.blockSize),
		scaleSize(u.avail, opts.human, opts.blockSize),
		u.usePct,
		e.target)
}

// formatTotalRow renders the grand-total row summing all entries.
func formatTotalRow(entries []fsEntry, opts options) string {
	if opts.inodes {
		var files, used, free uint64
		for _, e := range entries {
			u := computeInodeUsage(e.stat)
			files += u.files
			used += u.used
			free += u.free
		}
		return fmt.Sprintf("%-20s %10d %10d %10d %3d%% %s\n",
			"total", files, used, free, percent(used, files), "-")
	}
	var total, used, avail uint64
	for _, e := range entries {
		u := computeUsage(e.stat)
		total += u.total
		used += u.used
		avail += u.avail
	}
	return fmt.Sprintf("%-20s %10s %10s %10s %3d%% %s\n",
		"total",
		scaleSize(total, opts.human, opts.blockSize),
		scaleSize(used, opts.human, opts.blockSize),
		scaleSize(avail, opts.human, opts.blockSize),
		percent(used, used+avail),
		"-")
}

// keep preserves the first error so the exit code reflects a failure while the
// message has already been printed to stderr.
func keep(existing error) error {
	if existing != nil {
		return existing
	}
	return command.SilentFailure()
}
