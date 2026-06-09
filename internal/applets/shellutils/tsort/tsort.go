// Package tsort implements the tsort applet: topological sort. It reads
// whitespace-separated pairs (each "a b" meaning a must come before b) from a
// file or standard input and prints a total ordering, one item per line.
package tsort

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the tsort applet.
type Command struct{}

// New returns a tsort command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "tsort" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Topological sort of a directed graph" }

// Run executes tsort.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[FILE]", stdio.Err).WithHelp(command.Help{
		Description: "Read pairs of items (whitespace-separated; each pair 'a b' means a precedes b) " +
			"from FILE or standard input, and write a total ordering consistent with those " +
			"constraints, one item per line. A lone item with no pair is still listed.",
		Examples: []command.Example{
			{Command: "printf 'a b\\nb c\\n' | tsort", Explain: "Print a, b, c in dependency order."},
		},
		ExitStatus: "0  a valid ordering was produced.\n1  the input contained a cycle or an odd token.",
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	r := stdio.In
	if rest := fs.Args(); len(rest) > 0 && rest[0] != "-" {
		f, oerr := os.Open(rest[0]) //nolint:gosec // user-named file
		if oerr != nil {
			_, _ = fmt.Fprintf(stdio.Err, "tsort: %s\n", command.FileError(rest[0], oerr))
			return command.SilentFailure()
		}
		defer func() { _ = f.Close() }()
		r = f
	}

	tokens, err := readTokens(r)
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "tsort: %v\n", err)
		return command.SilentFailure()
	}
	if len(tokens)%2 != 0 {
		_, _ = fmt.Fprintln(stdio.Err, "tsort: input contains an odd number of tokens")
		return command.SilentFailure()
	}

	order, cyclic := topoSort(tokens)
	for _, n := range order {
		_, _ = fmt.Fprintln(stdio.Out, n)
	}
	if cyclic {
		_, _ = fmt.Fprintln(stdio.Err, "tsort: input contains a loop")
		return command.SilentFailure()
	}
	return nil
}

func readTokens(r io.Reader) ([]string, error) {
	sc := bufio.NewScanner(r)
	sc.Split(bufio.ScanWords)
	var tokens []string
	for sc.Scan() {
		tokens = append(tokens, sc.Text())
	}
	return tokens, sc.Err()
}

// topoSort returns a topological order of the graph built from consecutive
// token pairs, using Kahn's algorithm with deterministic tie-breaking. cyclic is
// true when not all nodes could be ordered (a loop), in which case the remaining
// nodes are appended in name order so output is still complete.
func topoSort(tokens []string) (order []string, cyclic bool) {
	adj := map[string]map[string]bool{}
	indeg := map[string]int{}
	var nodesInOrder []string
	seen := map[string]bool{}

	note := func(n string) {
		if !seen[n] {
			seen[n] = true
			nodesInOrder = append(nodesInOrder, n)
			indeg[n] = 0
			adj[n] = map[string]bool{}
		}
	}

	for i := 0; i+1 < len(tokens); i += 2 {
		a, b := tokens[i], tokens[i+1]
		note(a)
		note(b)
		if a != b && !adj[a][b] {
			adj[a][b] = true
			indeg[b]++
		}
	}

	// Ready set, kept sorted for deterministic output.
	var ready []string
	for _, n := range nodesInOrder {
		if indeg[n] == 0 {
			ready = append(ready, n)
		}
	}
	sort.Strings(ready)

	placed := map[string]bool{}
	for len(ready) > 0 {
		n := ready[0]
		ready = ready[1:]
		order = append(order, n)
		placed[n] = true
		var newly []string
		for m := range adj[n] {
			indeg[m]--
			if indeg[m] == 0 {
				newly = append(newly, m)
			}
		}
		sort.Strings(newly)
		ready = append(ready, newly...)
		sort.Strings(ready)
	}

	if len(order) != len(nodesInOrder) {
		cyclic = true
		var leftover []string
		for _, n := range nodesInOrder {
			if !placed[n] {
				leftover = append(leftover, n)
			}
		}
		sort.Strings(leftover)
		order = append(order, leftover...)
	}
	return order, cyclic
}
