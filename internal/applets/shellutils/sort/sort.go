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
	reverse      bool
	numeric      bool
	unique       bool
	ignoreCase   bool
	ignoreBlanks bool
	key          int    // 1-based start field for -k; 0 means whole line
	separator    string // -t separator; empty means runs of blanks
	hasSeparator bool
}

// Run executes sort.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... [FILE]...", stdio.Err)
	reverse := fs.BoolP("reverse", "r", false, "reverse the result of comparisons")
	numeric := fs.BoolP("numeric-sort", "n", false, "compare according to string numerical value")
	unique := fs.BoolP("unique", "u", false, "output only the first of an equal run")
	ignoreCase := fs.BoolP("ignore-case", "f", false, "fold lower case to upper case characters")
	ignoreBlanks := fs.BoolP("ignore-leading-blanks", "b", false, "ignore leading blanks")
	key := fs.StringP("key", "k", "", "sort via a key; KEYDEF gives location")
	separator := fs.StringP("field-separator", "t", "", "use SEP instead of non-blank to blank transition")
	check := fs.BoolP("check", "c", false, "check for sorted input; do not sort")
	output := fs.StringP("output", "o", "", "write result to FILE instead of standard output")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	opts := options{
		reverse:      *reverse,
		numeric:      *numeric,
		unique:       *unique,
		ignoreCase:   *ignoreCase,
		ignoreBlanks: *ignoreBlanks,
		separator:    *separator,
		hasSeparator: *separator != "",
	}
	if *key != "" {
		k, perr := parseKey(*key)
		if perr != nil {
			return command.Failuref("%s", perr)
		}
		opts.key = k
	}

	lines, readErr := read(stdio, fs.Args())
	if readErr != nil {
		return readErr
	}

	if *check {
		return checkSorted(stdio, lines, opts)
	}

	sorted := Sort(lines, opts)
	return writeLines(stdio, *output, sorted)
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
func read(stdio command.IO, files []string) ([]string, error) {
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
	return splitLines(b.String()), nil
}

// splitLines splits text into lines on '\n', dropping a single trailing newline
// so that "a\nb\n" yields two lines rather than three.
func splitLines(text string) []string {
	if text == "" {
		return nil
	}
	text = strings.TrimSuffix(text, "\n")
	return strings.Split(text, "\n")
}

// Sort returns the lines sorted according to opts. It is pure: it does not touch
// any I/O and returns a new slice, which makes it directly unit-testable.
func Sort(lines []string, opts options) []string {
	out := make([]string, len(lines))
	copy(out, lines)

	sort.SliceStable(out, func(i, j int) bool {
		c := compare(out[i], out[j], opts)
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

	if opts.numeric {
		na := leadingNumber(ka)
		nb := leadingNumber(kb)
		switch {
		case na < nb:
			return -1
		case na > nb:
			return 1
		default:
			return 0
		}
	}
	return strings.Compare(ka, kb)
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
func writeLines(stdio command.IO, output string, lines []string) error {
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
	var b strings.Builder
	for _, l := range lines {
		b.WriteString(l)
		b.WriteByte('\n')
	}
	if _, err := io.WriteString(w, b.String()); err != nil {
		return command.Failure(err)
	}
	return nil
}
