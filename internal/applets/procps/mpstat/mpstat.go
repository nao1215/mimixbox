// Package mpstat implements the mpstat applet: report per-processor CPU usage
// read from /proc/stat.
package mpstat

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the mpstat applet.
type Command struct{}

// New returns a mpstat command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "mpstat" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Report per-processor CPU statistics" }

// statPath is /proc/stat; tests point it at a fixture.
var statPath = "/proc/stat"

// Run executes mpstat.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "", stdio.Err).WithHelp(command.Help{
		Description: "Print the CPU utilization breakdown for the aggregate ('all') and each individual " +
			"processor, as the percentage of time spent in user, nice, system, iowait, irq, softirq, " +
			"steal, guest, and idle. The values are averages since boot.",
		Examples: []command.Example{
			{Command: "mpstat", Explain: "Show per-CPU usage."},
		},
		ExitStatus: "0  success.\n1  an error occurred (e.g. /proc could not be read or an argument was invalid).",
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	rows := readCPUs()
	if len(rows) == 0 {
		_, _ = fmt.Fprintln(stdio.Err, "mpstat: cannot read CPU statistics")
		return command.SilentFailure()
	}

	_, _ = fmt.Fprintf(stdio.Out, "%-4s %7s %7s %7s %7s %7s %7s %7s %7s %7s %7s\n",
		"CPU", "%usr", "%nice", "%sys", "%iowait", "%irq", "%soft", "%steal", "%guest", "%gnice", "%idle")
	for _, r := range rows {
		p := percentages(r.values)
		_, _ = fmt.Fprintf(stdio.Out, "%-4s %7.2f %7.2f %7.2f %7.2f %7.2f %7.2f %7.2f %7.2f %7.2f %7.2f\n",
			r.label, p[0], p[1], p[2], p[3], p[4], p[5], p[6], p[7], p[8], p[9])
	}
	return nil
}

type cpuRow struct {
	label  string
	values []int64
}

// readCPUs returns the aggregate and per-CPU stat rows in order.
func readCPUs() []cpuRow {
	data, err := os.ReadFile(statPath)
	if err != nil {
		return nil
	}
	var rows []cpuRow
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 || !strings.HasPrefix(fields[0], "cpu") {
			continue
		}
		label := "all"
		if fields[0] != "cpu" {
			label = strings.TrimPrefix(fields[0], "cpu")
		}
		var vals []int64
		for _, f := range fields[1:] {
			n, _ := strconv.ParseInt(f, 10, 64)
			vals = append(vals, n)
		}
		rows = append(rows, cpuRow{label: label, values: vals})
	}
	return rows
}

// percentages converts the cpu jiffy fields to the ten mpstat percentages:
// usr, nice, sys, iowait, irq, soft, steal, guest, gnice, idle.
func percentages(v []int64) [10]float64 {
	get := func(i int) int64 {
		if i < len(v) {
			return v[i]
		}
		return 0
	}
	user, nice, system, idle := get(0), get(1), get(2), get(3)
	iowait, irq, softirq, steal := get(4), get(5), get(6), get(7)
	guest, gnice := get(8), get(9)
	total := user + nice + system + idle + iowait + irq + softirq + steal
	var out [10]float64
	if total == 0 {
		out[9] = 100
		return out
	}
	pct := func(x int64) float64 { return float64(x) * 100 / float64(total) }
	out[0] = pct(user - guest) // mpstat counts guest time separately from user
	out[1] = pct(nice - gnice)
	out[2] = pct(system)
	out[3] = pct(iowait)
	out[4] = pct(irq)
	out[5] = pct(softirq)
	out[6] = pct(steal)
	out[7] = pct(guest)
	out[8] = pct(gnice)
	out[9] = pct(idle)
	return out
}
