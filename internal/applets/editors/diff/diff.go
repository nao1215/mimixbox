// Package diff implements the diff applet: compare two text files line by line
// and report how to change the first into the second. It produces the classic
// "normal" diff output by default and unified output with -u, and uses the GNU
// exit convention (0 identical, 1 different, 2 error).
package diff

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// maxDPCells caps the size of the LCS dynamic-programming table so that
// comparing very large files cannot exhaust memory (~50M cells of int).
const maxDPCells = 50_000_000

// Command is the diff applet.
type Command struct{}

// New returns a diff command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "diff" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Compare files line by line" }

// Run executes diff.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... FILE1 FILE2", stdio.Err)
	unified := fs.BoolP("unified", "u", false, "output NUM (default 3) lines of unified context")
	brief := fs.BoolP("brief", "q", false, "report only when files differ")
	ignoreCase := fs.BoolP("ignore-case", "i", false, "ignore case differences in file contents")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	names := fs.Args()
	if len(names) < 2 {
		_, _ = fmt.Fprintln(stdio.Err, "diff: missing operand")
		_, _ = fmt.Fprintln(stdio.Err, "diff: Try 'diff --help' for more information.")
		return &command.ExitError{Code: 2}
	}
	if len(names) > 2 {
		_, _ = fmt.Fprintf(stdio.Err, "diff: extra operand '%s'\n", names[2])
		_, _ = fmt.Fprintln(stdio.Err, "diff: Try 'diff --help' for more information.")
		return &command.ExitError{Code: 2}
	}

	a, err := readLines(names[0])
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "diff: %s\n", command.FileError(names[0], err))
		return &command.ExitError{Code: 2}
	}
	b, err := readLines(names[1])
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "diff: %s\n", command.FileError(names[1], err))
		return &command.ExitError{Code: 2}
	}

	// Guard the O(n*m) LCS table against pathological memory use on huge files.
	if len(a) > 0 && int64(len(a))*int64(len(b)) > maxDPCells {
		_, _ = fmt.Fprintln(stdio.Err, "diff: files too large to compare")
		return &command.ExitError{Code: 2}
	}

	ops := diffLines(a, b, *ignoreCase)
	if !hasChange(ops) {
		return nil
	}

	if *brief {
		_, _ = fmt.Fprintf(stdio.Out, "Files %s and %s differ\n", names[0], names[1])
		return &command.ExitError{Code: 1}
	}

	if *unified {
		writeUnified(stdio, names[0], names[1], a, b, ops)
	} else {
		writeNormal(stdio, a, b, ops)
	}
	return &command.ExitError{Code: 1}
}

// readLines reads a file into a slice of lines (without trailing newlines).
func readLines(name string) ([]string, error) {
	data, err := os.ReadFile(name) //nolint:gosec // user-named file
	if err != nil {
		return nil, err
	}
	s := string(data)
	if s == "" {
		return nil, nil
	}
	s = strings.TrimSuffix(s, "\n")
	return strings.Split(s, "\n"), nil
}

// opKind is the type of an edit-script operation.
type opKind int

const (
	opEqual opKind = iota
	opDelete
	opInsert
)

// op is one element of the edit script: ai/bi are 0-based indices into a and b.
type op struct {
	kind opKind
	ai   int
	bi   int
}

// hasChange reports whether the edit script contains any insert or delete.
func hasChange(ops []op) bool {
	for _, o := range ops {
		if o.kind != opEqual {
			return true
		}
	}
	return false
}

// diffLines computes a line-level edit script from a to b using the
// longest-common-subsequence dynamic program.
func diffLines(a, b []string, ignoreCase bool) []op {
	eq := func(x, y string) bool {
		if ignoreCase {
			return strings.EqualFold(x, y)
		}
		return x == y
	}

	n, m := len(a), len(b)
	// lcs[i][j] = length of LCS of a[i:] and b[j:].
	lcs := make([][]int, n+1)
	for i := range lcs {
		lcs[i] = make([]int, m+1)
	}
	for i := n - 1; i >= 0; i-- {
		for j := m - 1; j >= 0; j-- {
			if eq(a[i], b[j]) {
				lcs[i][j] = lcs[i+1][j+1] + 1
			} else if lcs[i+1][j] >= lcs[i][j+1] {
				lcs[i][j] = lcs[i+1][j]
			} else {
				lcs[i][j] = lcs[i][j+1]
			}
		}
	}

	var ops []op
	i, j := 0, 0
	for i < n && j < m {
		switch {
		case eq(a[i], b[j]):
			ops = append(ops, op{kind: opEqual, ai: i, bi: j})
			i++
			j++
		case lcs[i+1][j] >= lcs[i][j+1]:
			ops = append(ops, op{kind: opDelete, ai: i, bi: j})
			i++
		default:
			ops = append(ops, op{kind: opInsert, ai: i, bi: j})
			j++
		}
	}
	for ; i < n; i++ {
		ops = append(ops, op{kind: opDelete, ai: i, bi: j})
	}
	for ; j < m; j++ {
		ops = append(ops, op{kind: opInsert, ai: i, bi: j})
	}
	return ops
}
