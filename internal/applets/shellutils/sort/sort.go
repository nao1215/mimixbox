// Package sortcmd implements the sort applet: write the sorted concatenation of
// all FILE(s) (or standard input) to standard output, with the common GNU
// options. The package is named sortcmd rather than sort so it can import the
// standard library sort package.
package sortcmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// create opens name for writing, truncating any existing file. It is a thin
// wrapper kept separate so the I/O surface used by sort stays small.
func create(name string) (*os.File, error) {
	return os.Create(name) //nolint:gosec // operating on a user-named file is the whole point
}

// Command is the sort applet.
type Command struct{}

// New returns a sort command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "sort" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Sort lines of text files" }

// options holds the parsed flags that drive the comparison and post-processing.
type options struct {
	reverse        bool
	numeric        bool
	generalNumeric bool
	humanNumeric   bool
	versionSort    bool
	unique         bool
	ignoreCase     bool
	ignoreBlanks   bool
	stable         bool   // -s: disable the last-resort full-line comparison
	zeroTerminated bool   // -z: NUL line delimiter for input and output
	key            int    // 1-based start field for -k; 0 means whole line
	separator      string // -t separator; empty means runs of blanks
	hasSeparator   bool
}

// Run executes sort.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... [FILE]...", stdio.Err).WithHelp(command.Help{
		Description: "Write the sorted concatenation of all FILE(s) to standard output. " +
			"With no FILE, or when FILE is -, read standard input.",
		Examples: []command.Example{
			{Command: "sort file.txt", Explain: "Sort the lines of file.txt and print them."},
			{Command: "sort -u file.txt", Explain: "Sort lines and drop duplicates."},
			{Command: "sort -n -k 2 file.txt", Explain: "Sort numerically on the second field."},
		},
		ExitStatus: "0  success.\n1  the input was not sorted (with --check).\n2  an error occurred.",
	})
	reverse := fs.BoolP("reverse", "r", false, "reverse the result of comparisons")
	numeric := fs.BoolP("numeric-sort", "n", false, "compare according to string numerical value")
	generalNumeric := fs.BoolP("general-numeric-sort", "g", false, "compare according to general numerical value")
	humanNumeric := fs.BoolP("human-numeric-sort", "h", false, "compare human readable numbers (e.g., 2K 1G)")
	versionSort := fs.BoolP("version-sort", "V", false, "natural sort of (version) numbers within text")
	unique := fs.BoolP("unique", "u", false, "output only the first of an equal run")
	ignoreCase := fs.BoolP("ignore-case", "f", false, "fold lower case to upper case characters")
	ignoreBlanks := fs.BoolP("ignore-leading-blanks", "b", false, "ignore leading blanks")
	stable := fs.BoolP("stable", "s", false, "stabilize sort by disabling last-resort comparison")
	zeroTerminated := fs.BoolP("zero-terminated", "z", false, "line delimiter is NUL, not newline")
	merge := fs.BoolP("merge", "m", false, "merge already sorted files; do not sort")
	key := fs.StringP("key", "k", "", "sort via a key; KEYDEF gives location")
	separator := fs.StringP("field-separator", "t", "", "use SEP instead of non-blank to blank transition")
	check := fs.BoolP("check", "c", false, "check for sorted input; do not sort")
	output := fs.StringP("output", "o", "", "write result to FILE instead of standard output")
	// --parallel and --temporary-directory are accepted for GNU compatibility
	// but have no effect on this single-threaded, in-memory implementation.
	_ = fs.IntP("parallel", "", 0, "change the number of sorts run concurrently to N")
	_ = fs.StringP("temporary-directory", "T", "", "use DIR for temporaries, not $TMPDIR or /tmp")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	opts := options{
		reverse:        *reverse,
		numeric:        *numeric,
		generalNumeric: *generalNumeric,
		humanNumeric:   *humanNumeric,
		versionSort:    *versionSort,
		unique:         *unique,
		ignoreCase:     *ignoreCase,
		ignoreBlanks:   *ignoreBlanks,
		stable:         *stable,
		zeroTerminated: *zeroTerminated,
		separator:      *separator,
		hasSeparator:   *separator != "",
	}
	if *key != "" {
		k, perr := parseKey(*key)
		if perr != nil {
			return command.Failuref("%s", perr)
		}
		opts.key = k
	}

	lines, readErr := read(stdio, fs.Args(), opts.zeroTerminated)
	if readErr != nil {
		return readErr
	}

	if *check {
		return checkSorted(stdio, lines, opts)
	}

	// --merge assumes its inputs are already sorted; merging through the same
	// comparator yields the correct result, so sorting the concatenation is an
	// acceptable equivalent here.
	_ = *merge
	sorted := Sort(lines, opts)
	return writeLines(stdio, *output, sorted, opts.zeroTerminated)
}

// parseKey parses a -k KEYDEF of the simple form "N" (a single field number,
// meaning from field N to the end of the line) and returns the 1-based field.
func parseKey(def string) (int, error) {
	// Only the start field is honoured; a trailing ",M" is accepted but ignored.
	start := def
	if i := strings.IndexByte(def, ','); i >= 0 {
		start = def[:i]
	}
	// Strip any column offset / ordering flags such as "2.3" or "2n".
	for i, r := range start {
		if r < '0' || r > '9' {
			start = start[:i]
			break
		}
	}
	if start == "" {
		return 0, fmt.Errorf("invalid key definition: %q", def)
	}
	n, err := strconv.Atoi(start)
	if err != nil || n < 1 {
		return 0, fmt.Errorf("invalid key definition: %q", def)
	}
	return n, nil
}

// read reads every operand (defaulting to standard input when there are none)
// and returns the combined lines, splitting on '\n'. A trailing newline does not
// produce an empty final line.
func read(stdio command.IO, files []string, zeroTerminated bool) ([]string, error) {
	if len(files) == 0 {
		files = []string{"-"}
	}
	var b strings.Builder
	for _, name := range files {
		r, err := command.Open(stdio, name)
		if err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "sort: %s\n", command.FileError(name, err))
			return nil, command.SilentFailure()
		}
		_, err = io.Copy(&b, r)
		_ = r.Close()
		if err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "sort: %s\n", command.FileError(name, err))
			return nil, command.SilentFailure()
		}
	}
	return splitLines(b.String(), zeroTerminated), nil
}

// splitLines splits text into records on the delimiter ('\n' by default, or NUL
// when zeroTerminated), dropping a single trailing delimiter so that "a\nb\n"
// yields two lines rather than three.
func splitLines(text string, zeroTerminated bool) []string {
	if text == "" {
		return nil
	}
	delim := "\n"
	if zeroTerminated {
		delim = "\x00"
	}
	text = strings.TrimSuffix(text, delim)
	return strings.Split(text, delim)
}

// Sort returns the lines sorted according to opts. It is pure: it does not touch
// any I/O and returns a new slice, which makes it directly unit-testable.
func Sort(lines []string, opts options) []string {
	out := make([]string, len(lines))
	copy(out, lines)

	sort.SliceStable(out, func(i, j int) bool {
		c := compare(out[i], out[j], opts)
		// GNU sort applies a last-resort full-line comparison when the keys are
		// equal, unless --stable (-s) is given. With --stable the original
		// input order is preserved for equal keys (SliceStable handles that).
		if c == 0 && !opts.stable {
			c = strings.Compare(out[i], out[j])
		}
		if opts.reverse {
			return c > 0
		}
		return c < 0
	})

	if opts.unique {
		out = uniq(out, opts)
	}
	return out
}

// compare returns a negative, zero, or positive value reporting whether the key
// of a sorts before, equal to, or after the key of b.
func compare(a, b string, opts options) int {
	ka := key(a, opts)
	kb := key(b, opts)

	switch {
	case opts.versionSort:
		return versionCompare(ka, kb)
	case opts.humanNumeric:
		return numericCompare(humanNumber(ka), humanNumber(kb))
	case opts.generalNumeric:
		return numericCompare(generalNumber(ka), generalNumber(kb))
	case opts.numeric:
		return numericCompare(leadingNumber(ka), leadingNumber(kb))
	}
	return strings.Compare(ka, kb)
}

// numericCompare returns the three-way comparison of two floating-point keys.
func numericCompare(a, b float64) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}

// key extracts the comparison key from a line: the selected field (when -k is
// set) with the case- and blank-folding transformations applied.
func key(line string, opts options) string {
	s := line
	if opts.key > 0 {
		s = field(line, opts)
	}
	if opts.ignoreBlanks {
		s = strings.TrimLeft(s, " \t")
	}
	if opts.ignoreCase {
		s = strings.ToUpper(s)
	}
	return s
}

// field returns the substring of line from the start field selected by -k to the
// end of the line, using the -t separator (or runs of blanks by default).
func field(line string, opts options) string {
	var fields []string
	if opts.hasSeparator {
		fields = strings.Split(line, opts.separator)
	} else {
		fields = strings.FieldsFunc(line, func(r rune) bool {
			return r == ' ' || r == '\t'
		})
	}
	idx := opts.key - 1
	if idx >= len(fields) {
		return ""
	}
	if opts.hasSeparator {
		return strings.Join(fields[idx:], opts.separator)
	}
	return strings.Join(fields[idx:], " ")
}

// leadingNumber parses the leading numeric value of s, ignoring leading blanks.
// A string with no number parses as 0, matching GNU sort -n.
func leadingNumber(s string) float64 {
	s = strings.TrimLeft(s, " \t")
	end := 0
	for end < len(s) {
		ch := s[end]
		if (ch >= '0' && ch <= '9') || ch == '+' || ch == '-' || ch == '.' {
			end++
			continue
		}
		break
	}
	if end == 0 {
		return 0
	}
	n, err := strconv.ParseFloat(s[:end], 64)
	if err != nil {
		return 0
	}
	return n
}

// generalNumber parses the leading floating-point value of s for -g, accepting
// any prefix that strconv.ParseFloat understands (including exponents). A string
// with no number parses as 0, matching GNU sort -g.
func generalNumber(s string) float64 {
	s = strings.TrimLeft(s, " \t")
	end := 0
	for end < len(s) {
		ch := s[end]
		if (ch >= '0' && ch <= '9') || ch == '+' || ch == '-' || ch == '.' || ch == 'e' || ch == 'E' {
			end++
			continue
		}
		break
	}
	for end > 0 {
		if n, err := strconv.ParseFloat(s[:end], 64); err == nil {
			return n
		}
		end--
	}
	return 0
}

// humanNumber parses a leading number with an optional SI/IEC suffix for -h,
// returning its magnitude in bytes so that "2K" < "1M" < "1G". A string with no
// number parses as 0.
func humanNumber(s string) float64 {
	s = strings.TrimLeft(s, " \t")
	end := 0
	for end < len(s) {
		ch := s[end]
		if (ch >= '0' && ch <= '9') || ch == '+' || ch == '-' || ch == '.' {
			end++
			continue
		}
		break
	}
	if end == 0 {
		return 0
	}
	n, err := strconv.ParseFloat(s[:end], 64)
	if err != nil {
		return 0
	}
	if end >= len(s) {
		return n
	}
	switch s[end] {
	case 'K', 'k':
		return n * 1024
	case 'M', 'm':
		return n * 1024 * 1024
	case 'G', 'g':
		return n * 1024 * 1024 * 1024
	case 'T', 't':
		return n * 1024 * 1024 * 1024 * 1024
	case 'P', 'p':
		return n * 1024 * 1024 * 1024 * 1024 * 1024
	case 'E', 'e':
		return n * 1024 * 1024 * 1024 * 1024 * 1024 * 1024
	default:
		return n
	}
}

// versionCompare orders two strings the way "sort -V" does: it walks both
// strings together, comparing runs of digits as integers (so 2 < 10) and runs
// of non-digits lexically (so "1.2" < "1.10").
func versionCompare(a, b string) int {
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		da := a[i] >= '0' && a[i] <= '9'
		db := b[j] >= '0' && b[j] <= '9'
		if da && db {
			// Compare two numeric runs by value, skipping leading zeros.
			si, sj := i, j
			for i < len(a) && a[i] >= '0' && a[i] <= '9' {
				i++
			}
			for j < len(b) && b[j] >= '0' && b[j] <= '9' {
				j++
			}
			na := strings.TrimLeft(a[si:i], "0")
			nb := strings.TrimLeft(b[sj:j], "0")
			if len(na) != len(nb) {
				if len(na) < len(nb) {
					return -1
				}
				return 1
			}
			if c := strings.Compare(na, nb); c != 0 {
				return c
			}
			continue
		}
		if a[i] != b[j] {
			if a[i] < b[j] {
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

// uniq drops lines whose comparison key equals that of the preceding line. The
// slice is assumed to already be sorted, so equal keys are adjacent.
func uniq(lines []string, opts options) []string {
	if len(lines) == 0 {
		return lines
	}
	out := lines[:1]
	for _, l := range lines[1:] {
		if compare(out[len(out)-1], l, opts) != 0 {
			out = append(out, l)
		}
	}
	return out
}

// checkSorted verifies the input is already sorted, returning an error (and
// reporting the first disorder on stderr) when it is not.
func checkSorted(stdio command.IO, lines []string, opts options) error {
	for i := 1; i < len(lines); i++ {
		c := compare(lines[i-1], lines[i], opts)
		disordered := c > 0
		if opts.reverse {
			disordered = c < 0
		}
		if disordered {
			_, _ = fmt.Fprintf(stdio.Err, "sort: -:%d: disorder: %s\n", i+1, lines[i])
			return command.Failure(nil)
		}
	}
	return nil
}

// writeLines writes the sorted lines (each terminated by '\n') to the -o file,
// or to standard output when output is empty.
func writeLines(stdio command.IO, output string, lines []string, zeroTerminated bool) error {
	w := stdio.Out
	if output != "" {
		f, err := create(output)
		if err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "sort: %s\n", command.FileError(output, err))
			return command.SilentFailure()
		}
		defer func() { _ = f.Close() }()
		w = f
	}
	delim := byte('\n')
	if zeroTerminated {
		delim = 0
	}
	var b strings.Builder
	for _, l := range lines {
		b.WriteString(l)
		b.WriteByte(delim)
	}
	if _, err := io.WriteString(w, b.String()); err != nil {
		return command.Failure(err)
	}
	return nil
}
