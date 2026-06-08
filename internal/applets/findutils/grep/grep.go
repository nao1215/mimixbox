// Package grep implements the grep applet: search files (or standard input) for
// lines matching a pattern and print them. The pattern is a Go (RE2) regular
// expression, which covers the common extended-regexp use; -F switches to plain
// fixed-string matching.
package grep

import (
	"bufio"
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the grep applet.
type Command struct{}

// New returns a grep command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "grep" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Print lines that match patterns" }

// options holds the parsed switches.
type options struct {
	ignoreCase bool
	invert     bool
	lineNum    bool
	count      bool
	recursive  bool
	filesMatch bool
	fixed      bool
	word       bool
	quiet      bool
	withName   bool
	noName     bool
}

// Run executes grep.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... PATTERN [FILE]...", stdio.Err)
	patterns := fs.StringArrayP("regexp", "e", nil, "use PATTERN for matching (may be repeated)")
	ignoreCase := fs.BoolP("ignore-case", "i", false, "ignore case distinctions")
	invert := fs.BoolP("invert-match", "v", false, "select non-matching lines")
	lineNum := fs.BoolP("line-number", "n", false, "print line number with output lines")
	count := fs.BoolP("count", "c", false, "print only a count of matching lines per FILE")
	recursive := fs.BoolP("recursive", "r", false, "search directories recursively")
	fs.BoolP("recursive-R", "R", false, "search directories recursively")
	filesMatch := fs.BoolP("files-with-matches", "l", false, "print only names of FILEs with matches")
	fixed := fs.BoolP("fixed-strings", "F", false, "PATTERN is a set of newline-separated strings")
	_ = fs.BoolP("extended-regexp", "E", false, "PATTERN is an extended regular expression (default)")
	word := fs.BoolP("word-regexp", "w", false, "match only whole words")
	quiet := fs.BoolP("quiet", "q", false, "suppress all normal output")
	withName := fs.BoolP("with-filename", "H", false, "print the file name for each match")
	noName := fs.BoolP("no-filename", "h", false, "suppress the file name prefix on output")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	recurse := *recursive
	if r, _ := fs.GetBool("recursive-R"); r {
		recurse = true
	}

	opts := options{
		ignoreCase: *ignoreCase, invert: *invert, lineNum: *lineNum,
		count: *count, recursive: recurse, filesMatch: *filesMatch,
		fixed: *fixed, word: *word, quiet: *quiet,
		withName: *withName, noName: *noName,
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

// walk recursively searches every regular file under root.
func (g *grepper) walk(root string) {
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			_, _ = fmt.Fprintf(g.stdio.Err, "grep: %s\n", command.FileError(path, err))
			g.failed = true
			return nil
		}
		if d.IsDir() {
			return nil
		}
		g.searchFile(path)
		return nil
	})
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
	var count int
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := scanner.Text()
		if g.re.MatchString(line) == g.opts.invert {
			continue
		}
		g.matched = true
		count++
		if g.opts.quiet {
			return
		}
		if g.opts.count || g.opts.filesMatch {
			continue
		}
		g.printLine(display, lineNo, line)
	}

	if g.opts.filesMatch && count > 0 {
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

// printLine writes one matching line with the optional file-name and
// line-number prefixes.
func (g *grepper) printLine(name string, lineNo int, line string) {
	var b strings.Builder
	if g.showName() {
		b.WriteString(name)
		b.WriteByte(':')
	}
	if g.opts.lineNum {
		fmt.Fprintf(&b, "%d:", lineNo)
	}
	b.WriteString(line)
	_, _ = fmt.Fprintln(g.stdio.Out, b.String())
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
