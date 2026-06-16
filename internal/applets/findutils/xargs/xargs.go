// Package xargs implements the xargs applet: read items from standard input and
// run a command with those items appended as arguments. It covers the common
// switches (-n, -I, -0, -d, -t, -r, -L, -s, -P) used to glue command pipelines
// together.
package xargs

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"sync"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the xargs applet.
type Command struct{}

// New returns an xargs command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "xargs" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Build and execute command lines from standard input" }

// options holds the parsed switches.
type options struct {
	maxArgs    int    // -n: max arguments per command invocation (0 = unlimited)
	maxLines   int    // -L: max input lines per command invocation (0 = unused)
	maxChars   int    // -s: max length of each constructed command line (0 = unused)
	maxProcs   int    // -P: max concurrent invocations (0 = as many as possible)
	replace    string // -I: replace this token with each input item
	null       bool   // -0: input items are NUL-separated
	delim      string // -d: custom input delimiter
	verbose    bool   // -t: print each command before running it
	noRunEmpty bool   // -r: do not run the command if input is empty
}

// Run executes xargs.
func (c *Command) Run(ctx context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... [COMMAND [INITIAL-ARGS]...]", stdio.Err).WithHelp(command.Help{
		Description: "Read items from standard input and run COMMAND (echo by default) once or more with " +
			"those items appended as arguments. Items are separated by whitespace unless -0 or -d changes the delimiter.",
		Examples: []command.Example{
			{Command: "ls | xargs echo", Explain: "Print all file names on a single line."},
			{Command: "find . -name '*.tmp' | xargs rm", Explain: "Delete every matching file."},
			{Command: "xargs -n 1 -I {} echo {}", Explain: "Run the command once per input item."},
			{Command: "xargs -L 1 echo", Explain: "Run the command once per input line."},
			{Command: "xargs -P 4 -n 1 curl -O", Explain: "Run up to four commands concurrently."},
		},
		ExitStatus: "0  every command succeeded.\n1  the input could not be read or a command failed.",
	})
	maxArgs := fs.IntP("max-args", "n", 0, "use at most MAX-ARGS arguments per command line")
	maxLines := fs.IntP("max-lines", "L", 0, "use at most MAX-LINES non-blank input lines per command line")
	maxChars := fs.IntP("max-chars", "s", 0, "use at most MAX-CHARS characters per command line")
	maxProcs := fs.IntP("max-procs", "P", 1, "run up to MAX-PROCS commands at once (0 = as many as possible)")
	replace := fs.StringP("replace", "I", "", "replace occurrences of REPLACE-STR in the command with input")
	null := fs.BoolP("null", "0", false, "input items are terminated by a null, not whitespace")
	delim := fs.StringP("delimiter", "d", "", "input items are terminated by DELIM")
	verbose := fs.BoolP("verbose", "t", false, "print the command line before executing it")
	noRunEmpty := fs.BoolP("no-run-if-empty", "r", false, "do not run the command if input is empty")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	opts := options{
		maxArgs: *maxArgs, maxLines: *maxLines, maxChars: *maxChars, maxProcs: *maxProcs,
		replace: *replace, null: *null,
		delim: *delim, verbose: *verbose, noRunEmpty: *noRunEmpty,
	}

	cmdArgs := fs.Args()
	cmdName := "echo"
	var initial []string
	if len(cmdArgs) > 0 {
		cmdName = cmdArgs[0]
		initial = cmdArgs[1:]
	}

	items, lines, err := readItems(stdio, opts)
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "xargs: %v\n", err)
		return command.SilentFailure()
	}

	if len(items) == 0 && opts.noRunEmpty {
		return nil
	}

	if err := runCommands(ctx, stdio, cmdName, initial, items, lines, opts); err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "xargs: %v\n", err)
		return command.SilentFailure()
	}
	return nil
}

// readItems reads and splits the input into items according to the -0, -d and
// default-whitespace rules. It also returns, for each item, the zero-based index
// of the input line it came from; this lets -L group items by line. When items
// are not whitespace-split (i.e. -0 or -d), every item is treated as belonging
// to its own line.
func readItems(stdio command.IO, opts options) ([]string, []int, error) {
	// When splitting on NUL or a custom delimiter there are no real "lines", so
	// each item is its own line for the purposes of -L.
	if opts.null || opts.delim != "" {
		items, err := readDelimited(stdio, opts)
		if err != nil {
			return nil, nil, err
		}
		lines := make([]int, len(items))
		for i := range items {
			lines[i] = i
		}
		return items, lines, nil
	}

	// Default: split on whitespace but remember which input line each item came
	// from so -L can group by line.
	scanner := bufio.NewScanner(stdio.In)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	scanner.Split(bufio.ScanLines)

	var items []string
	var lines []int
	lineNo := 0
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		for _, f := range fields {
			items = append(items, f)
			lines = append(lines, lineNo)
		}
		lineNo++
	}
	return items, lines, scanner.Err()
}

// readDelimited reads items separated by NUL (-0) or a custom delimiter (-d).
func readDelimited(stdio command.IO, opts options) ([]string, error) {
	scanner := bufio.NewScanner(stdio.In)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	if opts.null {
		scanner.Split(splitOn(0))
	} else {
		scanner.Split(splitOn(opts.delim[0]))
	}

	var items []string
	for scanner.Scan() {
		tok := scanner.Text()
		if tok == "" {
			continue
		}
		items = append(items, tok)
	}
	return items, scanner.Err()
}

// splitOn returns a bufio.SplitFunc that splits on the byte sep.
func splitOn(sep byte) bufio.SplitFunc {
	return func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		if i := bytes.IndexByte(data, sep); i >= 0 {
			return i + 1, data[:i], nil
		}
		if atEOF {
			return len(data), data, nil
		}
		return 0, nil, nil
	}
}

// runCommands builds and executes the command line(s) from the items.
func runCommands(ctx context.Context, stdio command.IO, name string, initial, items []string, lines []int, opts options) error {
	// -I replaces a token and runs once per input item. This mode ignores the
	// batching switches (-n, -L, -s), matching GNU xargs behavior.
	if opts.replace != "" {
		invocations := make([]invocation, 0, len(items))
		for _, item := range items {
			argv := make([]string, len(initial))
			for i, a := range initial {
				argv[i] = strings.ReplaceAll(a, opts.replace, item)
			}
			invocations = append(invocations, invocation{name: name, argv: argv})
		}
		return runAll(ctx, stdio, invocations, opts)
	}

	batches := batch(items, lines, opts)
	if len(batches) == 0 {
		// No input: run the command once with just the initial args (GNU xargs
		// behavior unless -r was given, which is handled by the caller).
		return runAll(ctx, stdio, []invocation{{name: name, argv: initial}}, opts)
	}
	invocations := make([]invocation, 0, len(batches))
	for _, b := range batches {
		argv := append(append([]string{}, initial...), b...)
		invocations = append(invocations, invocation{name: name, argv: argv})
	}
	return runAll(ctx, stdio, invocations, opts)
}

// batch splits items into groups honoring -n (max args), -L (max input lines)
// and -s (max command-line characters). A new group is started whenever adding
// the next item would exceed any active limit. When no limit applies the items
// form a single group.
func batch(items []string, lines []int, opts options) [][]string {
	if len(items) == 0 {
		return nil
	}

	// charLen is the size of the constructed command-line tail (the appended
	// items) including a separating space before each item, used for -s.
	charLen := func(group []string) int {
		total := 0
		for _, it := range group {
			total += len(it) + 1
		}
		return total
	}

	var out [][]string
	var cur []string
	curLines := 0
	lastLine := -1

	flush := func() {
		if len(cur) > 0 {
			out = append(out, cur)
			cur = nil
			curLines = 0
			lastLine = -1
		}
	}

	for i, it := range items {
		// -L: a new input line begins; count it and split if the line budget for
		// this group is already full.
		if opts.maxLines > 0 && lines[i] != lastLine {
			if curLines >= opts.maxLines {
				flush()
			}
			curLines++
			lastLine = lines[i]
		}
		// -n: split when the per-invocation argument count is reached.
		if opts.maxArgs > 0 && len(cur) >= opts.maxArgs {
			flush()
			if opts.maxLines > 0 {
				curLines = 1
				lastLine = lines[i]
			}
		}
		// -s: split when adding this item would overflow the character budget.
		// A single item longer than the budget still goes out on its own line.
		if opts.maxChars > 0 && len(cur) > 0 && charLen(cur)+len(it)+1 > opts.maxChars {
			flush()
			if opts.maxLines > 0 {
				curLines = 1
				lastLine = lines[i]
			}
		}
		cur = append(cur, it)
	}
	flush()
	return out
}

// invocation is a single command to run: a name plus its full argument vector.
type invocation struct {
	name string
	argv []string
}

// runAll executes the invocations, honoring -P for concurrency. With the default
// of one process (-P 1) it runs them sequentially, streaming output directly to
// preserve the original single-process behavior. With -P > 1 (or -P 0 meaning
// "as many as possible") it uses a bounded worker pool and buffers each command's
// output so concurrent writes do not interleave. The first error is returned.
func runAll(ctx context.Context, stdio command.IO, invs []invocation, opts options) error {
	procs := workerCount(opts.maxProcs, len(invs))
	if procs <= 1 {
		for _, in := range invs {
			if err := exec1(ctx, stdio, in.name, in.argv, opts); err != nil {
				return err
			}
		}
		return nil
	}

	var mu sync.Mutex // serializes output flushes and firstErr access
	var firstErr error

	sem := make(chan struct{}, procs)
	var wg sync.WaitGroup
	for _, in := range invs {
		in := in
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			outBuf := &bytes.Buffer{}
			errBuf := &bytes.Buffer{}
			// Children get no stdin under -P: the shared input reader cannot be
			// read from multiple goroutines safely, and GNU xargs also detaches
			// child stdin when running in parallel.
			local := command.IO{In: nil, Out: outBuf, Err: errBuf}
			err := exec1(ctx, local, in.name, in.argv, opts)

			mu.Lock()
			_, _ = stdio.Out.Write(outBuf.Bytes())
			_, _ = stdio.Err.Write(errBuf.Bytes())
			if err != nil && firstErr == nil {
				firstErr = err
			}
			mu.Unlock()
		}()
	}
	wg.Wait()
	return firstErr
}

// workerCount resolves the effective number of concurrent workers from the -P
// value. Zero means "as many as possible", which we cap at the number of
// invocations (and at least the number of CPUs) so a huge input does not spawn
// an unbounded number of goroutines at once.
func workerCount(maxProcs, n int) int {
	if maxProcs > 0 {
		if maxProcs > n {
			return n
		}
		return maxProcs
	}
	// maxProcs == 0: run as many as possible, bounded by the work and CPUs.
	limit := runtime.NumCPU()
	if limit < 1 {
		limit = 1
	}
	if n < limit {
		return n
	}
	return limit
}

// exec1 runs a single command, wiring it to the shell streams and honoring -t.
func exec1(ctx context.Context, stdio command.IO, name string, argv []string, opts options) error {
	if opts.verbose {
		_, _ = fmt.Fprintln(stdio.Err, strings.TrimSpace(name+" "+strings.Join(argv, " ")))
	}
	cmd := exec.CommandContext(ctx, name, argv...) //nolint:gosec // running a user-named command is the purpose of xargs
	cmd.Stdin = stdio.In
	cmd.Stdout = stdio.Out
	cmd.Stderr = stdio.Err
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	return nil
}
