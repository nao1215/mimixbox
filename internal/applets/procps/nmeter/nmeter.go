// Package nmeter implements the nmeter applet: print a one-shot line of system
// statistics expanded from a format string.
package nmeter

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the nmeter applet.
type Command struct{}

// New returns a nmeter command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "nmeter" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Print system statistics from a format string" }

// Injected so the snapshot is testable.
var (
	statPath    = "/proc/stat"
	meminfoPath = "/proc/meminfo"
	now         = time.Now
)

// Run executes nmeter (a single snapshot of the format).
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "FORMAT", stdio.Err).WithHelp(command.Help{
		Description: "Expand FORMAT once and print the result. Supported directives: %t (time), " +
			"%c (CPU busy percent since boot), %m (used memory in MiB), %M (total memory in MiB), " +
			"and %% (a literal percent). Other text is copied verbatim.",
		Examples: []command.Example{
			{Command: "nmeter '%t cpu:%c mem:%m'", Explain: "Print time, CPU, and used memory."},
		},
		Notes: []string{
			"Only a single snapshot is printed; the continuous live display of real nmeter is not implemented.",
		},
		ExitStatus: "0  success.\n1  an error occurred (e.g. /proc could not be read or the FORMAT argument was invalid).",
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	rest := fs.Args()
	if len(rest) == 0 {
		_, _ = fmt.Fprintln(stdio.Err, "nmeter: a format string is required")
		return command.SilentFailure()
	}

	_, _ = fmt.Fprintln(stdio.Out, expand(rest[0]))
	return nil
}

// expand replaces the supported directives in format with current values.
func expand(format string) string {
	var b strings.Builder
	for i := 0; i < len(format); i++ {
		if format[i] != '%' || i+1 >= len(format) {
			b.WriteByte(format[i])
			continue
		}
		i++
		switch format[i] {
		case 't':
			b.WriteString(now().Format("15:04:05"))
		case 'c':
			fmt.Fprintf(&b, "%.0f%%", cpuBusy())
		case 'm':
			m := meminfo()
			fmt.Fprintf(&b, "%dM", (m["MemTotal"]-m["MemAvailable"])/1024)
		case 'M':
			fmt.Fprintf(&b, "%dM", meminfo()["MemTotal"]/1024)
		case '%':
			b.WriteByte('%')
		default:
			b.WriteByte('%')
			b.WriteByte(format[i])
		}
	}
	return b.String()
}

// cpuBusy returns the non-idle CPU percentage since boot.
func cpuBusy() float64 {
	data, err := os.ReadFile(statPath)
	if err != nil {
		return 0
	}
	for _, line := range strings.Split(string(data), "\n") {
		f := strings.Fields(line)
		if len(f) == 0 || f[0] != "cpu" {
			continue
		}
		var total, idle int64
		for i := 1; i < len(f); i++ {
			n, _ := strconv.ParseInt(f[i], 10, 64)
			total += n
			if i == 4 { // idle is the 4th value
				idle = n
			}
		}
		if total == 0 {
			return 0
		}
		return float64(total-idle) * 100 / float64(total)
	}
	return 0
}

func meminfo() map[string]int64 {
	m := map[string]int64{}
	data, err := os.ReadFile(meminfoPath)
	if err != nil {
		return m
	}
	for _, line := range strings.Split(string(data), "\n") {
		f := strings.Fields(line)
		if len(f) >= 2 {
			n, _ := strconv.ParseInt(f[1], 10, 64)
			m[strings.TrimSuffix(f[0], ":")] = n
		}
	}
	return m
}
