// Package uniq implements the uniq applet: filter adjacent matching lines from
// INPUT (or standard input), writing the result to OUTPUT (or standard output),
// with the common GNU options.
package uniq

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the uniq applet.
type Command struct{}

// New returns a uniq command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "uniq" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Report or omit repeated lines" }

type options struct {
	count      bool
	repeated   bool
	unique     bool
	ignoreCase bool
}

// Run executes uniq.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... [INPUT [OUTPUT]]", stdio.Err).WithHelp(command.Help{
		Description: "Filter adjacent matching lines from INPUT (or standard input), writing to OUTPUT (or standard output). " +
			"Only consecutive duplicates are collapsed, so the input is usually sorted first.",
		Examples: []command.Example{
			{Command: "uniq names.txt", Explain: "Collapse adjacent duplicate lines."},
			{Command: "uniq -c names.txt", Explain: "Prefix each line with its number of occurrences."},
			{Command: "uniq -d names.txt", Explain: "Print only lines that were repeated."},
		},
		ExitStatus: "0  the lines were filtered successfully.\n1  an input or output file could not be opened or read.",
	})
	count := fs.BoolP("count", "c", false, "prefix lines by the number of occurrences")
	repeated := fs.BoolP("repeated", "d", false, "only print duplicate lines, one for each group")
	unique := fs.BoolP("unique", "u", false, "only print unique lines")
	ignoreCase := fs.BoolP("ignore-case", "i", false, "ignore differences in case when comparing")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	opts := options{
		count:      *count,
		repeated:   *repeated,
		unique:     *unique,
		ignoreCase: *ignoreCase,
	}

	operands := fs.Args()
	inputName := "-"
	if len(operands) >= 1 {
		inputName = operands[0]
	}

	r, err := command.Open(stdio, inputName)
	if err != nil {
		return command.Failuref("%s", command.FileError(inputName, err))
	}
	defer func() { _ = r.Close() }()

	lines, readErr := readLines(r)
	if readErr != nil {
		return command.Failuref("%s", command.FileError(inputName, readErr))
	}

	out := filter(lines, opts)

	w, closeW, err := openOutput(stdio, operands)
	if err != nil {
		return command.Failuref("%s", command.FileError(operands[1], err))
	}
	defer closeW()

	if _, err := io.WriteString(w, out); err != nil {
		return command.Failure(err)
	}
	return nil
}

// openOutput resolves the optional OUTPUT operand to a writer, defaulting to
// standard output. The returned func closes the writer (a no-op for stdout).
func openOutput(stdio command.IO, operands []string) (io.Writer, func(), error) {
	if len(operands) < 2 || operands[1] == "-" {
		return stdio.Out, func() {}, nil
	}
	f, err := os.Create(operands[1]) //nolint:gosec // operating on a user-named file is the whole point
	if err != nil {
		return nil, nil, err
	}
	return f, func() { _ = f.Close() }, nil
}

// readLines reads r into its lines, dropping the trailing newline of each line.
func readLines(r io.Reader) ([]string, error) {
	var lines []string
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}

// group is a run of adjacent lines that compared equal.
type group struct {
	line  string // the line as first seen (original case preserved)
	count int
}

// filter applies GNU uniq's grouping and option logic to lines, returning the
// rendered output text (each emitted line terminated by a newline). This is the
// pure core, exercised directly by the unit tests.
func filter(lines []string, opts options) string {
	groups := makeGroups(lines, opts.ignoreCase)

	var b strings.Builder
	for _, g := range groups {
		if opts.repeated && g.count < 2 {
			continue
		}
		if opts.unique && g.count > 1 {
			continue
		}
		if opts.count {
			fmt.Fprintf(&b, "%7d %s\n", g.count, g.line)
			continue
		}
		b.WriteString(g.line)
		b.WriteByte('\n')
	}
	return b.String()
}

// makeGroups collapses adjacent equal lines into groups. With ignoreCase the
// comparison is case-insensitive, but the first line of each group is emitted
// verbatim, matching GNU uniq.
func makeGroups(lines []string, ignoreCase bool) []group {
	var groups []group
	for _, l := range lines {
		if n := len(groups); n > 0 && equal(groups[n-1].line, l, ignoreCase) {
			groups[n-1].count++
			continue
		}
		groups = append(groups, group{line: l, count: 1})
	}
	return groups
}

func equal(a, b string, ignoreCase bool) bool {
	if ignoreCase {
		return strings.EqualFold(a, b)
	}
	return a == b
}
