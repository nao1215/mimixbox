// Package cal implements the cal applet: display a simple calendar for a
// month (or a whole year), in the layout of util-linux cal.
package cal

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
)

// nowFn returns the current time. It is a package variable so tests can pin
// "today" and exercise the no-operand (current month) and -y (current year)
// paths deterministically.
var nowFn = time.Now

// calWidth is the printed width of a week row: seven day columns of two
// characters each, separated by single spaces (7*2 + 6 = 20).
const calWidth = 7*2 + 6

// Command is the cal applet.
type Command struct{}

// New returns a cal command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "cal" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Display a calendar" }

// Run executes cal.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... [[MONTH] YEAR]", stdio.Err)
	monday := fs.BoolP("monday", "m", false, "Monday as the first day of the week")
	yearMode := fs.BoolP("year", "y", false, "display a calendar for the whole current year")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	now := nowFn()
	mondayFirst := *monday

	operands := fs.Args()
	switch {
	case *yearMode && len(operands) == 0:
		_, _ = io.WriteString(stdio.Out, year(now.Year(), mondayFirst))
		return nil
	case len(operands) == 0:
		_, _ = io.WriteString(stdio.Out, month(now.Year(), now.Month(), mondayFirst))
		return nil
	case len(operands) == 1:
		y, perr := parseYear(operands[0])
		if perr != nil {
			return usageErr(stdio, c.Name(), perr)
		}
		_, _ = io.WriteString(stdio.Out, year(y, mondayFirst))
		return nil
	case len(operands) == 2:
		m, perr := parseMonth(operands[0])
		if perr != nil {
			return usageErr(stdio, c.Name(), perr)
		}
		y, perr := parseYear(operands[1])
		if perr != nil {
			return usageErr(stdio, c.Name(), perr)
		}
		_, _ = io.WriteString(stdio.Out, month(y, m, mondayFirst))
		return nil
	default:
		return usageErr(stdio, c.Name(), fmt.Errorf("too many arguments"))
	}
}

func usageErr(stdio command.IO, name string, err error) error {
	_, _ = fmt.Fprintf(stdio.Err, "%s: %v\n", name, err)
	_, _ = fmt.Fprintf(stdio.Err, "Try '%s --help' for more information.\n", name)
	return command.SilentFailure()
}

func parseYear(s string) (int, error) {
	y, err := strconv.Atoi(s)
	if err != nil || y < 1 || y > 9999 {
		return 0, fmt.Errorf("illegal year value: use 1-9999")
	}
	return y, nil
}

func parseMonth(s string) (time.Month, error) {
	m, err := strconv.Atoi(s)
	if err != nil || m < 1 || m > 12 {
		return 0, fmt.Errorf("illegal month value: use 1-12")
	}
	return time.Month(m), nil
}

// month renders one month block: a centered "Month YYYY" header, a weekday
// header, then one line per week with right-aligned two-digit days. Days sit in
// three-character columns (a space separator plus two digits); trailing spaces
// on each line are trimmed, matching util-linux cal. The block ends with a
// trailing newline. When mondayFirst is true the week starts on Monday.
func month(y int, m time.Month, mondayFirst bool) string {
	var b strings.Builder

	title := fmt.Sprintf("%s %d", m.String(), y)
	b.WriteString(center(title, calWidth))
	b.WriteByte('\n')

	b.WriteString(weekdayHeader(mondayFirst))
	b.WriteByte('\n')

	// Number of leading blank columns before the 1st of the month.
	first := time.Date(y, m, 1, 0, 0, 0, 0, time.UTC)
	lead := weekdayIndex(first.Weekday(), mondayFirst)

	daysInMonth := time.Date(y, m+1, 1, 0, 0, 0, 0, time.UTC).Add(-24 * time.Hour).Day()

	col := 0
	var line strings.Builder
	writeCell := func(s string) {
		if col > 0 {
			line.WriteByte(' ')
		}
		line.WriteString(s)
		col++
		if col == 7 {
			b.WriteString(strings.TrimRight(line.String(), " "))
			b.WriteByte('\n')
			line.Reset()
			col = 0
		}
	}

	for i := 0; i < lead; i++ {
		writeCell("  ")
	}
	for d := 1; d <= daysInMonth; d++ {
		writeCell(fmt.Sprintf("%2d", d))
	}
	if col > 0 {
		b.WriteString(strings.TrimRight(line.String(), " "))
		b.WriteByte('\n')
	}

	return b.String()
}

// year renders all twelve months of y, one month block per stanza separated by
// a blank line, preceded by a centered year header.
func year(y int, mondayFirst bool) string {
	var b strings.Builder
	b.WriteString(center(strconv.Itoa(y), calWidth))
	b.WriteString("\n\n")
	for m := time.January; m <= time.December; m++ {
		b.WriteString(month(y, m, mondayFirst))
		if m != time.December {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

// weekdayHeader returns the "Su Mo Tu We Th Fr Sa" line (or its Monday-first
// rotation).
func weekdayHeader(mondayFirst bool) string {
	sunFirst := []string{"Su", "Mo", "Tu", "We", "Th", "Fr", "Sa"}
	if mondayFirst {
		return "Mo Tu We Th Fr Sa Su"
	}
	return strings.Join(sunFirst, " ")
}

// weekdayIndex maps a weekday to its column (0..6) given the week start.
func weekdayIndex(w time.Weekday, mondayFirst bool) int {
	if mondayFirst {
		return (int(w) + 6) % 7
	}
	return int(w)
}

// center returns s padded with leading spaces so it sits centered in width.
// Like util-linux, the extra odd space goes on the right (floor division for
// the left padding), and the result is not right-padded.
func center(s string, width int) string {
	if len(s) >= width {
		return s
	}
	left := (width - len(s)) / 2
	return strings.Repeat(" ", left) + s
}
