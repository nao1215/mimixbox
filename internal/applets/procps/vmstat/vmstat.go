// Package vmstat implements the vmstat applet: report virtual-memory, process,
// I/O, and CPU statistics in a single snapshot read from /proc.
package vmstat

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the vmstat applet.
type Command struct{}

// New returns a vmstat command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "vmstat" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Report virtual memory statistics" }

// Injected so the snapshot is testable.
var (
	statPath    = "/proc/stat"
	meminfoPath = "/proc/meminfo"
	vmPath      = "/proc/vmstat"
	uptimePath  = "/proc/uptime"
)

// pageKB is the page size in KiB used to convert swap-page counters.
const pageKB = 4

// Run executes vmstat.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "", stdio.Err).WithHelp(command.Help{
		Description: "Print a single line of system statistics: runnable/blocked processes, memory " +
			"(swap used, free, buffers, cache), swap and block I/O, interrupts and context switches, " +
			"and the CPU time breakdown. The I/O and system columns are averages since boot.",
		Examples: []command.Example{
			{Command: "vmstat", Explain: "Show one snapshot of system statistics."},
		},
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	mem := readKeyed(meminfoPath)
	vm := readKeyed(vmPath)
	stat := readStat()
	up := readUptime()
	if up < 1 {
		up = 1
	}

	swpd := mem["SwapTotal"] - mem["SwapFree"]
	us, sy, id, wa, st := cpuPercents(stat.cpu)

	_, _ = fmt.Fprintln(stdio.Out, "procs -----------memory---------- ---swap-- -----io---- -system-- ------cpu-----")
	_, _ = fmt.Fprintln(stdio.Out, " r  b   swpd   free   buff  cache   si   so    bi    bo   in   cs us sy id wa st")
	_, _ = fmt.Fprintf(stdio.Out, "%2d %2d %6d %6d %6d %6d %4d %4d %5d %5d %4d %4d %2d %2d %2d %2d %2d\n",
		stat.running, stat.blocked,
		swpd, mem["MemFree"], mem["Buffers"], mem["Cached"],
		vm["pswpin"]*pageKB/up, vm["pswpout"]*pageKB/up,
		vm["pgpgin"]/up, vm["pgpgout"]/up,
		stat.intr/up, stat.ctxt/up,
		us, sy, id, wa, st)
	return nil
}

type statData struct {
	cpu              []int64
	intr, ctxt       int64
	running, blocked int64
}

// readStat parses the cpu, intr, ctxt and procs lines of /proc/stat.
func readStat() statData {
	var s statData
	data, err := os.ReadFile(statPath)
	if err != nil {
		return s
	}
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		switch fields[0] {
		case "cpu":
			for _, f := range fields[1:] {
				n, _ := strconv.ParseInt(f, 10, 64)
				s.cpu = append(s.cpu, n)
			}
		case "intr":
			s.intr = atoi64(fields)
		case "ctxt":
			s.ctxt = atoi64(fields)
		case "procs_running":
			s.running = atoi64(fields)
		case "procs_blocked":
			s.blocked = atoi64(fields)
		}
	}
	return s
}

func atoi64(fields []string) int64 {
	if len(fields) < 2 {
		return 0
	}
	n, _ := strconv.ParseInt(fields[1], 10, 64)
	return n
}

// cpuPercents converts the cumulative cpu jiffies to integer percentages.
func cpuPercents(cpu []int64) (us, sy, id, wa, st int64) {
	get := func(i int) int64 {
		if i < len(cpu) {
			return cpu[i]
		}
		return 0
	}
	user, nice, system, idle := get(0), get(1), get(2), get(3)
	iowait, irq, softirq, steal := get(4), get(5), get(6), get(7)
	total := user + nice + system + idle + iowait + irq + softirq + steal
	if total == 0 {
		return 0, 0, 100, 0, 0
	}
	pct := func(v int64) int64 { return v * 100 / total }
	return pct(user + nice), pct(system + irq + softirq), pct(idle), pct(iowait), pct(steal)
}

// readKeyed reads a "Key: value kB" or "key value" file into a map of int64.
func readKeyed(path string) map[string]int64 {
	m := map[string]int64{}
	data, err := os.ReadFile(path)
	if err != nil {
		return m
	}
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		key := strings.TrimSuffix(fields[0], ":")
		n, err := strconv.ParseInt(fields[1], 10, 64)
		if err == nil {
			m[key] = n
		}
	}
	return m
}

func readUptime() int64 {
	data, err := os.ReadFile(uptimePath)
	if err != nil {
		return 0
	}
	fields := strings.Fields(string(data))
	if len(fields) == 0 {
		return 0
	}
	secs, _ := strconv.ParseFloat(fields[0], 64)
	return int64(secs)
}
