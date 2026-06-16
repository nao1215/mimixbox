package pager

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
	"golang.org/x/sys/unix"
)

// core is the shared paging engine behind the more and less front-ends. It owns
// input handling, terminal detection, scrolling, rendering and key input; the
// only front-end-specific behavior is carried in cfg.
type core struct {
	cfg config
}

// run opens the inputs and either pages them (on a terminal) or copies them
// straight through (in a pipe). files is the operand list; with none, standard
// input is used.
func (e core) run(stdio command.IO, files []string) error {
	readers, cleanup, err := openInputs(stdio, files)
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "%s: %v\n", e.cfg.name, err)
		return command.SilentFailure()
	}
	defer cleanup()
	input := io.MultiReader(readers...)

	if !isTerminal(stdio.Out) {
		if _, err := io.Copy(stdio.Out, input); err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "%s: %v\n", e.cfg.name, err)
			return command.SilentFailure()
		}
		return nil
	}
	return e.page(stdio, input)
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
func (e core) page(stdio command.IO, input io.Reader) error {
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
				_, _ = fmt.Fprint(stdio.Err, e.cfg.prompt)
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
