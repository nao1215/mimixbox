// Package factor implements the factor applet: print the prime factorization of
// each integer operand (or of integers read from standard input).
package factor

import (
	"bufio"
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the factor applet.
type Command struct{}

// New returns a factor command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "factor" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Print the prime factors of each NUMBER" }

// Run executes factor.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[NUMBER]...", stdio.Err).WithHelp(command.Help{
		Description: "Print the prime factors of each NUMBER. With no NUMBER, read whitespace-" +
			"separated numbers from standard input. Numbers must fit in a 64-bit unsigned integer.",
		Examples: []command.Example{
			{Command: "factor 360", Explain: "Print: 360: 2 2 2 3 3 5"},
			{Command: "echo 97 | factor", Explain: "Factor numbers read from standard input."},
		},
		ExitStatus: "0  all operands were factored.\n1  an operand was not a valid number.",
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	numbers := fs.Args()
	if len(numbers) == 0 {
		return factorStream(stdio)
	}

	var failed bool
	for _, n := range numbers {
		if err := factorOne(stdio, n); err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "factor: %v\n", err)
			failed = true
		}
	}
	if failed {
		return command.SilentFailure()
	}
	return nil
}

// factorStream factors every whitespace-separated token on standard input.
func factorStream(stdio command.IO) error {
	sc := bufio.NewScanner(stdio.In)
	sc.Split(bufio.ScanWords)
	var failed bool
	for sc.Scan() {
		if err := factorOne(stdio, sc.Text()); err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "factor: %v\n", err)
			failed = true
		}
	}
	if failed {
		return command.SilentFailure()
	}
	return nil
}

// factorOne prints "N: f1 f2 ..." for a single token.
func factorOne(stdio command.IO, token string) error {
	n, err := strconv.ParseUint(strings.TrimSpace(token), 10, 64)
	if err != nil {
		return fmt.Errorf("%q is not a valid positive integer", token)
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%d:", n)
	for _, f := range primeFactors(n) {
		fmt.Fprintf(&b, " %d", f)
	}
	_, _ = fmt.Fprintln(stdio.Out, b.String())
	return nil
}

// primeFactors returns n's prime factors in ascending order (with multiplicity).
// 0 and 1 have no prime factors.
func primeFactors(n uint64) []uint64 {
	var factors []uint64
	for n%2 == 0 {
		factors = append(factors, 2)
		n /= 2
	}
	for f := uint64(3); f*f <= n; f += 2 {
		for n%f == 0 {
			factors = append(factors, f)
			n /= f
		}
	}
	if n > 1 {
		factors = append(factors, n)
	}
	return factors
}
