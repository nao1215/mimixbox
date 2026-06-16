// Package pager implements the more and less applets: page through files (or
// standard input) when standard output is a terminal, and stream straight
// through otherwise so they are safe in pipelines.
package pager

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
	"golang.org/x/sys/unix"
)

// Command is the more or less applet.
type Command struct {
	name   string
	prompt string
}

// NewMore returns the more applet.
func NewMore() *Command { return &Command{name: "more", prompt: "--More--"} }

// NewLess returns the less applet.
func NewLess() *Command { return &Command{name: "less", prompt: ":"} }

// Name returns the command name.
func (c *Command) Name() string { return c.name }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Page through text one screen at a time" }

// Run executes the pager.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[FILE]...", stdio.Err).WithHelp(command.Help{
		Description: "Display FILEs (or standard input) one screen at a time when standard output is " +
			"a terminal. When standard output is not a terminal (a pipe or file), the input is " +
			"copied straight through unchanged.",
		Examples: []command.Example{
			{Command: c.Name() + " file.txt", Explain: "Page through file.txt."},
			{Command: "ls -l | " + c.Name(), Explain: "Page command output on a terminal, or pass it through in a pipe."},
		},
		ExitStatus: "0  success.\n1  an error occurred.",
		Notes: []string{
			"Paging keys: Enter advances one screen, a line starting with q quits. Input is line-buffered, so keys take effect on Enter; raw-mode single-key paging and backward scrolling are not implemented.",
		},
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	readers, cleanup, err := openInputs(stdio, fs.Args())
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "%s: %v\n", c.Name(), err)
		return command.SilentFailure()
	}
	defer cleanup()
	input := io.MultiReader(readers...)

	if !isTerminal(stdio.Out) {
		if _, err := io.Copy(stdio.Out, input); err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "%s: %v\n", c.Name(), err)
			return command.SilentFailure()
		}
		return nil
	}
	return c.page(stdio, input)
}

// openInputs returns a reader per FILE, or standard input when none are given.
func openInputs(stdio command.IO, files []string) ([]io.Reader, func(), error) {
	if len(files) == 0 {
		return []io.Reader{stdio.In}, func() {}, nil
	}
	var readers []io.Reader
	var closers []io.Closer
	cleanup := func() {
		for _, c := range closers {
			_ = c.Close()
		}
	}
	for _, name := range files {
		if name == "-" {
			readers = append(readers, stdio.In)
			continue
		}
		f, err := os.Open(name) //nolint:gosec // user-named file
		if err != nil {
			cleanup()
			return nil, nil, err
		}
		readers = append(readers, f)
		closers = append(closers, f)
	}
	return readers, cleanup, nil
}

// page shows input one screenful at a time, reading control keys from the
// terminal.
func (c *Command) page(stdio command.IO, input io.Reader) error {
	rows := terminalRows(stdio.Out)
	pageSize := rows - 1
	if pageSize < 1 {
		pageSize = 1
	}

	tty, keys := openTTY(stdio.In)
	if tty != nil {
		defer func() { _ = tty.Close() }()
	}
	keyReader := bufio.NewReader(keys)

	// A bufio.Reader (rather than bufio.Scanner) has no per-line length cap, so a
	// very long line cannot break paging the way it would the passthrough path.
	r := bufio.NewReader(input)
	shown := 0
	for {
		line, err := r.ReadString('\n')
		if line != "" {
			_, _ = fmt.Fprint(stdio.Out, line)
			shown++
			if shown >= pageSize {
				_, _ = fmt.Fprint(stdio.Err, c.prompt)
				key, _ := keyReader.ReadString('\n')
				if strings.HasPrefix(strings.TrimSpace(key), "q") {
					return nil
				}
				shown = 0
			}
		}
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
	}
}

// openTTY returns the terminal to read control keys from: /dev/tty when it can
// be opened, otherwise the provided standard input.
func openTTY(stdin io.Reader) (*os.File, io.Reader) {
	if f, err := os.Open("/dev/tty"); err == nil {
		return f, f
	}
	return nil, stdin
}

// isTerminal reports whether w is a real terminal. It queries the terminal
// attributes, which succeeds only on a tty - so non-interactive character
// devices such as /dev/null are correctly treated as not a terminal.
func isTerminal(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	_, err := unix.IoctlGetTermios(int(f.Fd()), unix.TCGETS)
	return err == nil
}

// terminalRows returns the height of w's terminal, defaulting to 24.
func terminalRows(w io.Writer) int {
	f, ok := w.(*os.File)
	if !ok {
		return 24
	}
	if ws, err := unix.IoctlGetWinsize(int(f.Fd()), unix.TIOCGWINSZ); err == nil && ws.Row > 0 {
		return int(ws.Row)
	}
	return 24
}
