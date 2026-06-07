// Package printf implements the printf applet: format and print data the way
// GNU printf does. Like echo, printf is a POSIX utility that does not use
// getopt-style option parsing: the leading FORMAT operand may itself look like
// an option (for example "%-5s"), so the package never builds a command.FlagSet
// and treats args[0] as the format and args[1:] as the conversion arguments.
package printf

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the printf applet.
type Command struct{}

// New returns a printf command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "printf" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Formats and print data" }

// Run executes printf. args[0] is the FORMAT string and args[1:] are the
// arguments consumed by the conversion specifications. When the format consumes
// fewer arguments than are supplied, the format is reused until the arguments
// are exhausted, matching GNU printf.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	if len(args) == 0 {
		_, _ = fmt.Fprintln(stdio.Err, "printf: missing operand")
		return command.SilentFailure()
	}

	format := args[0]
	operands := args[1:]

	var b strings.Builder
	idx := 0
	for {
		consumed := format2(&b, format, operands[idx:])
		idx += consumed
		// Stop once every operand has been consumed. Reuse the format only
		// when the last pass actually consumed at least one operand; a format
		// with no conversions consumes zero and must run exactly once.
		if idx >= len(operands) || consumed == 0 {
			break
		}
	}

	if _, err := fmt.Fprint(stdio.Out, b.String()); err != nil {
		return command.Failure(err)
	}
	return nil
}

// format2 expands one pass of format, writing the result to b and pulling
// arguments from operands as conversion specifications require them. It returns
// the number of operands consumed. Missing operands are treated as the empty
// string / zero.
func format2(b *strings.Builder, format string, operands []string) int {
	argi := 0
	next := func() string {
		if argi < len(operands) {
			s := operands[argi]
			argi++
			return s
		}
		argi++
		return ""
	}

	for i := 0; i < len(format); i++ {
		ch := format[i]
		switch ch {
		case '\\':
			consumed := formatEscape(b, format[i+1:])
			if consumed == 0 {
				b.WriteByte('\\')
			} else {
				i += consumed
			}
		case '%':
			consumed := conversion(b, format[i:], next)
			i += consumed
		default:
			b.WriteByte(ch)
		}
	}
	// Even a format with no conversion still "consumes" nothing, but report the
	// real number requested so the caller can decide whether to reuse it.
	if argi > len(operands) {
		return len(operands)
	}
	return argi
}

// formatEscape interprets a backslash escape in the FORMAT string. s is the text
// after the backslash. It returns the number of bytes consumed from s (0 when
// the escape is not recognized, so the caller can emit a literal backslash).
func formatEscape(b *strings.Builder, s string) int {
	if len(s) == 0 {
		return 0
	}
	switch s[0] {
	case 'a':
		b.WriteByte('\a')
		return 1
	case 'b':
		b.WriteByte('\b')
		return 1
	case 'f':
		b.WriteByte('\f')
		return 1
	case 'n':
		b.WriteByte('\n')
		return 1
	case 'r':
		b.WriteByte('\r')
		return 1
	case 't':
		b.WriteByte('\t')
		return 1
	case 'v':
		b.WriteByte('\v')
		return 1
	case '\\':
		b.WriteByte('\\')
		return 1
	case '0':
		n, consumed := octal(s[1:])
		b.WriteByte(n)
		return 1 + consumed
	case 'x':
		n, consumed := hex(s[1:])
		if consumed == 0 {
			return 0
		}
		b.WriteByte(n)
		return 1 + consumed
	default:
		return 0
	}
}

// conversion handles a single conversion specification beginning at s[0] == '%'.
// next supplies the argument when the specification needs one. It returns the
// number of bytes after the '%' that were consumed.
func conversion(b *strings.Builder, s string, next func() string) int {
	// s[0] is '%'. Scan flags, width and precision (digits, -, +, space, #, 0,
	// and a single '.') up to the verb so we can pass them through to fmt.
	end := 1
	for end < len(s) {
		c := s[end]
		if strings.IndexByte("-+ #0.", c) >= 0 || (c >= '0' && c <= '9') {
			end++
			continue
		}
		break
	}
	if end >= len(s) {
		// Trailing '%' with no verb: emit literally.
		b.WriteString(s)
		return len(s) - 1
	}

	verb := s[end]
	spec := s[:end] // flags/width/precision without the verb

	switch verb {
	case '%':
		b.WriteByte('%')
		return end
	case 's':
		_, _ = fmt.Fprintf(b, spec+"s", next())
		return end
	case 'b':
		// %b: like %s but interpret backslash escapes in the argument.
		expanded, _ := expandEscapes(next())
		b.WriteString(expanded)
		return end
	case 'c':
		arg := next()
		if arg == "" {
			return end
		}
		b.WriteByte(arg[0])
		return end
	case 'd', 'i':
		_, _ = fmt.Fprintf(b, spec+"d", toInt(next()))
		return end
	case 'u':
		_, _ = fmt.Fprintf(b, spec+"d", toUint(next()))
		return end
	case 'o':
		_, _ = fmt.Fprintf(b, spec+"o", toUint(next()))
		return end
	case 'x':
		_, _ = fmt.Fprintf(b, spec+"x", toUint(next()))
		return end
	case 'X':
		_, _ = fmt.Fprintf(b, spec+"X", toUint(next()))
		return end
	default:
		// Unknown verb: emit the specification literally.
		b.WriteString(s[:end+1])
		return end
	}
}

// toInt converts a printf argument to a signed integer, defaulting to 0 when the
// argument is empty or not a number (GNU printf warns but still prints 0).
func toInt(s string) int64 {
	if s == "" {
		return 0
	}
	v, err := strconv.ParseInt(s, 0, 64)
	if err != nil {
		return 0
	}
	return v
}

// toUint converts a printf argument to an unsigned integer for %u/%o/%x/%X.
func toUint(s string) uint64 {
	if s == "" {
		return 0
	}
	if v, err := strconv.ParseUint(s, 0, 64); err == nil {
		return v
	}
	if v, err := strconv.ParseInt(s, 0, 64); err == nil {
		return uint64(v)
	}
	return 0
}

// expandEscapes interprets the backslash escapes that printf's %b conversion
// understands. The returned bool is true when a \c escape was seen.
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
