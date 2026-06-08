// Package shuf implements the shuf applet: write a random permutation of its
// input lines (or of a numeric range, or of its arguments).
package shuf

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"math/rand"
	"strconv"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the shuf applet.
type Command struct{}

// New returns a shuf command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "shuf" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Generate a random permutation of input lines" }

// Run executes shuf.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... [FILE]", stdio.Err)
	echo := fs.BoolP("echo", "e", false, "treat each ARG as an input line")
	inputRange := fs.StringP("input-range", "i", "", "treat each number LO through HI as an input line")
	count := fs.IntP("head-count", "n", -1, "output at most COUNT lines")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	lines, err := c.collect(stdio, fs.Args(), *echo, *inputRange)
	if err != nil {
		return err
	}

	rand.Shuffle(len(lines), func(i, j int) { lines[i], lines[j] = lines[j], lines[i] })
	if *count >= 0 && *count < len(lines) {
		lines = lines[:*count]
	}

	for _, line := range lines {
		if _, err := io.WriteString(stdio.Out, line+"\n"); err != nil {
			return command.Failure(err)
		}
	}
	return nil
}

// collect resolves the lines to shuffle from the chosen input mode: -e uses the
// arguments, -i uses a numeric range, otherwise lines come from the file (or
// standard input).
func (c *Command) collect(stdio command.IO, args []string, echo bool, inputRange string) ([]string, error) {
	switch {
	case echo:
		return args, nil
	case inputRange != "":
		return expandRange(stdio, inputRange)
	default:
		name := "-"
		if len(args) > 0 {
			name = args[0]
		}
		return readLines(stdio, name)
	}
}

// expandRange turns "LO-HI" into the slice of numbers from LO to HI inclusive.
func expandRange(stdio command.IO, spec string) ([]string, error) {
	lo, hi, ok := strings.Cut(spec, "-")
	loN, err1 := strconv.Atoi(lo)
	hiN, err2 := strconv.Atoi(hi)
	if !ok || err1 != nil || err2 != nil || loN > hiN {
		_, _ = fmt.Fprintf(stdio.Err, "shuf: invalid input range %q\n", spec)
		return nil, command.SilentFailure()
	}
	lines := make([]string, 0, hiN-loN+1)
	for i := loN; i <= hiN; i++ {
		lines = append(lines, strconv.Itoa(i))
	}
	return lines, nil
}

// readLines reads name fully and returns its lines without line endings.
func readLines(stdio command.IO, name string) ([]string, error) {
	r, err := command.Open(stdio, name)
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "shuf: %s\n", command.FileError(name, err))
		return nil, command.SilentFailure()
	}
	defer func() { _ = r.Close() }()

	var lines []string
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	if err := sc.Err(); err != nil {
		return nil, command.Failure(err)
	}
	return lines, nil
}
