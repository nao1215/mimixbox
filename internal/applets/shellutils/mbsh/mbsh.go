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
	"strings"

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
// the exit status of the last command (exposed as $?) and a stop flag set by
// the "exit"/"quit" command.
type shell struct {
	lastStatus int
	stop       bool
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
			{Command: "echo one; echo two", Explain: "Run commands in sequence."},
			{Command: "echo hi | wc -c", Explain: "Pipe one command into another."},
			{Command: "echo hi > out.txt", Explain: "Redirect output (>, >>, < are supported)."},
		},
		ExitStatus: "The exit status of the last command executed.\n2  a syntax or usage error in the shell itself.",
		Notes: []string{
			"Tokenizing supports single/double quotes, backslash escapes, $VAR/${VAR}/$? expansion, ~ home expansion, and NAME=value prefixes.",
			"Operators: ; sequences commands, | pipes them, and < > >> redirect I/O. A pipeline's status is its last command's; a list's is its last pipeline's.",
			"Scripting (if/for/while) and background jobs are not yet supported.",
		},
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	Interpret(ctx, stdio, true)
	return nil
}

// Interpret runs the mbsh read-eval loop over stdio until EOF or an "exit"
// command. When prompt is true it prints the interactive prompt before each
// line; shell front-ends such as sh/bash pass false so non-interactive scripts
// produce no prompt noise. It is exported so those launchers can reuse the
// interpreter without re-implementing it.
func Interpret(ctx context.Context, stdio command.IO, prompt bool) {
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
		if prompt {
			_, _ = fmt.Fprint(stdio.Out, sh.prompt())
		}

		line, err := readLine(stdio.In)
		if err != nil {
			// Run whatever was read before EOF (a final line without a
			// trailing newline), then stop the loop cleanly.
			if errors.Is(err, io.EOF) {
				// At EOF the loop ends regardless, so run the final partial line
				// (if any) for its side effects and stop.
				if strings.TrimSpace(line) != "" {
					sh.execInput(ctx, stdio, line)
				}
				return
			}
			_, _ = fmt.Fprintln(stdio.Err, err)
			return
		}

		if stop := sh.execInput(ctx, stdio, line); stop {
			return
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

	// Tokenize (quotes, escapes, $VAR/${VAR}/$?/~ expansion, operators), then
	// parse into a list of pipelines and run it. See parser.go and executor.go.
	toks, err := tokenize(input, sh.lastStatus)
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "mbsh: %v\n", err)
		sh.lastStatus = 2
		return false
	}
	list, err := parse(toks)
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "mbsh: %v\n", err)
		sh.lastStatus = 2
		return false
	}

	sh.lastStatus = sh.execList(ctx, stdio, list)
	return sh.stop
}
