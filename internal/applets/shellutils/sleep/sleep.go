// Package sleep implements the sleep applet: pause for a given amount of time.
package sleep

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the sleep applet.
type Command struct{}

// New returns a sleep command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "sleep" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Pause for NUMBER seconds(minutes, hours, days)" }

// Run executes sleep. Each operand is a number with an optional suffix:
// s (seconds, the default), m (minutes), h (hours) or d (days). The command
// sleeps for the sum of all operands, and the sleep is cancelled if ctx is
// done.
func (c *Command) Run(ctx context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "NUMBER[smhd]...", stdio.Err)

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	operands := fs.Args()
	if len(operands) == 0 {
		fmt.Fprintln(stdio.Err, "sleep: missing operand")
		return command.SilentFailure()
	}

	d, err := parseDuration(operands)
	if err != nil {
		fmt.Fprintf(stdio.Err, "sleep: invalid time interval '%s'\n", err.Error())
		return command.SilentFailure()
	}

	select {
	case <-time.After(d):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// parseDuration sums each operand into a single time.Duration. An operand is a
// number with an optional suffix s (seconds, default), m (minutes), h (hours)
// or d (days). On an invalid operand it returns an error whose message is the
// offending operand, so the caller can format the GNU-style diagnostic.
func parseDuration(args []string) (time.Duration, error) {
	var total time.Duration
	for _, arg := range args {
		if arg == "" {
			return 0, fmt.Errorf("%s", arg)
		}

		num := arg
		unit := time.Second
		switch arg[len(arg)-1] {
		case 's':
			num, unit = arg[:len(arg)-1], time.Second
		case 'm':
			num, unit = arg[:len(arg)-1], time.Minute
		case 'h':
			num, unit = arg[:len(arg)-1], time.Hour
		case 'd':
			num, unit = arg[:len(arg)-1], 24*time.Hour
		}

		val, err := strconv.ParseFloat(num, 64)
		if err != nil || val < 0 {
			return 0, fmt.Errorf("%s", arg)
		}
		total += time.Duration(val * float64(unit))
	}
	return total, nil
}
