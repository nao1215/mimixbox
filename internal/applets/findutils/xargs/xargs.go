// Package xargs implements the xargs applet: read items from standard input and
// run a command with those items appended as arguments. It covers the common
// switches (-n, -I, -0, -d, -t, -r) used to glue command pipelines together.
package xargs

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"

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
	replace    string // -I: replace this token with each input item
	null       bool   // -0: input items are NUL-separated
	delim      string // -d: custom input delimiter
	verbose    bool   // -t: print each command before running it
	noRunEmpty bool   // -r: do not run the command if input is empty
}

// Run executes xargs.
func (c *Command) Run(ctx context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... [COMMAND [INITIAL-ARGS]...]", stdio.Err)
	maxArgs := fs.IntP("max-args", "n", 0, "use at most MAX-ARGS arguments per command line")
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
		maxArgs: *maxArgs, replace: *replace, null: *null,
		delim: *delim, verbose: *verbose, noRunEmpty: *noRunEmpty,
	}

	cmdArgs := fs.Args()
	cmdName := "echo"
	var initial []string
	if len(cmdArgs) > 0 {
		cmdName = cmdArgs[0]
		initial = cmdArgs[1:]
	}

	items, err := readItems(stdio, opts)
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "xargs: %v\n", err)
		return command.SilentFailure()
	}

	if len(items) == 0 && opts.noRunEmpty {
		return nil
	}

	if err := runCommands(ctx, stdio, cmdName, initial, items, opts); err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "xargs: %v\n", err)
		return command.SilentFailure()
	}
	return nil
}

// readItems reads and splits the input into items according to the -0, -d and
// default-whitespace rules.
func readItems(stdio command.IO, opts options) ([]string, error) {
	scanner := bufio.NewScanner(stdio.In)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)

	switch {
	case opts.null:
		scanner.Split(splitOn(0))
	case opts.delim != "":
		scanner.Split(splitOn(opts.delim[0]))
	default:
		scanner.Split(bufio.ScanWords)
	}

	var items []string
	for scanner.Scan() {
		tok := scanner.Text()
		if opts.null || opts.delim != "" {
			if tok == "" {
				continue
			}
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
func runCommands(ctx context.Context, stdio command.IO, name string, initial, items []string, opts options) error {
	// -I replaces a token and runs once per input item.
	if opts.replace != "" {
		for _, item := range items {
			argv := make([]string, len(initial))
			for i, a := range initial {
				argv[i] = strings.ReplaceAll(a, opts.replace, item)
			}
			if err := exec1(ctx, stdio, name, argv, opts); err != nil {
				return err
			}
		}
		return nil
	}

	batches := batch(items, opts.maxArgs)
	if len(batches) == 0 {
		// No input: run the command once with just the initial args (GNU xargs
		// behavior unless -r was given, which is handled by the caller).
		return exec1(ctx, stdio, name, initial, opts)
	}
	for _, b := range batches {
		argv := append(append([]string{}, initial...), b...)
		if err := exec1(ctx, stdio, name, argv, opts); err != nil {
			return err
		}
	}
	return nil
}

// batch splits items into groups of at most n (n<=0 means a single group).
func batch(items []string, n int) [][]string {
	if len(items) == 0 {
		return nil
	}
	if n <= 0 {
		return [][]string{items}
	}
	var out [][]string
	for i := 0; i < len(items); i += n {
		end := i + n
		if end > len(items) {
			end = len(items)
		}
		out = append(out, items[i:end])
	}
	return out
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
