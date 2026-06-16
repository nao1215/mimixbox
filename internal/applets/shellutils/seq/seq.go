// Package seq implements the seq applet: print a column of numbers from FIRST
// to LAST, optionally stepping by INCREMENT, matching GNU seq semantics.
package seq

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the seq applet.
type Command struct{}

// New returns a seq command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "seq" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Print a column of numbers" }

// Run executes seq.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... LAST | FIRST LAST | FIRST INCREMENT LAST", stdio.Err).WithHelp(command.Help{
		Description: "Print a column of numbers from FIRST to LAST, stepping by INCREMENT. FIRST and " +
			"INCREMENT default to 1 when omitted, and a negative INCREMENT counts down. By default " +
			"numbers are separated by a newline; -s sets a different separator, -w pads with " +
			"leading zeroes so all numbers share the same width, and -f gives a printf-style " +
			"floating-point format.",
		Examples: []command.Example{
			{Command: "seq 5", Explain: "Print 1 through 5, one per line."},
			{Command: "seq 2 2 10", Explain: "Print the even numbers from 2 to 10."},
			{Command: "seq -s , 1 5", Explain: "Print 1,2,3,4,5 separated by commas."},
			{Command: "seq -w 1 10", Explain: "Print 01 through 10, zero-padded to equal width."},
		},
		ExitStatus: "0  the sequence was printed successfully.\n1  an operand was missing or not a valid number, or the increment was zero.",
	})
	separator := fs.StringP("separator", "s", "\n", "use STRING to separate numbers (default \\n)")
	equalWidth := fs.BoolP("equal-width", "w", false, "equalize width by padding with leading zeroes")
	format := fs.StringP("format", "f", "", "use printf style floating-point FORMAT")

	proceed, err := fs.Parse(stdio, escapeNegativeNumbers(args))
	if err != nil || !proceed {
		return err
	}

	operands := unescapeNegativeNumbers(fs.Args())
	if len(operands) == 0 || len(operands) > 3 {
		_, _ = fmt.Fprintf(stdio.Err, "seq: missing operand\n")
		_, _ = fmt.Fprintf(stdio.Err, "Try 'seq --help' for more information.\n")
		return command.SilentFailure()
	}

	// Defaults: FIRST and INCREMENT are 1.
	first, increment, last := 1.0, 1.0, 0.0
	firstStr, incStr, lastStr := "1", "1", ""

	var perr error
	switch len(operands) {
	case 1:
		last, lastStr, perr = parse(operands[0])
	case 2:
		if first, firstStr, perr = parse(operands[0]); perr == nil {
			last, lastStr, perr = parse(operands[1])
		}
	case 3:
		if first, firstStr, perr = parse(operands[0]); perr == nil {
			if increment, incStr, perr = parse(operands[1]); perr == nil {
				last, lastStr, perr = parse(operands[2])
			}
		}
	}
	if perr != nil {
		_, _ = fmt.Fprintf(stdio.Err, "seq: invalid floating point argument: '%s'\n", perr.Error())
		return command.SilentFailure()
	}

	if increment == 0 {
		_, _ = fmt.Fprintf(stdio.Err, "seq: invalid Zero increment value: '%s'\n", incStr)
		return command.SilentFailure()
	}

	values := generate(first, increment, last)
	out := render(values, *separator, *equalWidth, *format, firstStr, incStr, lastStr)
	_, werr := fmt.Fprint(stdio.Out, out)
	if werr != nil {
		return command.Failure(werr)
	}
	return nil
}

// negSentinel marks an escaped negative-number operand. It is a control byte
// that cannot occur in a real argument, so the round-trip is unambiguous.
const negSentinel = "\x00neg\x00"

// escapeNegativeNumbers rewrites a negative-number operand such as "-1" so the
// flag parser does not treat it as a short option, while preserving argument
// order. GNU seq accepts negative numbers as FIRST/INCREMENT/LAST operands.
// A token that is the value of -s/-f (attached or as the following argument) is
// left untouched so a negative separator or format string still works.
func escapeNegativeNumbers(args []string) []string {
	out := make([]string, 0, len(args))
	valueTakingShort := map[byte]bool{'s': true, 'f': true}
	afterTerminator := false

	for i := 0; i < len(args); i++ {
		a := args[i]
		if afterTerminator {
			out = append(out, escapeIfNegative(a))
			continue
		}
		switch {
		case a == "--":
			afterTerminator = true
			out = append(out, a)
		case isNegativeNumber(a):
			out = append(out, negSentinel+a)
		case strings.HasPrefix(a, "--"):
			out = append(out, a)
			if !strings.Contains(a, "=") && takesValueLong(a) && i+1 < len(args) {
				out = append(out, args[i+1])
				i++
			}
		case strings.HasPrefix(a, "-") && len(a) > 1:
			out = append(out, a)
			if last := a[len(a)-1]; valueTakingShort[last] && i+1 < len(args) {
				out = append(out, args[i+1])
				i++
			}
		default:
			out = append(out, a)
		}
	}
	return out
}

// escapeIfNegative escapes a negative-number operand regardless of position,
// used for tokens that follow an explicit "--" terminator.
func escapeIfNegative(a string) string {
	if isNegativeNumber(a) {
		return negSentinel + a
	}
	return a
}

// unescapeNegativeNumbers reverses escapeNegativeNumbers on the parsed operands.
func unescapeNegativeNumbers(operands []string) []string {
	out := make([]string, len(operands))
	for i, a := range operands {
		out[i] = strings.TrimPrefix(a, negSentinel)
	}
	return out
}

// takesValueLong reports whether a long option consumes the following argument
// as its value (only --separator and --format do here).
func takesValueLong(a string) bool {
	return a == "--separator" || a == "--format"
}

// isNegativeNumber reports whether s is a minus sign followed by a numeric
// literal, i.e. a negative-number operand rather than an option.
func isNegativeNumber(s string) bool {
	if !strings.HasPrefix(s, "-") || len(s) < 2 {
		return false
	}
	_, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	return err == nil
}

// parseError carries the original operand text so the caller can report it.
type parseError struct{ arg string }

func (e parseError) Error() string { return e.arg }

// parse converts a numeric operand to a float, returning the original text so a
// faithful representation can be printed when no -f format is given.
func parse(s string) (float64, string, error) {
	v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return 0, s, parseError{arg: s}
	}
	return v, s, nil
}

// generate returns the sequence FIRST, FIRST+INCREMENT, ... bounded by LAST.
// When INCREMENT is positive the sequence ascends and stops once it exceeds
// LAST; when negative it descends and stops once it drops below LAST. Counting
// by step index avoids accumulated floating-point drift.
func generate(first, increment, last float64) []float64 {
	var out []float64
	if increment > 0 {
		for i := 0; ; i++ {
			v := first + float64(i)*increment
			if v > last {
				break
			}
			out = append(out, v)
		}
	} else {
		for i := 0; ; i++ {
			v := first + float64(i)*increment
			if v < last {
				break
			}
			out = append(out, v)
		}
	}
	return out
}

// render formats the sequence into the final output string. The number format
// follows GNU seq: if -f is supplied it is used verbatim; otherwise an integer
// format is used when every operand is an integer, and a fixed-precision float
// format (using the maximum operand precision) is used when any operand is a
// float. -w left-pads with zeros to the widest formatted value.
func render(values []float64, separator string, equalWidth bool, format, firstStr, incStr, lastStr string) string {
	if len(values) == 0 {
		return ""
	}

	useFloat := format != "" || isFloat(firstStr) || isFloat(incStr) || isFloat(lastStr)
	prec := 0
	if useFloat {
		prec = maxPrecision(firstStr, incStr, lastStr)
	}

	formatted := make([]string, len(values))
	for i, v := range values {
		switch {
		case format != "":
			formatted[i] = fmt.Sprintf(format, v)
		case useFloat:
			formatted[i] = strconv.FormatFloat(v, 'f', prec, 64)
		default:
			formatted[i] = strconv.FormatInt(int64(math.Round(v)), 10)
		}
	}

	if equalWidth {
		padToEqualWidth(formatted)
	}

	return strings.Join(formatted, separator) + "\n"
}

// padToEqualWidth left-pads every entry with zeros so they share the widest
// width, inserting the zeros after any leading sign, matching GNU seq -w.
func padToEqualWidth(values []string) {
	width := 0
	for _, v := range values {
		if len(v) > width {
			width = len(v)
		}
	}
	for i, v := range values {
		if len(v) >= width {
			continue
		}
		pad := strings.Repeat("0", width-len(v))
		if strings.HasPrefix(v, "-") || strings.HasPrefix(v, "+") {
			values[i] = v[:1] + pad + v[1:]
		} else {
			values[i] = pad + v
		}
	}
}

// isFloat reports whether the operand text denotes a non-integer literal (it
// contains a decimal point or exponent).
func isFloat(s string) bool {
	return strings.ContainsAny(s, ".eEpP")
}

// maxPrecision returns the greatest number of fractional digits among the
// operands, used as the float print precision.
func maxPrecision(operands ...string) int {
	max := 0
	for _, s := range operands {
		if p := precision(s); p > max {
			max = p
		}
	}
	return max
}

// precision returns the count of digits after the decimal point in s.
func precision(s string) int {
	s = strings.TrimSpace(s)
	// Ignore exponent part when measuring fractional digits.
	if i := strings.IndexAny(s, "eE"); i >= 0 {
		s = s[:i]
	}
	if i := strings.IndexByte(s, '.'); i >= 0 {
		return len(s) - i - 1
	}
	return 0
}
