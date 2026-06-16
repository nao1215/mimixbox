// Package stty implements the stty applet: print or change the line settings of
// the terminal on standard input. It covers the everyday workflow - showing
// settings and toggling the common boolean modes such as echo.
package stty

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
	"github.com/nao1215/mimixbox/internal/version"
	"golang.org/x/sys/unix"
)

// Command is the stty applet.
type Command struct{}

// New returns an stty command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "stty" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Print or change terminal line settings" }

// Injected so the print/set logic is testable without a real terminal.
var (
	getTermios = func(fd int) (*unix.Termios, error) { return unix.IoctlGetTermios(fd, unix.TCGETS) }
	setTermios = func(fd int, t *unix.Termios) error { return unix.IoctlSetTermios(fd, unix.TCSETS, t) }
)

// flag describes one boolean terminal mode in the local (lflag) set.
type flag struct {
	name string
	bit  uint32
}

var lflags = []flag{
	{"echo", unix.ECHO},
	{"echoe", unix.ECHOE},
	{"echok", unix.ECHOK},
	{"icanon", unix.ICANON},
	{"isig", unix.ISIG},
	{"iexten", unix.IEXTEN},
}

// Run executes stty.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-a] [SETTING]...", stdio.Err).WithHelp(command.Help{
		Description: "Print or change the terminal line settings on standard input. With no SETTING, " +
			"print a summary; with -a, print all modes. A SETTING enables a boolean mode (e.g. echo) " +
			"or, prefixed with '-', disables it; 'sane' restores sensible defaults.",
		Examples: []command.Example{
			{Command: "stty -a", Explain: "Show all terminal settings."},
			{Command: "stty -echo", Explain: "Turn off input echoing."},
		},
		ExitStatus: "0  success.\n1  standard input is not a terminal or a setting was invalid.",
	})

	// Settings such as "-echo" begin with '-', so options are parsed by hand
	// rather than through the flag set, which would treat them as flags.
	all := false
	var settings []string
	for _, a := range args {
		switch a {
		case "--help", "-h":
			fs.WriteUsage(stdio.Out)
			return nil
		case "--version":
			version.Print(stdio.Out, c.Name())
			return nil
		case "-a", "--all":
			all = true
		default:
			settings = append(settings, a)
		}
	}

	fd, ok := fdOf(stdio.In)
	if !ok {
		_, _ = fmt.Fprintln(stdio.Err, "stty: standard input: not a tty")
		return command.SilentFailure()
	}
	t, err := getTermios(fd)
	if err != nil {
		_, _ = fmt.Fprintln(stdio.Err, "stty: standard input: not a tty")
		return command.SilentFailure()
	}

	if len(settings) == 0 {
		printSettings(stdio.Out, t, all)
		return nil
	}

	if err := applySettings(t, settings); err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "stty: %v\n", err)
		return command.SilentFailure()
	}
	if err := setTermios(fd, t); err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "stty: %v\n", err)
		return command.SilentFailure()
	}
	return nil
}

// applySettings mutates t per the requested settings.
func applySettings(t *unix.Termios, settings []string) error {
	for _, s := range settings {
		switch s {
		case "sane":
			t.Lflag |= unix.ECHO | unix.ECHOE | unix.ECHOK | unix.ICANON | unix.ISIG | unix.IEXTEN
			continue
		case "raw":
			t.Lflag &^= unix.ICANON | unix.ECHO | unix.ISIG | unix.IEXTEN
			continue
		case "cooked":
			t.Lflag |= unix.ICANON | unix.ISIG
			continue
		}
		off := strings.HasPrefix(s, "-")
		name := strings.TrimPrefix(s, "-")
		f, ok := findFlag(name)
		if !ok {
			return fmt.Errorf("invalid argument '%s'", s)
		}
		if off {
			t.Lflag &^= f.bit
		} else {
			t.Lflag |= f.bit
		}
	}
	return nil
}

func findFlag(name string) (flag, bool) {
	for _, f := range lflags {
		if f.name == name {
			return f, true
		}
	}
	return flag{}, false
}

// printSettings writes the current modes. With all, every known flag is shown;
// otherwise only a brief summary.
func printSettings(out io.Writer, t *unix.Termios, all bool) {
	_, _ = fmt.Fprintf(out, "speed %d baud; line = 0;\n", baud(t))
	var parts []string
	for _, f := range lflags {
		on := t.Lflag&f.bit != 0
		switch {
		case all:
			if on {
				parts = append(parts, f.name)
			} else {
				parts = append(parts, "-"+f.name)
			}
		case !on:
			// In the brief view show only modes that differ from the common
			// default (which has these enabled), i.e. the disabled ones.
			parts = append(parts, "-"+f.name)
		}
	}
	if len(parts) > 0 {
		_, _ = fmt.Fprintln(out, strings.Join(parts, " "))
	}
}

// baud maps the termios speed code to a baud number for display.
func baud(t *unix.Termios) int {
	switch t.Ospeed {
	case unix.B9600:
		return 9600
	case unix.B19200:
		return 19200
	case unix.B38400:
		return 38400
	case unix.B57600:
		return 57600
	case unix.B115200:
		return 115200
	default:
		return 38400
	}
}

// fdOf returns the file descriptor of r when it is an *os.File.
func fdOf(r io.Reader) (int, bool) {
	if f, ok := r.(*os.File); ok {
		return int(f.Fd()), true
	}
	return 0, false
}
