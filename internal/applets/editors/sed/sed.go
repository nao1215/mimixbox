// Package sed implements the sed applet: a stream editor that applies a small
// but practical subset of GNU sed scripts to its input. It supports the
// substitute command (s/re/repl/flags), p (print), d (delete) and q (quit),
// each with an optional line-number, $ or /regexp/ address (single or range).
package sed

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the sed applet.
type Command struct{}

// New returns a sed command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "sed" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Stream editor for filtering and transforming text" }

// Run executes sed.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... {SCRIPT} [FILE]...", stdio.Err).WithHelp(command.Help{
		Description: "A stream editor that applies a SCRIPT to each FILE (or to standard input) and " +
			"writes the result to standard output. It supports a practical subset of GNU sed: the " +
			"substitute command s/re/repl/flags, plus p (print), d (delete) and q (quit), each " +
			"with an optional line-number, $ or /regexp/ address, given singly or as a range. The " +
			"script is taken from -e/-f or from the first operand; -n suppresses automatic " +
			"printing, -E/-r enable extended regular expressions, and -i edits files in place.",
		Examples: []command.Example{
			{Command: "sed 's/foo/bar/' file.txt", Explain: "Replace the first 'foo' on each line with 'bar'."},
			{Command: "sed -n '2,4p' file.txt", Explain: "Print only lines 2 through 4."},
			{Command: "sed '/^#/d' config", Explain: "Delete every comment line."},
			{Command: "sed -i 's/old/new/g' file.txt", Explain: "Edit file.txt in place, replacing all 'old' with 'new'."},
		},
		ExitStatus: "0  the script ran successfully on all input.\n1  the script was invalid or a file could not be read.",
	})
	scripts := fs.StringArrayP("expression", "e", nil, "add the script to the commands to be executed")
	scriptFile := fs.StringP("file", "f", "", "add the contents of FILE to the commands")
	quiet := fs.BoolP("quiet", "n", false, "suppress automatic printing of pattern space")
	extended := fs.BoolP("regexp-extended", "E", false, "use extended regular expressions")
	fs.BoolP("regexp-extended-r", "r", false, "use extended regular expressions (alias of -E)")
	inplace := fs.BoolP("in-place", "i", false, "edit files in place")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	// -E and -r are aliases that select extended regular expressions.
	useExtended := *extended
	if r, _ := fs.GetBool("regexp-extended-r"); r {
		useExtended = true
	}

	operands := fs.Args()
	scriptText, files, err := gatherScript(*scripts, *scriptFile, operands)
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "sed: %v\n", err)
		return command.SilentFailure()
	}

	program, err := parse(scriptText, useExtended)
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "sed: -e expression: %v\n", err)
		return command.SilentFailure()
	}

	if *inplace {
		if len(files) == 0 {
			_, _ = fmt.Fprintln(stdio.Err, "sed: no input files for in-place editing")
			return command.SilentFailure()
		}
		return runInPlace(stdio, program, files, *quiet)
	}

	if len(files) == 0 {
		files = []string{"-"}
	}
	return runStreaming(stdio, program, files, *quiet)
}

// gatherScript resolves the script text from -e, -f or the first operand, and
// returns the remaining operands as the file list.
func gatherScript(exprs []string, file string, operands []string) (string, []string, error) {
	var parts []string
	parts = append(parts, exprs...)
	if file != "" {
		data, err := os.ReadFile(file) //nolint:gosec // user-named script file
		if err != nil {
			return "", nil, err
		}
		parts = append(parts, string(data))
	}

	files := operands
	if len(parts) == 0 {
		if len(operands) == 0 {
			return "", nil, fmt.Errorf("no script specified")
		}
		parts = append(parts, operands[0])
		files = operands[1:]
	}
	return strings.Join(parts, "\n"), files, nil
}

// runStreaming applies the program to the concatenation of files, writing to
// stdout.
func runStreaming(stdio command.IO, program []cmd, files []string, quiet bool) error {
	ed := &editor{program: program, quiet: quiet, out: stdio.Out}
	var lines []string
	for _, f := range files {
		ls, err := readLines(stdio, f)
		if err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "sed: %s\n", command.FileError(f, err))
			return command.SilentFailure()
		}
		lines = append(lines, ls...)
	}
	ed.run(lines)
	return nil
}

// runInPlace rewrites each file with the result of applying the program. Each
// file gets a fresh copy of the program so that range (two-address) state from
// one file does not leak into the next.
func runInPlace(stdio command.IO, program []cmd, files []string, quiet bool) error {
	var failed bool
	for _, f := range files {
		lines, err := readLines(stdio, f)
		if err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "sed: %s\n", command.FileError(f, err))
			failed = true
			continue
		}
		var b strings.Builder
		ed := &editor{program: cloneProgram(program), quiet: quiet, out: &b}
		ed.run(lines)
		if err := writeFilePreservingMode(f, b.String()); err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "sed: %v\n", err)
			failed = true
		}
	}
	if failed {
		return command.SilentFailure()
	}
	return nil
}

// cloneProgram returns a copy of the program with all range state reset, so a
// fresh run starts with no command mid-range.
func cloneProgram(program []cmd) []cmd {
	out := make([]cmd, len(program))
	copy(out, program)
	for i := range out {
		out[i].active = false
	}
	return out
}

// writeFilePreservingMode writes content to name atomically (via a temp file in
// the same directory and a rename) while preserving the file's original
// permission bits.
func writeFilePreservingMode(name, content string) error {
	info, err := os.Stat(name)
	mode := os.FileMode(0o644)
	if err == nil {
		mode = info.Mode().Perm()
	}

	dir := filepath.Dir(name)
	tmp, err := os.CreateTemp(dir, ".sed-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if _, err := tmp.WriteString(content); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	if err := os.Chmod(tmpName, mode); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	if err := os.Rename(tmpName, name); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	return nil
}

// readLines reads name (or stdin for "-") into a slice of lines without the
// trailing newline.
func readLines(stdio command.IO, name string) ([]string, error) {
	r, err := command.Open(stdio, name)
	if err != nil {
		return nil, err
	}
	defer func() { _ = r.Close() }()

	var lines []string
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// strconvAtoi is a thin wrapper so callers read clearly.
func strconvAtoi(s string) (int, error) { return strconv.Atoi(s) }
