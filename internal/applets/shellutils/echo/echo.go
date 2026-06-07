// Package echo implements the echo applet: write its arguments to standard
// output. Like the GNU shell builtin, echo does not use getopt parsing: it only
// recognizes a leading run of -n, -e and -E flags and treats everything else,
// including --help and --version, as literal text.
package echo

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the echo applet.
type Command struct{}

// New returns an echo command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "echo" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Display a line of text" }

// Run executes echo.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	noNewline := false
	interpret := false

	i := 0
	for ; i < len(args); i++ {
		if !isFlag(args[i]) {
			break
		}
		for _, ch := range args[i][1:] {
			switch ch {
			case 'n':
				noNewline = true
			case 'e':
				interpret = true
			case 'E':
				interpret = false
			}
		}
	}

	text := strings.Join(args[i:], " ")
	if interpret {
		var stop bool
		text, stop = expandEscapes(text)
		if stop {
			noNewline = true
		}
	}

	if _, err := fmt.Fprint(stdio.Out, text); err != nil {
		return command.Failure(err)
	}
	if !noNewline {
		if _, err := fmt.Fprintln(stdio.Out); err != nil {
			return command.Failure(err)
		}
	}
	return nil
}

// isFlag reports whether s is a run of echo's own flags (-n, -e, -E). Any other
// token, including "-" or "--help", ends flag processing.
func isFlag(s string) bool {
	if len(s) < 2 || s[0] != '-' {
		return false
	}
	for _, ch := range s[1:] {
		if ch != 'n' && ch != 'e' && ch != 'E' {
			return false
		}
	}
	return true
}

// expandEscapes interprets the backslash escapes that echo -e understands. The
// returned bool is true when a \c escape was seen, which suppresses all further
// output including the trailing newline.
func expandEscapes(s string) (string, bool) {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] != '\\' || i+1 >= len(s) {
			b.WriteByte(s[i])
			continue
		}
		i++
		switch s[i] {
		case 'a':
			b.WriteByte('\a')
		case 'b':
			b.WriteByte('\b')
		case 'c':
			return b.String(), true
		case 'f':
			b.WriteByte('\f')
		case 'n':
			b.WriteByte('\n')
		case 'r':
			b.WriteByte('\r')
		case 't':
			b.WriteByte('\t')
		case 'v':
			b.WriteByte('\v')
		case '\\':
			b.WriteByte('\\')
		case '0':
			n, consumed := octal(s[i+1:])
			b.WriteByte(n)
			i += consumed
		case 'x':
			n, consumed := hex(s[i+1:])
			if consumed == 0 {
				b.WriteByte('\\')
				b.WriteByte('x')
			} else {
				b.WriteByte(n)
				i += consumed
			}
		default:
			b.WriteByte('\\')
			b.WriteByte(s[i])
		}
	}
	return b.String(), false
}

func octal(s string) (byte, int) {
	end := 0
	for end < len(s) && end < 3 && s[end] >= '0' && s[end] <= '7' {
		end++
	}
	if end == 0 {
		return 0, 0
	}
	v, _ := strconv.ParseInt(s[:end], 8, 16)
	return byte(v), end
}

func hex(s string) (byte, int) {
	end := 0
	for end < len(s) && end < 2 && isHex(s[end]) {
		end++
	}
	if end == 0 {
		return 0, 0
	}
	v, _ := strconv.ParseInt(s[:end], 16, 16)
	return byte(v), end
}

func isHex(b byte) bool {
	return (b >= '0' && b <= '9') || (b >= 'a' && b <= 'f') || (b >= 'A' && b <= 'F')
}
