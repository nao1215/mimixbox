// Package resize implements the resize applet: print the terminal size as shell
// commands that set the COLUMNS and LINES environment variables, like the
// xterm "resize" utility. Evaluate its output (eval `resize`) to update the
// current shell after the window has changed size.
package resize

import (
	"context"
	"fmt"
	"os"

	"github.com/nao1215/mimixbox/internal/command"
	"golang.org/x/sys/unix"
)

// Command is the resize applet.
type Command struct{}

// New returns a resize command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "resize" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Print commands to set the terminal size" }

// winsize reports the current terminal size as (rows, cols).
//
// Issue #492 asked whether resize should keep probing the process file
// descriptors directly or grow an abstraction around terminal discovery. We
// keep the direct TIOCGWINSZ probe of the std fds on purpose: querying the
// controlling terminal's geometry *is* the job of resize, the size cannot be
// injected through command.IO (whose streams are plain io.Reader/io.Writer and
// may not be terminals), and the seam below already makes the applet fully
// unit-testable. winsize is a package var so a test can replace it with a
// deterministic fake instead of needing a real controlling terminal.
var winsize = func() (rows, cols uint16, err error) {
	for _, fd := range []int{int(os.Stdout.Fd()), int(os.Stderr.Fd()), int(os.Stdin.Fd())} {
		ws, werr := unix.IoctlGetWinsize(fd, unix.TIOCGWINSZ)
		if werr == nil && ws.Row > 0 && ws.Col > 0 {
			return ws.Row, ws.Col, nil
		}
		err = werr
	}
	if err == nil {
		err = fmt.Errorf("could not determine terminal size")
	}
	return 0, 0, err
}

// Run executes resize.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]", stdio.Err)
	sh := fs.BoolP("sh", "u", false, "write Bourne-shell (sh/ksh/bash) commands (default)")
	csh := fs.BoolP("csh", "c", false, "write C-shell (csh/tcsh) commands")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	rows, cols, err := winsize()
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "resize: %v\n", err)
		return command.SilentFailure()
	}

	_, _ = fmt.Fprint(stdio.Out, format(rows, cols, *csh && !*sh))
	return nil
}

// format renders the COLUMNS/LINES assignment for either C-shell (csh) or
// Bourne-shell syntax.
func format(rows, cols uint16, csh bool) string {
	if csh {
		return fmt.Sprintf("set noglob;\nsetenv COLUMNS '%d';\nsetenv LINES '%d';\nunset noglob;\n", cols, rows)
	}
	return fmt.Sprintf("COLUMNS=%d;\nLINES=%d;\nexport COLUMNS LINES;\n", cols, rows)
}
