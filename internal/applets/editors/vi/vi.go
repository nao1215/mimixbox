// Package vi implements a small vi-style modal text editor. It is intentionally
// minimal: it supports the everyday motions (h/j/k/l, 0, $, gg, G), edits
// (x, dd, i, a, A, o, O) and the ex commands :w, :q, :q!, :wq and ZZ.
//
// When standard input is a terminal it runs interactively with a redrawn
// screen; when input is a pipe or file (as in tests) it consumes the input as a
// keystroke script and writes the file on :w/:wq/ZZ, which keeps the whole
// editor automatically testable.
package vi

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/nao1215/mimixbox/internal/command"
	"golang.org/x/sys/unix"
)

// Command is the vi applet.
type Command struct{}

// New returns a vi command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "vi" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "A minimal vi-style screen text editor" }

// Run executes vi.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[FILE]", stdio.Err).WithHelp(command.Help{
		Description: "A minimal vi-style screen editor. It opens FILE (or an empty buffer) in a " +
			"full-screen terminal. Like vi it starts in normal mode; press i to insert text and " +
			"Esc to return to normal mode.",
		Examples: []command.Example{
			{Command: "vi notes.txt", Explain: "Edit notes.txt (created on save if missing)."},
			{Command: "vi", Explain: "Open an empty buffer."},
		},
		Notes: []string{
			"Motions: h j k l, w b e (word), 0 $ (line), gg G (file); a count repeats them (3j, 2w).",
			"Edits: i a A o O insert, x delete char, dd delete line, yy/p/P yank & paste, u undo (counts apply, e.g. 2x, 3dd).",
			"Search: /pattern and ?pattern, then n / N for the next/previous match. Ex commands: :w :q :q! :wq and ZZ.",
		},
		ExitStatus: "0  the file was edited and written without error.\n1  an error occurred.",
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	filename := ""
	if rest := fs.Args(); len(rest) > 0 {
		filename = rest[0]
	}

	content := ""
	if filename != "" {
		if data, rerr := os.ReadFile(filename); rerr == nil { //nolint:gosec // user-named file
			content = string(data)
		} else if !os.IsNotExist(rerr) {
			_, _ = fmt.Fprintf(stdio.Err, "vi: %s\n", command.FileError(filename, rerr))
			return command.SilentFailure()
		}
	}

	ed := newEditor(filename, content)

	var runErr error
	if isTerminal(stdio.In) {
		runErr = runInteractive(stdio, ed)
	} else {
		runErr = runBatch(stdio, ed)
	}
	if runErr != nil {
		_, _ = fmt.Fprintf(stdio.Err, "vi: %v\n", runErr)
		return command.SilentFailure()
	}

	if ed.save {
		if filename == "" {
			_, _ = fmt.Fprintln(stdio.Err, "vi: no file name")
			return command.SilentFailure()
		}
		if err := os.WriteFile(filename, []byte(ed.content()), 0o644); err != nil { //nolint:gosec // simple default mode
			_, _ = fmt.Fprintf(stdio.Err, "vi: %v\n", err)
			return command.SilentFailure()
		}
	}
	return nil
}

// runBatch feeds all input bytes to the editor as a keystroke script. It
// returns any error encountered while reading standard input so the caller can
// surface an unreadable stdin instead of silently editing an empty buffer.
func runBatch(stdio command.IO, ed *editor) error {
	data, err := io.ReadAll(stdio.In)
	if err != nil {
		return err
	}
	ed.feedString(string(data))
	return nil
}

// isTerminal reports whether r is a character device (a real terminal).
func isTerminal(r io.Reader) bool {
	f, ok := r.(*os.File)
	if !ok {
		return false
	}
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

// runInteractive drives the editor in raw mode, redrawing the screen after each
// keystroke until the editor asks to quit. When standard input is not a real
// terminal it falls back to batch mode and propagates any read error.
func runInteractive(stdio command.IO, ed *editor) error {
	f, ok := stdio.In.(*os.File)
	if !ok {
		return runBatch(stdio, ed)
	}
	fd := int(f.Fd())
	old, err := unix.IoctlGetTermios(fd, unix.TCGETS)
	if err != nil {
		return runBatch(stdio, ed)
	}
	raw := *old
	raw.Lflag &^= unix.ECHO | unix.ICANON | unix.ISIG | unix.IEXTEN
	raw.Iflag &^= unix.IXON | unix.ICRNL | unix.BRKINT | unix.INPCK | unix.ISTRIP
	if err := unix.IoctlSetTermios(fd, unix.TCSETS, &raw); err != nil {
		return runBatch(stdio, ed)
	}
	defer func() { _ = unix.IoctlSetTermios(fd, unix.TCSETS, old) }()

	buf := make([]byte, 1)
	for !ed.quit {
		redraw(stdio.Out, ed)
		n, err := f.Read(buf)
		if err != nil || n == 0 {
			return nil
		}
		ed.feed(buf[0])
	}
	ed.flush()
	redraw(stdio.Out, ed)
	return nil
}

// redraw clears the screen and paints the buffer, a status line and the cursor.
func redraw(w io.Writer, ed *editor) {
	var b []byte
	b = append(b, "\x1b[2J\x1b[H"...) // clear, home
	for _, line := range ed.lines {
		b = append(b, line...)
		b = append(b, "\r\n"...)
	}
	status := fmt.Sprintf("\x1b[7m %s%s mode=%s \x1b[0m", ed.filename, dirtyMark(ed), modeName(ed.mode))
	if ed.mode == modeCommand {
		status = ":" + ed.cmdline
	}
	if ed.mode == modeSearch {
		prefix := "/"
		if !ed.searchForward {
			prefix = "?"
		}
		status = prefix + ed.searchPat
	}
	if ed.message != "" {
		status = ed.message
	}
	b = append(b, status...)
	// Position the cursor (1-based).
	b = append(b, fmt.Sprintf("\x1b[%d;%dH", ed.cy+1, ed.cx+1)...)
	_, _ = w.Write(b)
}

func dirtyMark(ed *editor) string {
	if ed.dirty {
		return " [+]"
	}
	return ""
}

func modeName(m mode) string {
	switch m {
	case modeInsert:
		return "INSERT"
	case modeCommand:
		return "COMMAND"
	case modeSearch:
		return "SEARCH"
	default:
		return "NORMAL"
	}
}
