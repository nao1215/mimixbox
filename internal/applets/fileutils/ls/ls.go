// Package ls implements the ls applet: list directory contents. It covers the
// everyday desktop subset - plain, -1, -a, -A, -d, -l, -F, -h, -R - with stable
// name sorting and deterministic error reporting, plus the GNU presentation
// flags --color, --file-type/--indicator-style, --sort/--time/--group-
// directories-first, --hide/--ignore, and --inode/--block-size/--kibibytes.
package ls

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
	"golang.org/x/term"
)

// Command is the ls applet.
type Command struct{}

// New returns an ls command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "ls" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "List directory contents" }

// indicatorStyle selects which trailing suffix is appended to entry names.
type indicatorStyle int

const (
	indicatorNone     indicatorStyle = iota // no suffix
	indicatorSlash                          // directories only, "/"
	indicatorFileType                       // like classify but no "*" for executables
	indicatorClassify                       // full set: / * @ | =
)

// sortKey selects the comparison used to order entries.
type sortKey int

const (
	sortName      sortKey = iota // default: lexicographic name
	sortNone                     // -U / --sort=none: directory order
	sortSize                     // --sort=size
	sortTime                     // --sort=time
	sortVersion                  // --sort=version
	sortExtension                // --sort=extension
)

// timeField selects which timestamp --sort=time (and -l) considers.
type timeField int

const (
	timeMtime timeField = iota
	timeAtime
	timeCtime
)

// GNU-style default ANSI colors (LS_COLORS-like minimal set).
const (
	colorDir     = "\x1b[01;34m" // bold blue
	colorExec    = "\x1b[01;32m" // bold green
	colorSymlink = "\x1b[01;36m" // bold cyan
	colorReset   = "\x1b[0m"
)

type options struct {
	all       bool // -a: include . and ..
	almostAll bool // -A: include dotfiles but not . and ..
	long      bool // -l
	dirSelf   bool // -d
	human     bool // -h
	recursive bool // -R
	inode     bool // -i: print inode number

	indicator indicatorStyle // -F / --file-type / --indicator-style
	color     bool           // resolved color decision (after auto)

	sortBy    sortKey
	timeBy    timeField
	groupDirs bool // --group-directories-first

	hide   string // --hide=PATTERN
	ignore string // --ignore=PATTERN

	blockSize int64 // bytes per block for -l sizes (0 == raw bytes)
}

// Run executes ls.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... [FILE]...", stdio.Err).WithHelp(command.Help{
		Description: "List information about FILEs (the current directory by default), sorted by name.",
		Examples: []command.Example{
			{Command: "ls -la", Explain: "Long format, including dotfiles."},
			{Command: "ls -R dir", Explain: "List dir and its subdirectories recursively."},
			{Command: "ls --color=always --sort=size", Explain: "Colorized output, largest first."},
		},
		ExitStatus: "0  success.\n2  a FILE could not be accessed.",
	})
	all := fs.BoolP("all", "a", false, "do not ignore entries starting with .")
	almost := fs.BoolP("almost-all", "A", false, "like -a but omit . and ..")
	long := fs.BoolP("long", "l", false, "use a long listing format")
	dirSelf := fs.BoolP("directory", "d", false, "list directories themselves, not their contents")
	classify := fs.BoolP("classify", "F", false, "append an indicator (one of */=@|) to entries")
	human := fs.BoolP("human-readable", "h", false, "with -l, print sizes like 1K 234M")
	recursive := fs.BoolP("recursive", "R", false, "list subdirectories recursively")
	inode := fs.BoolP("inode", "i", false, "print the index number of each file")
	_ = fs.BoolP("one-per-line", "1", false, "list one file per line (the default for non-terminals)")

	// #722 color.
	color := fs.String("color", "never", "colorize the output; WHEN is 'always', 'auto', or 'never'")
	fs.Lookup("color").NoOptDefVal = "always" // bare --color == --color=always

	// #723 indicator styles.
	fileType := fs.Bool("file-type", false, "append indicators (/=@|) but not '*' for executables")
	indicatorStyleFlag := fs.String("indicator-style", "", "append indicator STYLE: none, slash, file-type, classify")

	// #724 sorting.
	sortFlag := fs.String("sort", "name", "sort by WORD: none, size, time, version, extension")
	timeFlag := fs.String("time", "mtime", "with --sort=time, use atime, mtime, or ctime")
	groupDirs := fs.Bool("group-directories-first", false, "list directories before files")
	noneSort := fs.BoolP("none-sort", "U", false, "do not sort; list entries in directory order")

	// #725 hide/ignore.
	hide := fs.String("hide", "", "do not list entries matching shell PATTERN (overridden by -a/-A)")
	ignore := fs.String("ignore", "", "do not list entries matching shell PATTERN")

	// #726 inode / block size.
	blockSize := fs.String("block-size", "", "with -l, scale sizes by SIZE (e.g. K, M, 1024)")
	kibibytes := fs.BoolP("kibibytes", "k", false, "with -l, use 1024-byte blocks for sizes")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	opts := options{
		all:       *all,
		almostAll: *almost,
		long:      *long,
		dirSelf:   *dirSelf,
		human:     *human,
		recursive: *recursive,
		inode:     *inode,
		groupDirs: *groupDirs,
		hide:      *hide,
		ignore:    *ignore,
	}

	opts.indicator = resolveIndicator(*classify, *fileType, *indicatorStyleFlag)

	cm, err := parseColorMode(*color)
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "ls: %s\n", err)
		return &command.ExitError{Code: 2}
	}
	opts.color = resolveColor(cm, stdio.Out)

	opts.sortBy, err = parseSort(*sortFlag, *noneSort)
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "ls: %s\n", err)
		return &command.ExitError{Code: 2}
	}
	opts.timeBy, err = parseTime(*timeFlag)
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "ls: %s\n", err)
		return &command.ExitError{Code: 2}
	}

	opts.blockSize, err = resolveBlockSize(*blockSize, *kibibytes)
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "ls: %s\n", err)
		return &command.ExitError{Code: 2}
	}

	operands := fs.Args()
	if len(operands) == 0 {
		operands = []string{"."}
	}

	var files []entry
	var dirs []string
	exitErr := false
	for _, name := range operands {
		info, err := os.Lstat(name)
		if err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "ls: cannot access '%s': %s\n", name, errMessage(err))
			exitErr = true
			continue
		}
		if info.IsDir() && !opts.dirSelf {
			dirs = append(dirs, name)
		} else {
			// Reuse the Lstat result so the listing pipeline never restats this
			// command-line operand.
			files = append(files, entry{name: name, dir: "", info: info})
		}
	}

	if len(files) > 0 {
		sortEntries(files, opts)
		c.listEntries(stdio.Out, files, opts)
		if len(dirs) > 0 {
			_, _ = fmt.Fprintln(stdio.Out)
		}
	}

	header := len(operands) > 1 || opts.recursive
	for i, dir := range dirs {
		if header {
			if i > 0 || len(files) > 0 {
				_, _ = fmt.Fprintln(stdio.Out)
			}
			_, _ = fmt.Fprintf(stdio.Out, "%s:\n", dir)
		}
		if err := c.listDir(stdio.Out, stdio.Err, dir, opts); err != nil {
			exitErr = true
		}
	}

	if exitErr {
		return &command.ExitError{Code: 2}
	}
	return nil
}

// resolveIndicator collapses -F, --file-type, and --indicator-style into one
// indicator style. An explicit --indicator-style wins; otherwise --file-type
// implies file-type and -F implies classify.
func resolveIndicator(classify, fileType bool, style string) indicatorStyle {
	switch strings.ToLower(strings.TrimSpace(style)) {
	case "none":
		return indicatorNone
	case "slash":
		return indicatorSlash
	case "file-type":
		return indicatorFileType
	case "classify":
		return indicatorClassify
	}
	if fileType {
		return indicatorFileType
	}
	if classify {
		return indicatorClassify
	}
	return indicatorNone
}

// parseColorMode validates the --color WHEN argument.
func parseColorMode(when string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(when)) {
	case "", "always", "yes", "force":
		if when == "" {
			return "never", nil
		}
		return "always", nil
	case "auto", "tty", "if-tty":
		return "auto", nil
	case "never", "no", "none":
		return "never", nil
	default:
		return "", fmt.Errorf("invalid argument '%s' for '--color'", when)
	}
}

// resolveColor turns a color mode into a concrete on/off decision. "auto"
// colors only when out is a terminal.
func resolveColor(mode string, out io.Writer) bool {
	switch mode {
	case "always":
		return true
	case "auto":
		if f, ok := out.(*os.File); ok {
			return term.IsTerminal(int(f.Fd()))
		}
		return false
	default:
		return false
	}
}

// parseSort maps --sort/-U to a sort key.
func parseSort(word string, none bool) (sortKey, error) {
	if none {
		return sortNone, nil
	}
	switch strings.ToLower(strings.TrimSpace(word)) {
	case "", "name":
		return sortName, nil
	case "none":
		return sortNone, nil
	case "size":
		return sortSize, nil
	case "time":
		return sortTime, nil
	case "version":
		return sortVersion, nil
	case "extension":
		return sortExtension, nil
	default:
		return sortName, fmt.Errorf("invalid argument '%s' for '--sort'", word)
	}
}

// parseTime maps --time to a timestamp field.
func parseTime(word string) (timeField, error) {
	switch strings.ToLower(strings.TrimSpace(word)) {
	case "", "mtime", "modification", "mod":
		return timeMtime, nil
	case "atime", "access", "use":
		return timeAtime, nil
	case "ctime", "status":
		return timeCtime, nil
	default:
		return timeMtime, fmt.Errorf("invalid argument '%s' for '--time'", word)
	}
}

// resolveBlockSize converts -k / --block-size into a byte count per block. A
// zero result means "report raw bytes" (the default).
func resolveBlockSize(spec string, kibibytes bool) (int64, error) {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		if kibibytes {
			return 1024, nil
		}
		return 0, nil
	}
	n, err := parseSize(spec)
	if err != nil {
		return 0, fmt.Errorf("invalid --block-size argument '%s'", spec)
	}
	return n, nil
}

// parseSize parses a size like "1024", "K", "1M", "512" into a byte count.
// A bare suffix (e.g. "K") means one unit (1024 bytes).
func parseSize(spec string) (int64, error) {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return 0, fmt.Errorf("empty size")
	}
	// Split leading digits from a trailing unit suffix.
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

// listDir lists one directory's entries, recursing when -R is set.
func (c *Command) listDir(out, errw io.Writer, dir string, opts options) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		_, _ = fmt.Fprintf(errw, "ls: cannot open directory '%s': %s\n", dir, errMessage(err))
		return err
	}

	names := make([]string, 0, len(entries)+2)
	if opts.all {
		names = append(names, ".", "..")
	}
	for _, e := range entries {
		name := e.Name()
		if !opts.all && !opts.almostAll && strings.HasPrefix(name, ".") {
			continue
		}
		if filtered(name, opts) {
			continue
		}
		names = append(names, name)
	}

	// Populate each entry's metadata once; sorting, decoration, long-format
	// rendering, and recursion all consume this cached info instead of
	// restatting the path.
	items := make([]entry, 0, len(names))
	for _, n := range names {
		items = append(items, newEntry(dir, n))
	}

	sortEntries(items, opts)
	c.listEntries(out, items, opts)

	if opts.recursive {
		var subdirs []string
		for _, e := range items {
			if e.name == "." || e.name == ".." {
				continue
			}
			if e.isDir() {
				subdirs = append(subdirs, e.path())
			}
		}
		for _, sd := range subdirs {
			_, _ = fmt.Fprintln(out)
			_, _ = fmt.Fprintf(out, "%s:\n", sd)
			if err := c.listDir(out, errw, sd, opts); err != nil {
				return err
			}
		}
	}
	return nil
}

// filtered reports whether name should be omitted by --ignore/--hide. --ignore
// always applies; --hide is suppressed when -a/-A is given (GNU precedence).
func filtered(name string, opts options) bool {
	if opts.ignore != "" {
		if ok, _ := path.Match(opts.ignore, name); ok {
			return true
		}
	}
	if opts.hide != "" && !opts.all && !opts.almostAll {
		if ok, _ := path.Match(opts.hide, name); ok {
			return true
		}
	}
	return false
}

// entry is a listing item paired with its already-resolved metadata. Building
// it once per operand or directory entry lets sorting, decoration, long-format
// rendering, and recursive traversal share a single os.Lstat instead of
// restatting the same path through every helper.
type entry struct {
	name string      // display name (may be "." or "..")
	dir  string      // directory the name lives in ("" for command-line operands)
	info os.FileInfo // from Lstat; nil when the path could not be stated
}

// newEntry resolves name (within dir) to an entry, caching its Lstat result. A
// stat failure yields an entry with a nil info, which the consumers treat the
// same way the old per-helper Lstat-error branches did.
func newEntry(dir, name string) entry {
	info, err := os.Lstat(pathOf(dir, name))
	if err != nil {
		return entry{name: name, dir: dir}
	}
	return entry{name: name, dir: dir, info: info}
}

// path returns the filesystem path the entry refers to.
func (e entry) path() string { return pathOf(e.dir, e.name) }

// isDir reports whether the entry is a directory (false when unstated).
func (e entry) isDir() bool { return e.info != nil && e.info.IsDir() }

// size returns the entry's byte size, or 0 when unstated.
func (e entry) size() int64 {
	if e.info == nil {
		return 0
	}
	return e.info.Size()
}

// inode returns the entry's inode number, or 0 when unavailable.
func (e entry) inode() uint64 {
	if e.info == nil {
		return 0
	}
	if st, ok := e.info.Sys().(*syscall.Stat_t); ok {
		return st.Ino
	}
	return 0
}

// modTime returns the entry's selected timestamp, or the zero time when
// unstated.
func (e entry) modTime(field timeField) time.Time {
	if e.info == nil {
		return time.Time{}
	}
	return timeFromInfo(e.info, field)
}

// sortEntries orders entries in place according to the active sort key.
// --group-directories-first hoists directories ahead of files within the chosen
// order. The comparison logic mirrors the previous name-based sort exactly so
// output stays byte-for-byte stable.
func sortEntries(entries []entry, opts options) {
	if opts.sortBy == sortNone && !opts.groupDirs {
		return
	}

	less := func(i, j int) bool {
		a, b := entries[i], entries[j]
		if opts.groupDirs {
			ad, bd := a.isDir(), b.isDir()
			if ad != bd {
				return ad
			}
		}
		switch opts.sortBy {
		case sortNone:
			return false // keep directory order within the group
		case sortSize:
			if a.size() != b.size() {
				return a.size() > b.size() // larger first, like GNU
			}
			return a.name < b.name
		case sortTime:
			ta, tb := a.modTime(opts.timeBy), b.modTime(opts.timeBy)
			if !ta.Equal(tb) {
				return ta.After(tb) // newest first
			}
			return a.name < b.name
		case sortVersion:
			if cmp := versionCompare(a.name, b.name); cmp != 0 {
				return cmp < 0
			}
			return a.name < b.name
		case sortExtension:
			ea, eb := filepath.Ext(a.name), filepath.Ext(b.name)
			if ea != eb {
				return ea < eb
			}
			return a.name < b.name
		default:
			return a.name < b.name
		}
	}

	sort.SliceStable(entries, less)
}

// listEntries prints the given entries in the selected format.
func (c *Command) listEntries(out io.Writer, entries []entry, opts options) {
	if !opts.long {
		for _, e := range entries {
			prefix := ""
			if opts.inode {
				prefix = fmt.Sprintf("%d ", e.inode())
			}
			_, _ = fmt.Fprintln(out, prefix+c.decorate(e, opts))
		}
		return
	}
	for _, e := range entries {
		_, _ = fmt.Fprintln(out, c.longLine(e, opts))
	}
}

// decorate returns the display name with color (if enabled) and indicator (if
// requested) applied.
func (c *Command) decorate(e entry, opts options) string {
	return c.colorize(e, opts) + c.indicator(e, opts)
}

// colorize wraps the entry's name in ANSI escapes based on its type, when color
// is on.
func (c *Command) colorize(e entry, opts options) string {
	if !opts.color || e.info == nil {
		return e.name
	}
	var col string
	switch {
	case e.info.IsDir():
		col = colorDir
	case e.info.Mode()&os.ModeSymlink != 0:
		col = colorSymlink
	case e.info.Mode()&0o111 != 0 && e.info.Mode().IsRegular():
		col = colorExec
	default:
		return e.name
	}
	return col + e.name + colorReset
}

// indicator returns the type suffix for the entry per the active indicator
// style.
func (c *Command) indicator(e entry, opts options) string {
	if opts.indicator == indicatorNone || e.info == nil {
		return ""
	}
	if e.info.IsDir() {
		return "/"
	}
	if opts.indicator == indicatorSlash {
		return ""
	}
	switch {
	case e.info.Mode()&os.ModeSymlink != 0:
		return "@"
	case e.info.Mode()&os.ModeNamedPipe != 0:
		return "|"
	case e.info.Mode()&os.ModeSocket != 0:
		return "="
	case e.info.Mode()&0o111 != 0 && e.info.Mode().IsRegular():
		if opts.indicator == indicatorFileType {
			return "" // file-type omits the executable marker
		}
		return "*"
	default:
		return ""
	}
}

// longLine formats one entry for -l.
func (c *Command) longLine(e entry, opts options) string {
	if e.info == nil {
		return e.name
	}
	info := e.info
	mode := modeString(info)
	nlink := uint64(1)
	owner, group := "?", "?"
	if st, ok := info.Sys().(*syscall.Stat_t); ok {
		nlink = uint64(st.Nlink)
		owner = lookupUser(st.Uid)
		group = lookupGroup(st.Gid)
	}
	size := sizeString(info.Size(), opts.human, opts.blockSize)
	when := timeString(timeFromInfo(info, opts.timeBy))
	display := c.colorize(e, opts) + c.indicator(e, opts)
	if info.Mode()&os.ModeSymlink != 0 {
		if target, err := os.Readlink(e.path()); err == nil {
			display = c.colorize(e, opts) + " -> " + target
		}
	}
	prefix := ""
	if opts.inode {
		prefix = fmt.Sprintf("%d ", e.inode())
	}
	return fmt.Sprintf("%s%s %d %s %s %s %s %s", prefix, mode, nlink, owner, group, size, when, display)
}

// timeFromInfo extracts the requested timestamp from a FileInfo.
func timeFromInfo(info os.FileInfo, field timeField) time.Time {
	if st, ok := info.Sys().(*syscall.Stat_t); ok {
		switch field {
		case timeAtime:
			return time.Unix(st.Atim.Sec, st.Atim.Nsec)
		case timeCtime:
			return time.Unix(st.Ctim.Sec, st.Ctim.Nsec)
		}
	}
	return info.ModTime()
}

// versionCompare compares two strings "naturally", so file2 sorts before
// file10. It returns -1, 0, or 1.
func versionCompare(a, b string) int {
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		ai, bi := a[i], b[j]
		if isDigit(ai) && isDigit(bi) {
			// Compare runs of digits numerically.
			si, ei := i, i
			for ei < len(a) && isDigit(a[ei]) {
				ei++
			}
			sj, ej := j, j
			for ej < len(b) && isDigit(b[ej]) {
				ej++
			}
			na := strings.TrimLeft(a[si:ei], "0")
			nb := strings.TrimLeft(b[sj:ej], "0")
			if len(na) != len(nb) {
				if len(na) < len(nb) {
					return -1
				}
				return 1
			}
			if na != nb {
				if na < nb {
					return -1
				}
				return 1
			}
			i, j = ei, ej
			continue
		}
		if ai != bi {
			if ai < bi {
				return -1
			}
			return 1
		}
		i++
		j++
	}
	switch {
	case i < len(a):
		return 1
	case j < len(b):
		return -1
	default:
		return 0
	}
}

func isDigit(b byte) bool { return b >= '0' && b <= '9' }

func pathOf(dir, name string) string {
	if dir == "" || name == "." || name == ".." {
		if dir == "" {
			return name
		}
	}
	return filepath.Join(dir, name)
}

// modeString renders the rwx-style permission string for ls -l.
func modeString(info os.FileInfo) string {
	m := info.Mode()
	var b strings.Builder
	switch {
	case m&os.ModeDir != 0:
		b.WriteByte('d')
	case m&os.ModeSymlink != 0:
		b.WriteByte('l')
	case m&os.ModeDevice != 0 && m&os.ModeCharDevice != 0:
		b.WriteByte('c')
	case m&os.ModeDevice != 0:
		b.WriteByte('b')
	case m&os.ModeNamedPipe != 0:
		b.WriteByte('p')
	case m&os.ModeSocket != 0:
		b.WriteByte('s')
	default:
		b.WriteByte('-')
	}
	const rwx = "rwxrwxrwx"
	perm := m.Perm()
	for i := 0; i < 9; i++ {
		if perm&(1<<uint(8-i)) != 0 {
			b.WriteByte(rwx[i])
		} else {
			b.WriteByte('-')
		}
	}
	return b.String()
}

// sizeString renders a size for ls -l. When blockSize > 0 the size is scaled to
// that many bytes per block (rounding up), matching GNU --block-size/-k.
func sizeString(size int64, human bool, blockSize int64) string {
	if blockSize > 0 {
		blocks := size / blockSize
		if size%blockSize != 0 {
			blocks++
		}
		return strconv.FormatInt(blocks, 10)
	}
	if !human {
		return strconv.FormatInt(size, 10)
	}
	const unit = 1024
	if size < unit {
		return strconv.FormatInt(size, 10)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	value := float64(size) / float64(div)
	suffix := "KMGTPE"[exp]
	if value < 10 {
		return fmt.Sprintf("%.1f%c", value, suffix)
	}
	return fmt.Sprintf("%.0f%c", value, suffix)
}

// timeString formats a modification time the way ls does (recent vs old).
func timeString(t time.Time) string {
	now := time.Now()
	sixMonths := time.Hour * 24 * 182
	if now.Sub(t) > sixMonths || t.Sub(now) > sixMonths {
		return t.Format("Jan _2  2006")
	}
	return t.Format("Jan _2 15:04")
}

var userCache = map[uint32]string{}
var groupCache = map[uint32]string{}

func lookupUser(uid uint32) string {
	if name, ok := userCache[uid]; ok {
		return name
	}
	name := strconv.FormatUint(uint64(uid), 10)
	if u, err := user.LookupId(name); err == nil {
		name = u.Username
	}
	userCache[uid] = name
	return name
}

func lookupGroup(gid uint32) string {
	if name, ok := groupCache[gid]; ok {
		return name
	}
	name := strconv.FormatUint(uint64(gid), 10)
	if g, err := user.LookupGroupId(name); err == nil {
		name = g.Name
	}
	groupCache[gid] = name
	return name
}

// errMessage returns the human-friendly tail of a path error.
func errMessage(err error) string {
	if os.IsNotExist(err) {
		return "No such file or directory"
	}
	if os.IsPermission(err) {
		return "Permission denied"
	}
	return err.Error()
}
