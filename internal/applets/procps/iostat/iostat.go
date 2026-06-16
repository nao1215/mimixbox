// Package iostat implements the iostat applet: report CPU utilization and
// per-device I/O statistics read from /proc/stat and /proc/diskstats.
package iostat

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the iostat applet.
type Command struct{}

// New returns an iostat command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "iostat" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Report CPU and device I/O statistics" }

// Injected so the report is testable.
var (
	statPath      = "/proc/stat"
	diskstatsPath = "/proc/diskstats"
	uptimePath    = "/proc/uptime"
)

// Run executes iostat.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "", stdio.Err).WithHelp(command.Help{
		Description: "Print the average CPU utilization since boot, then a per-device table of " +
			"transfers per second and kilobytes read and written, from /proc/stat and " +
			"/proc/diskstats. Devices with no activity are omitted.",
		Examples: []command.Example{
			{Command: "iostat", Explain: "Show CPU and disk I/O statistics."},
		},
		ExitStatus: "0  success.\n1  an error occurred (e.g. /proc could not be read or an argument was invalid).",
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	up := readUptime()
	if up < 1 {
		up = 1
	}

	user, nice, system, iowait, steal, idle := cpuPercents()
	_, _ = io.WriteString(stdio.Out, "avg-cpu:  %user   %nice %system %iowait  %steal   %idle\n")
	_, _ = fmt.Fprintf(stdio.Out, "        %7.2f %7.2f %7.2f %7.2f %7.2f %7.2f\n\n",
		user, nice, system, iowait, steal, idle)

	_, _ = fmt.Fprintf(stdio.Out, "%-13s %8s %12s %12s %10s %10s\n",
		"Device", "tps", "kB_read/s", "kB_wrtn/s", "kB_read", "kB_wrtn")
	for _, d := range readDisks() {
		if d.reads == 0 && d.writes == 0 {
			continue // omit inactive devices, like iostat
		}
		kbRead := d.sectorsRead / 2
		kbWrtn := d.sectorsWritten / 2
		tps := float64(d.reads+d.writes) / float64(up)
		_, _ = fmt.Fprintf(stdio.Out, "%-13s %8.2f %12.2f %12.2f %10d %10d\n",
			d.name, tps, float64(kbRead)/float64(up), float64(kbWrtn)/float64(up), kbRead, kbWrtn)
	}
	return nil
}

// cpuPercents returns the avg-cpu breakdown from /proc/stat.
func cpuPercents() (user, nice, system, iowait, steal, idle float64) {
	data, err := os.ReadFile(statPath)
	if err != nil {
		return 0, 0, 0, 0, 0, 100
	}
	for _, line := range strings.Split(string(data), "\n") {
		f := strings.Fields(line)
		if len(f) == 0 || f[0] != "cpu" {
			continue
		}
		v := make([]int64, 10)
		for i := 1; i < len(f) && i <= 10; i++ {
			v[i-1], _ = strconv.ParseInt(f[i], 10, 64)
		}
		total := v[0] + v[1] + v[2] + v[3] + v[4] + v[5] + v[6] + v[7]
		if total == 0 {
			return 0, 0, 0, 0, 0, 100
		}
		p := func(x int64) float64 { return float64(x) * 100 / float64(total) }
		return p(v[0] - v[8]), p(v[1] - v[9]), p(v[2] + v[5] + v[6]), p(v[4]), p(v[7]), p(v[3])
	}
	return 0, 0, 0, 0, 0, 100
}

type disk struct {
	name                        string
	reads, writes               int64
	sectorsRead, sectorsWritten int64
}

// readDisks parses the per-device counters of /proc/diskstats.
func readDisks() []disk {
	data, err := os.ReadFile(diskstatsPath)
	if err != nil {
		return nil
	}
	var disks []disk
	for _, line := range strings.Split(string(data), "\n") {
		f := strings.Fields(line)
		if len(f) < 10 {
			continue
		}
		disks = append(disks, disk{
			name:           f[2],
			reads:          atoi(f[3]),
			sectorsRead:    atoi(f[5]),
			writes:         atoi(f[7]),
			sectorsWritten: atoi(f[9]),
		})
	}
	return disks
}

func atoi(s string) int64 {
	n, _ := strconv.ParseInt(s, 10, 64)
	return n
}

func readUptime() int64 {
	data, err := os.ReadFile(uptimePath)
	if err != nil {
		return 0
	}
	f := strings.Fields(string(data))
	if len(f) == 0 {
		return 0
	}
	secs, _ := strconv.ParseFloat(f[0], 64)
	return int64(secs)
}
