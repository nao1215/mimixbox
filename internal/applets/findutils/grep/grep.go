// Package grep implements the grep applet: search files (or standard input) for
// lines matching a pattern and print them. The pattern is a Go (RE2) regular
// expression, which covers the common extended-regexp use; -F switches to plain
// fixed-string matching.
package grep

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"io/fs"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
	"golang.org/x/term"
)

// Command is the grep applet.
type Command struct{}

// New returns a grep command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "grep" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Print lines that match patterns" }

// colorMatch is the GNU default highlight for matched text (bold red).
const (
	colorStart = "\x1b[01;31m"
	colorReset = "\x1b[0m"
)

// options holds the parsed switches.
type options struct {
	ignoreCase bool
	invert     bool
	lineNum    bool
	count      bool
	recursive  bool
	filesMatch bool
	filesNoMat bool
	fixed      bool
	word       bool
	quiet      bool
	withName   bool
	noName     bool
	byteOffset bool

	after  int // -A: trailing context lines
	before int // -B: leading context lines

	color bool // emit color escapes around matches

	include    string // --include glob (only in recursive mode)
	exclude    string // --exclude glob
	excludeDir string // --exclude-dir glob
}

// isTerminal reports whether w is a terminal; tests can replace it.
var isTerminal = func(w io.Writer) bool {
	f, ok := w.(interface{ Fd() uintptr })
	return ok && term.IsTerminal(int(f.Fd()))
}

// Run executes grep.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... PATTERN [FILE]...", stdio.Err).WithHelp(command.Help{
		Description: "Search each FILE (or standard input when no FILE is given) for lines matching PATTERN and print " +
			"them. PATTERN is a Go (RE2) regular expression by default; -F treats it as a fixed string.",
		Examples: []command.Example{
			{Command: "grep -rn TODO .", Explain: "Recursively search the current directory and print matching lines with line numbers."},
			{Command: "grep -i error log.txt", Explain: "Print lines containing \"error\", ignoring case."},
			{Command: "grep -c -v ^# config.ini", Explain: "Count the lines that are not comments."},
		},
		ExitStatus: "0  a line matched.\n1  no lines matched.\n2  an error occurred.",
	})
	patterns := fs.StringArrayP("regexp", "e", nil, "use PATTERN for matching (may be repeated)")
	ignoreCase := fs.BoolP("ignore-case", "i", false, "ignore case distinctions")
	invert := fs.BoolP("invert-match", "v", false, "select non-matching lines")
	lineNum := fs.BoolP("line-number", "n", false, "print line number with output lines")
	count := fs.BoolP("count", "c", false, "print only a count of matching lines per FILE")
	recursive := fs.BoolP("recursive", "r", false, "search directories recursively")
	fs.BoolP("recursive-R", "R", false, "search directories recursively")
	filesMatch := fs.BoolP("files-with-matches", "l", false, "print only names of FILEs with matches")
	filesNoMat := fs.BoolP("files-without-match", "L", false, "print only names of FILEs with no match")
	fixed := fs.BoolP("fixed-strings", "F", false, "PATTERN is a set of newline-separated strings")
	_ = fs.BoolP("extended-regexp", "E", false, "PATTERN is an extended regular expression (default)")
	word := fs.BoolP("word-regexp", "w", false, "match only whole words")
	quiet := fs.BoolP("quiet", "q", false, "suppress all normal output")
	withName := fs.BoolP("with-filename", "H", false, "print the file name for each match")
	noName := fs.BoolP("no-filename", "h", false, "suppress the file name prefix on output")
	byteOffset := fs.BoolP("byte-offset", "b", false, "print the byte offset with output lines")

	after := fs.IntP("after-context", "A", 0, "print NUM lines of trailing context")
	before := fs.IntP("before-context", "B", 0, "print NUM lines of leading context")
	contextN := fs.IntP("context", "C", 0, "print NUM lines of output context")

	colorWhen := fs.String("color", "", "use markers to highlight matches; WHEN is always, never, or auto")
	colourWhen := fs.String("colour", "", "alias for --color")

	include := fs.String("include", "", "search only files that match GLOB (recursive)")
	exclude := fs.String("exclude", "", "skip files that match GLOB")
	excludeDir := fs.String("exclude-dir", "", "skip directories that match GLOB (recursive)")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	recurse := *recursive
	if r, _ := fs.GetBool("recursive-R"); r {
		recurse = true
	}

	// -C NUM sets both before and after unless they were given explicitly.
	a, b := *after, *before
	if *contextN > 0 {
		if !fs.Changed("after-context") {
			a = *contextN
		}
		if !fs.Changed("before-context") {
			b = *contextN
		}
	}

	when := *colorWhen
	if fs.Changed("colour") {
		when = *colourWhen
	}
	color, cerr := resolveColor(fs, when, stdio.Out)
	if cerr != nil {
		_, _ = fmt.Fprintf(stdio.Err, "grep: %v\n", cerr)
		return &command.ExitError{Code: 2}
	}

	opts := options{
		ignoreCase: *ignoreCase, invert: *invert, lineNum: *lineNum,
		count: *count, recursive: recurse, filesMatch: *filesMatch,
		filesNoMat: *filesNoMat, fixed: *fixed, word: *word, quiet: *quiet,
		withName: *withName, noName: *noName, byteOffset: *byteOffset,
		after: a, before: b, color: color,
		include: *include, exclude: *exclude, excludeDir: *excludeDir,
	}

	operands := fs.Args()
	pats := *patterns
	if len(pats) == 0 {
		if len(operands) == 0 {
			_, _ = fmt.Fprintln(stdio.Err, "grep: missing pattern")
			return &command.ExitError{Code: 2}
		}
		pats = []string{operands[0]}
		operands = operands[1:]
	}

	re, err := compile(pats, opts)
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "grep: %v\n", err)
		return &command.ExitError{Code: 2}
	}

	files := operands
	if len(files) == 0 {
		files = []string{"-"}
	}

	g := &grepper{opts: opts, re: re, stdio: stdio, multi: len(files) > 1 || opts.recursive}
	return g.run(files)
}

// resolveColor maps the --color WHEN value to a boolean. An empty WHEN means the
// flag was used with no value (GNU treats "--color" as "always"); when the flag
// is absent entirely, color is off.
func resolveColor(fs *command.FlagSet, when string, out io.Writer) (bool, error) {
	if !fs.Changed("color") && !fs.Changed("colour") {
		return false, nil
	}
	switch when {
	case "", "always":
		return true, nil
	case "never":
		return false, nil
	case "auto":
		return isTerminal(out), nil
	default:
		return false, fmt.Errorf("invalid argument %q for '--color'", when)
	}
}

// compile builds a single regular expression that matches any of the patterns.
func compile(pats []string, opts options) (*regexp.Regexp, error) {
	parts := make([]string, 0, len(pats))
	for _, p := range pats {
		if opts.fixed {
			p = regexp.QuoteMeta(p)
		}
		if opts.word {
			p = `\b(?:` + p + `)\b`
		}
		parts = append(parts, "(?:"+p+")")
	}
	expr := strings.Join(parts, "|")
	if opts.ignoreCase {
		expr = "(?i)" + expr
	}
	return regexp.Compile(expr)
}

// grepper carries the shared state for one grep invocation.
type grepper struct {
	opts    options
	re      *regexp.Regexp
	stdio   command.IO
	multi   bool
	matched bool
	failed  bool

	// printedAny tracks whether any output line has been written so far, so
	// that the "--" group separator is only emitted between groups.
	printedAny bool
}

// run searches every file (recursing into directories when -r is set) and
// returns the GNU-style exit status: 0 if any line matched, 1 if none, 2 on
// error.
func (g *grepper) run(files []string) error {
	for _, f := range files {
		if g.opts.recursive && f != "-" {
			g.walk(f)
			continue
		}
		g.searchFile(f)
	}
	if g.failed {
		return &command.ExitError{Code: 2}
	}
	if !g.matched {
		return &command.ExitError{Code: 1}
	}
	return nil
}

// walk recursively searches every regular file under root, honoring the
// --include, --exclude and --exclude-dir glob filters.
func (g *grepper) walk(root string) {
	_ = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			_, _ = fmt.Fprintf(g.stdio.Err, "grep: %s\n", command.FileError(p, err))
			g.failed = true
			return nil
		}
		base := filepath.Base(p)
		if d.IsDir() {
			if g.opts.excludeDir != "" && p != root && globMatch(g.opts.excludeDir, base) {
				return filepath.SkipDir
			}
			return nil
		}
		if !g.includeFile(base) {
			return nil
		}
		g.searchFile(p)
		return nil
	})
}

// includeFile reports whether a regular file with the given base name should be
// searched given the --include / --exclude filters.
func (g *grepper) includeFile(base string) bool {
	if g.opts.include != "" && !globMatch(g.opts.include, base) {
		return false
	}
	if g.opts.exclude != "" && globMatch(g.opts.exclude, base) {
		return false
	}
	return true
}

// globMatch reports whether name matches the shell pattern. A malformed pattern
// never matches.
func globMatch(pattern, name string) bool {
	ok, err := path.Match(pattern, name)
	return err == nil && ok
}

// searchFile opens name (or stdin for "-") and scans it line by line.
func (g *grepper) searchFile(name string) {
	r, err := command.Open(g.stdio, name)
	if err != nil {
		_, _ = fmt.Fprintf(g.stdio.Err, "grep: %s\n", command.FileError(name, err))
		g.failed = true
		return
	}
	defer func() { _ = r.Close() }()

	display := name
	if name == "-" {
		display = "(standard input)"
	}

	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	contextNeeded := g.opts.after > 0 || g.opts.before > 0
	if g.opts.count || g.opts.filesMatch || g.opts.filesNoMat || g.opts.quiet {
		contextNeeded = false
	}

	var count int
	lineNo := 0
	offset := 0
	// ring holds recent non-matching lines for -B context.
	var ring []ctxLine
	pending := 0     // remaining -A trailing-context lines to print
	lastPrinted := 0 // line number of the most recently printed line (0 = none)

	for scanner.Scan() {
		lineNo++
		line := scanner.Text()
		byteOff := offset
		offset += len(line) + 1 // +1 for the stripped newline

		isMatch := g.re.MatchString(line) != g.opts.invert
		if !isMatch {
			if contextNeeded {
				// Emit trailing context for a recent match.
				if pending > 0 {
					g.printContextLine(display, lineNo, byteOff, line, &lastPrinted)
					pending--
				} else if g.opts.before > 0 {
					ring = appendRing(ring, ctxLine{no: lineNo, off: byteOff, text: line}, g.opts.before)
				}
			}
			continue
		}

		g.matched = true
		count++
		if g.opts.quiet {
			return
		}
		if g.opts.count || g.opts.filesMatch || g.opts.filesNoMat {
			continue
		}

		if contextNeeded {
			// Print buffered leading context, with a "--" group separator
			// when this group is detached from the previous output.
			g.emitGroupSeparator(ring, lineNo, lastPrinted)
			for _, cl := range ring {
				g.printContextLine(display, cl.no, cl.off, cl.text, &lastPrinted)
			}
			ring = ring[:0]
			g.printMatchLine(display, lineNo, byteOff, line, &lastPrinted)
			pending = g.opts.after
			continue
		}

		g.printMatchLine(display, lineNo, byteOff, line, &lastPrinted)
	}

	if g.opts.filesMatch && count > 0 {
		_, _ = fmt.Fprintln(g.stdio.Out, display)
	}
	if g.opts.filesNoMat && count == 0 {
		_, _ = fmt.Fprintln(g.stdio.Out, display)
	}
	if g.opts.count {
		if g.showName() {
			_, _ = fmt.Fprintf(g.stdio.Out, "%s:%d\n", display, count)
		} else {
			_, _ = fmt.Fprintf(g.stdio.Out, "%d\n", count)
		}
	}
}

// ctxLine is a buffered context line awaiting possible printing.
type ctxLine struct {
	no   int
	off  int
	text string
}

// appendRing appends cl to ring, keeping at most max trailing entries.
func appendRing(ring []ctxLine, cl ctxLine, max int) []ctxLine {
	ring = append(ring, cl)
	if len(ring) > max {
		ring = ring[len(ring)-max:]
	}
	return ring
}

// emitGroupSeparator prints the GNU "--" separator before a new context group
// when the group does not directly continue the previous output. matchLine is
// the line number of the match that opens the group.
func (g *grepper) emitGroupSeparator(ring []ctxLine, matchLine, lastPrinted int) {
	if !g.printedAny {
		return
	}
	first := matchLine
	if len(ring) > 0 {
		first = ring[0].no
	}
	if first > lastPrinted+1 {
		_, _ = fmt.Fprintln(g.stdio.Out, "--")
	}
}

// printMatchLine prints a matching line, applying color highlighting when
// enabled.
func (g *grepper) printMatchLine(name string, lineNo, byteOff int, line string, lastPrinted *int) {
	g.writeLine(name, lineNo, byteOff, line, ':', true)
	*lastPrinted = lineNo
	g.printedAny = true
}

// printContextLine prints a non-matching context line using the '-' separator.
func (g *grepper) printContextLine(name string, lineNo, byteOff int, line string, lastPrinted *int) {
	g.writeLine(name, lineNo, byteOff, line, '-', false)
	*lastPrinted = lineNo
	g.printedAny = true
}

// writeLine assembles the prefix (file name, line number, byte offset) and the
// line body, using sep as the field separator. When highlight is true and color
// is enabled, the matched substrings are wrapped in ANSI escapes.
func (g *grepper) writeLine(name string, lineNo, byteOff int, line string, sep byte, highlight bool) {
	var bld strings.Builder
	if g.showName() {
		bld.WriteString(name)
		bld.WriteByte(sep)
	}
	if g.opts.lineNum {
		bld.WriteString(strconv.Itoa(lineNo))
		bld.WriteByte(sep)
	}
	if g.opts.byteOffset {
		bld.WriteString(strconv.Itoa(byteOff))
		bld.WriteByte(sep)
	}
	if g.opts.color && highlight {
		bld.WriteString(g.colorize(line))
	} else {
		bld.WriteString(line)
	}
	_, _ = fmt.Fprintln(g.stdio.Out, bld.String())
}

// colorize wraps every non-overlapping match in line with the GNU highlight
// escapes. Non-matching text passes through unchanged.
func (g *grepper) colorize(line string) string {
	idx := g.re.FindAllStringIndex(line, -1)
	if len(idx) == 0 {
		return line
	}
	var bld strings.Builder
	last := 0
	for _, m := range idx {
		if m[0] < last { // overlapping; skip
			continue
		}
		bld.WriteString(line[last:m[0]])
		bld.WriteString(colorStart)
		bld.WriteString(line[m[0]:m[1]])
		bld.WriteString(colorReset)
		last = m[1]
	}
	bld.WriteString(line[last:])
	return bld.String()
}

// showName reports whether output lines should be prefixed with the file name.
func (g *grepper) showName() bool {
	if g.opts.noName {
		return false
	}
	if g.opts.withName {
		return true
	}
	return g.multi
}
