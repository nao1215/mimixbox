// Package mbsh implements the Mimix Box Shell applet: a small interactive REPL
// that reads a command line, runs it (a built-in such as cd, otherwise an
// external program), and loops until end-of-input or the exit command.
//
// The shell is still in development. Its defining design choice is that it reads
// from and writes to the injected command.IO streams rather than the process
// standard streams, so a test can drive the whole REPL by feeding a script
// through stdio.In and inspecting stdio.Out/stdio.Err.
package mbsh

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/nao1215/mimixbox/internal/applets/shellutils/mbsh/builtin"
	"github.com/nao1215/mimixbox/internal/command"
)

// prompt is written to stdio.Out before each line is read.
const prompt = "> "

// Command is the mbsh applet.
type Command struct{}

// New returns an mbsh command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "mbsh" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Mimix Box Shell" }

// Run starts the read-eval-print loop. It parses its own flags (only the
// standard --help/--version), then repeatedly prints the prompt, reads a line
// from stdio.In, and executes it. The loop ends, returning nil, when the reader
// reaches EOF or the user types "exit".
func (c *Command) Run(ctx context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]", stdio.Err)
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	reader := bufio.NewReader(stdio.In)
	for {
		fmt.Fprint(stdio.Out, prompt)

		line, err := reader.ReadString('\n')
		if err != nil {
			// Run whatever was read before EOF (a final line without a
			// trailing newline), then stop the loop cleanly.
			if errors.Is(err, io.EOF) {
				if strings.TrimSpace(line) != "" {
					if stop := execInput(ctx, stdio, line); stop {
						return nil
					}
				}
				return nil
			}
			fmt.Fprintln(stdio.Err, err)
			return nil
		}

		if stop := execInput(ctx, stdio, line); stop {
			return nil
		}
	}
}

// execInput parses and runs a single input line. It returns true when the shell
// should stop (the user typed "exit"). A failure to run the command is reported
// on stdio.Err but does not stop the loop.
func execInput(ctx context.Context, stdio command.IO, input string) (stop bool) {
	input = strings.TrimSuffix(input, "\n")
	input = strings.TrimSuffix(input, "\r")

	args := strings.Fields(input)
	if len(args) == 0 {
		return false
	}

	if args[0] == "exit" {
		return true
	}

	// A built-in is preferred over an external command of the same name.
	if builtin.IsBuiltinCmd(args[0]) {
		if err := builtin.Run(stdio, args[0], args[1:]); err != nil {
			fmt.Fprintln(stdio.Err, err)
		}
		return false
	}

	// Run an external command, wiring its streams to the shell's streams so
	// the test (and the user) sees the command's output on stdio.Out/Err.
	cmd := exec.CommandContext(ctx, args[0], args[1:]...) //nolint:gosec // running a user-typed command is the whole point of a shell
	cmd.Stdin = stdio.In
	cmd.Stdout = stdio.Out
	cmd.Stderr = stdio.Err
	if err := cmd.Run(); err != nil {
		fmt.Fprintln(stdio.Err, err)
	}
	return false
}
