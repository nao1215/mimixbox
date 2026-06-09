// Package mbsh implements the Mimix Box Shell applet: a small interactive REPL
// that reads a command line, runs it (a built-in such as cd, an external
// program, or a MimixBox applet), and loops until end-of-input or the exit
// command.
//
// Its defining design choice is that it reads from and writes to the injected
// command.IO streams rather than the process standard streams, so a test can
// drive the whole REPL by feeding a script through stdio.In and inspecting
// stdio.Out/stdio.Err.
package mbsh

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/nao1215/mimixbox/internal/applets/shellutils/mbsh/builtin"
	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the mbsh applet.
type Command struct{}

// New returns an mbsh command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "mbsh" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Mimix Box Shell" }

// shell holds the small amount of mutable state the REPL keeps between lines:
// the exit status of the last command (exposed as $?).
type shell struct {
	lastStatus int
}

// Run starts the read-eval-print loop. It parses its own flags (only the
// standard --help/--version), then repeatedly prints the prompt, reads a line
// from stdio.In, and executes it. The loop ends, returning nil, when the reader
// reaches EOF or the user types "exit".
func (c *Command) Run(ctx context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]", stdio.Err).WithHelp(command.Help{
		Description: "MimixBox Shell: a minimal interactive shell. It reads a command and its " +
			"arguments from each line and runs the matching MimixBox applet. With no terminal " +
			"it reads commands from standard input, so it can run a script piped on stdin.",
		Examples: []command.Example{
			{Command: "mbsh", Explain: "Start an interactive prompt; type 'exit' or Ctrl-D to quit."},
			{Command: "echo 'echo hello' | mbsh", Explain: "Run commands fed on standard input."},
		},
		Notes: []string{
			"Whitespace splits tokens; quoting, parameter expansion, pipelines, and redirections are not yet supported.",
			"Each line runs one applet; there is no scripting (if/for/while) or variable assignment.",
		},
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	sh := &shell{}
	// Read commands one byte at a time rather than through a buffered reader.
	// A buffered reader would read past the current line, so a command launched
	// below (which shares stdio.In) would see EOF instead of the bytes still
	// sitting in the shell's buffer. Reading byte by byte leaves stdio.In
	// positioned exactly after the line, which is what POSIX shells do when
	// their input is a non-seekable stream. In script mode this means a
	// stdin-consuming command (cat, sed, ...) reads the rest of the script, just
	// as it would under any other shell; in interactive mode each Enter yields
	// one line and the foreground command shares the terminal.
	for {
		_, _ = fmt.Fprint(stdio.Out, sh.prompt())

		line, err := readLine(stdio.In)
		if err != nil {
			// Run whatever was read before EOF (a final line without a
			// trailing newline), then stop the loop cleanly.
			if errors.Is(err, io.EOF) {
				if strings.TrimSpace(line) != "" {
					if stop := sh.execInput(ctx, stdio, line); stop {
						return nil
					}
				}
				return nil
			}
			_, _ = fmt.Fprintln(stdio.Err, err)
			return nil
		}

		if stop := sh.execInput(ctx, stdio, line); stop {
			return nil
		}
	}
}

// readLine reads a single line (including the trailing newline) from r without
// reading any further, so r is left positioned exactly after the line. The
// returned error is io.EOF when the stream ends; any bytes read before EOF are
// still returned so a final line without a newline is executed.
func readLine(r io.Reader) (string, error) {
	var b strings.Builder
	buf := make([]byte, 1)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			b.WriteByte(buf[0])
			if buf[0] == '\n' {
				return b.String(), nil
			}
		}
		if err != nil {
			return b.String(), err
		}
	}
}

// prompt renders the prompt, showing the current working directory so the user
// always knows where they are: "mbsh:/path> ".
func (sh *shell) prompt() string {
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "?"
	}
	return fmt.Sprintf("mbsh:%s> ", cwd)
}

// execInput parses and runs a single input line. It returns true when the shell
// should stop (the user typed "exit"). A failure to run the command is reported
// on stdio.Err but does not stop the loop; the command's exit status is recorded
// in sh.lastStatus so the next line can read it as $?.
func (sh *shell) execInput(ctx context.Context, stdio command.IO, input string) (stop bool) {
	input = strings.TrimSuffix(input, "\n")
	input = strings.TrimSuffix(input, "\r")

	// A line whose first non-blank character is '#' is a comment.
	trimmed := strings.TrimSpace(input)
	if trimmed == "" || strings.HasPrefix(trimmed, "#") {
		return false
	}

	// Expand $? to the previous command's exit status before splitting.
	input = strings.ReplaceAll(input, "$?", strconv.Itoa(sh.lastStatus))

	args := strings.Fields(input)
	if len(args) == 0 {
		return false
	}

	switch args[0] {
	case "exit", "quit":
		return true
	}

	// A built-in is preferred over an external command of the same name.
	if builtin.IsBuiltinCmd(args[0]) {
		if err := builtin.Run(stdio, args[0], args[1:]); err != nil {
			_, _ = fmt.Fprintln(stdio.Err, err)
			sh.lastStatus = 1
		} else {
			sh.lastStatus = 0
		}
		return false
	}

	sh.lastStatus = sh.runExternal(ctx, stdio, args)
	return false
}

// runExternal runs args[0] as an external program, falling back to running it as
// a MimixBox applet (by re-executing this binary) when it is not on the PATH.
// It returns the command's exit status.
func (sh *shell) runExternal(ctx context.Context, stdio command.IO, args []string) int {
	name := args[0]
	if _, err := exec.LookPath(name); err != nil {
		if self, e := os.Executable(); e == nil {
			return run(ctx, stdio, self, append([]string{name}, args[1:]...))
		}
		_, _ = fmt.Fprintf(stdio.Err, "mbsh: %s: command not found\n", name)
		return 127
	}
	return run(ctx, stdio, name, args[1:])
}

// run executes name with argv wired to the shell streams and returns its exit
// status (127 when it cannot start).
func run(ctx context.Context, stdio command.IO, name string, argv []string) int {
	cmd := exec.CommandContext(ctx, name, argv...) //nolint:gosec // running a user-typed command is the whole point of a shell
	cmd.Stdin = stdio.In
	cmd.Stdout = stdio.Out
	cmd.Stderr = stdio.Err
	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return exitErr.ExitCode()
		}
		_, _ = fmt.Fprintf(stdio.Err, "mbsh: %v\n", err)
		return 127
	}
	return 0
}
