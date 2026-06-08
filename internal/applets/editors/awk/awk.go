// Package awk implements the awk applet: a deliberately small subset of the awk
// language that covers the everyday one-liners. It supports BEGIN/END blocks,
// /regexp/ and comparison patterns, field variables ($0, $1..$NF, NR, NF) and
// the print statement with -F to choose the field separator and -v to preset
// variables.
package awk

import (
	"bufio"
	"context"
	"fmt"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the awk applet.
type Command struct{}

// New returns an awk command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "awk" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Pattern scanning and processing language" }

// Run executes awk.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-F sep] [-v var=val]... 'program' [FILE]...", stdio.Err)
	sep := fs.StringP("field-separator", "F", "", "use sep as the input field separator")
	assigns := fs.StringArrayP("assign", "v", nil, "assign var=value before execution")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	operands := fs.Args()
	if len(operands) == 0 {
		_, _ = fmt.Fprintln(stdio.Err, "awk: no program text")
		return command.SilentFailure()
	}
	programText := operands[0]
	files := operands[1:]

	rules, err := parseProgram(programText)
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "awk: %v\n", err)
		return command.SilentFailure()
	}

	st := &state{
		out:  stdio.Out,
		fs:   *sep,
		ofs:  " ",
		vars: map[string]string{},
	}
	for _, a := range *assigns {
		if k, v, ok := strings.Cut(a, "="); ok {
			st.vars[k] = v
		}
	}

	// BEGIN blocks run before any input.
	for _, r := range rules {
		if r.kind == ruleBegin {
			st.exec(r)
		}
	}

	if hasMainOrEnd(rules) {
		if err := st.process(stdio, files, rules); err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "awk: %v\n", err)
			return command.SilentFailure()
		}
	}

	for _, r := range rules {
		if r.kind == ruleEnd {
			st.exec(r)
		}
	}
	return nil
}

// hasMainOrEnd reports whether any rule needs the input to be read.
func hasMainOrEnd(rules []rule) bool {
	for _, r := range rules {
		if r.kind != ruleBegin {
			return true
		}
	}
	return false
}

// process reads each input file (or stdin) line by line and applies the
// non-BEGIN rules.
func (st *state) process(stdio command.IO, files []string, rules []rule) error {
	if len(files) == 0 {
		files = []string{"-"}
	}
	for _, f := range files {
		r, err := command.Open(stdio, f)
		if err != nil {
			return fmt.Errorf("%s", command.FileError(f, err))
		}
		scanner := bufio.NewScanner(r)
		scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
		for scanner.Scan() {
			st.nr++
			st.setLine(scanner.Text())
			for _, rl := range rules {
				if rl.kind == ruleMain && st.match(rl.pattern) {
					st.exec(rl)
				}
			}
		}
		_ = r.Close()
		if err := scanner.Err(); err != nil {
			return err
		}
	}
	return nil
}
