// Package date implements the date applet: print the system date and time,
// optionally formatted with strftime-style conversion specifiers, matching the
// common GNU date semantics. Setting the system clock is out of scope.
package date

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
)

// nowFn is the clock the command reads "now" from. It is a package variable so
// tests can replace it with a fixed time and assert on deterministic output.
var nowFn = time.Now

// Command is the date applet.
type Command struct{}

// New returns a date command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "date" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Print or set the system date and time" }

// Run executes date.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... [+FORMAT]", stdio.Err)
	utc := fs.BoolP("utc", "u", false, "print Coordinated Universal Time (UTC)")
	dateStr := fs.StringP("date", "d", "", "display time described by STRING, not 'now'")
	rfcEmail := fs.BoolP("rfc-email", "R", false, "output date and time in RFC 5322 format")
	iso := fs.StringP("iso-8601", "I", "", "output date/time in ISO 8601 format; FMT may be 'date', 'hours', 'minutes', 'seconds', or 'ns'")
	// Allow -I with no value (GNU treats a bare -I as --iso-8601=date).
	fs.Lookup("iso-8601").NoOptDefVal = "date"
	setStr := fs.StringP("set", "s", "", "set time described by STRING")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	if *setStr != "" {
		return command.Failuref("setting the system clock is not supported")
	}

	t, err := resolveTime(*dateStr)
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "date: %v\n", err)
		return command.SilentFailure()
	}
	if *utc {
		t = t.UTC()
	}

	operands := fs.Args()
	format, err := operandFormat(operands)
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "date: %v\n", err)
		return command.SilentFailure()
	}

	out, err := output(t, format, *rfcEmail, *iso, fs.Changed("iso-8601"))
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "date: %v\n", err)
		return command.SilentFailure()
	}
	_, _ = fmt.Fprintln(stdio.Out, out)
	return nil
}

// operandFormat extracts the optional +FORMAT operand. GNU date accepts at most
// one operand and it must begin with '+'.
func operandFormat(operands []string) (string, error) {
	switch len(operands) {
	case 0:
		return "", nil
	case 1:
		op := operands[0]
		if !strings.HasPrefix(op, "+") {
			return "", fmt.Errorf("invalid date %q", op)
		}
		return op[1:], nil
	default:
		return "", fmt.Errorf("extra operand %q", operands[1])
	}
}

// output renders the time according to the highest-priority format requested:
// -R (RFC email), then -I (ISO 8601), then a +FORMAT operand, then the default
// "date"-style layout.
func output(t time.Time, format string, rfcEmail bool, isoFmt string, isoSet bool) (string, error) {
	if rfcEmail {
		return t.Format("Mon, 02 Jan 2006 15:04:05 -0700"), nil
	}
	if isoSet {
		return isoFormat(t, isoFmt)
	}
	if format != "" {
		return formatTime(t, format), nil
	}
	// GNU default, e.g. "Tue Nov 14 22:13:20 UTC 2023".
	return t.Format("Mon Jan  2 15:04:05 MST 2006"), nil
}

// isoFormat implements the granularity argument of --iso-8601.
func isoFormat(t time.Time, fmtArg string) (string, error) {
	switch fmtArg {
	case "", "date":
		return t.Format("2006-01-02"), nil
	case "hours":
		return t.Format("2006-01-02T15-07:00"), nil
	case "minutes":
		return t.Format("2006-01-02T15:04-07:00"), nil
	case "seconds":
		return t.Format("2006-01-02T15:04:05-07:00"), nil
	case "ns":
		return t.Format("2006-01-02T15:04:05,999999999-07:00"), nil
	default:
		return "", fmt.Errorf("invalid argument %q for '--iso-8601'", fmtArg)
	}
}

// resolveTime returns the time to display: now when s is empty, otherwise the
// time described by s. Supported forms are "@SECONDS" (epoch), RFC3339, and a
// bare "YYYY-MM-DD" date.
func resolveTime(s string) (time.Time, error) {
	if s == "" {
		return nowFn(), nil
	}
	if strings.HasPrefix(s, "@") {
		sec, err := strconv.ParseInt(s[1:], 10, 64)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid date %q", s)
		}
		return time.Unix(sec, 0), nil
	}
	for _, layout := range []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02",
	} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid date %q", s)
}

// formatTime converts a strftime-style format string into a rendered string for
// t. It is a pure function (its only input is t and format) so it can be unit
// tested exhaustively. Unknown specifiers are emitted verbatim including the
// leading '%', matching GNU date's lenient behavior.
func formatTime(t time.Time, format string) string {
	var b strings.Builder
	for i := 0; i < len(format); i++ {
		if format[i] != '%' || i+1 >= len(format) {
			b.WriteByte(format[i])
			continue
		}
		i++
		b.WriteString(specifier(t, format[i]))
	}
	return b.String()
}

// specifier renders a single strftime conversion character for t.
func specifier(t time.Time, c byte) string {
	switch c {
	case '%':
		return "%"
	case 'n':
		return "\n"
	case 't':
		return "\t"
	case 'Y':
		return strconv.Itoa(t.Year())
	case 'y':
		return fmt.Sprintf("%02d", t.Year()%100)
	case 'C':
		return fmt.Sprintf("%02d", t.Year()/100)
	case 'm':
		return fmt.Sprintf("%02d", int(t.Month()))
	case 'd':
		return fmt.Sprintf("%02d", t.Day())
	case 'e':
		return fmt.Sprintf("%2d", t.Day())
	case 'H':
		return fmt.Sprintf("%02d", t.Hour())
	case 'I':
		return fmt.Sprintf("%02d", hour12(t.Hour()))
	case 'M':
		return fmt.Sprintf("%02d", t.Minute())
	case 'S':
		return fmt.Sprintf("%02d", t.Second())
	case 'p':
		return ampm(t.Hour(), true)
	case 'P':
		return ampm(t.Hour(), false)
	case 'A':
		return t.Weekday().String()
	case 'a':
		return t.Weekday().String()[:3]
	case 'B':
		return t.Month().String()
	case 'b', 'h':
		return t.Month().String()[:3]
	case 'j':
		return fmt.Sprintf("%03d", t.YearDay())
	case 'u':
		// ISO weekday, Monday=1..Sunday=7.
		if w := int(t.Weekday()); w == 0 {
			return "7"
		} else {
			return strconv.Itoa(w)
		}
	case 'w':
		// Weekday, Sunday=0..Saturday=6.
		return strconv.Itoa(int(t.Weekday()))
	case 'F':
		return t.Format("2006-01-02")
	case 'T':
		return t.Format("15:04:05")
	case 'R':
		return t.Format("15:04")
	case 'D':
		return t.Format("01/02/06")
	case 'r':
		return t.Format("03:04:05 PM")
	case 'Z':
		name, _ := t.Zone()
		return name
	case 'z':
		return t.Format("-0700")
	case 's':
		return strconv.FormatInt(t.Unix(), 10)
	case 'c':
		return t.Format("Mon Jan  2 15:04:05 2006")
	case 'x':
		return t.Format("01/02/06")
	case 'X':
		return t.Format("15:04:05")
	default:
		return "%" + string(c)
	}
}

func hour12(h int) int {
	h %= 12
	if h == 0 {
		return 12
	}
	return h
}

func ampm(hour int, upper bool) string {
	s := "am"
	if hour >= 12 {
		s = "pm"
	}
	if upper {
		return strings.ToUpper(s)
	}
	return s
}
