// Package top implements the top applet in batch mode: print a one-shot snapshot
// of the system summary and the processes, sorted by resident memory.
package top

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the top applet.
type Command struct{}

// New returns a top command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "top" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Display system summary and top processes" }

// Injected so the snapshot is testable.
var (
	procDir     = "/proc"
	meminfoPath = "/proc/meminfo"
	uptimePath  = "/proc/uptime"
	loadavgPath = "/proc/loadavg"
	now         = time.Now
)

const pageKB = 4

// Run executes top (always a single batch snapshot).
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-b] [-n COUNT]", stdio.Err).WithHelp(command.Help{
		Description: "Print a one-shot snapshot: the uptime/load summary, a task-state count, a memory " +
			"summary, and the processes sorted by resident memory. Only batch mode (one iteration) " +
			"is supported; -b and -n are accepted for compatibility.",
		Examples: []command.Example{
			{Command: "top -bn1", Explain: "Print one batch snapshot."},
		},
		Notes: []string{
			"The interactive display and the %CPU/PR/NI columns of real top are not implemented.",
		},
	})
	_ = fs.BoolP("batch", "b", false, "batch mode (the only supported mode)")
	_ = fs.IntP("iterations", "n", 1, "number of iterations (only 1 is performed)")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	procs := readProcs()
	c.summary(stdio.Out, procs)

	sort.Slice(procs, func(i, j int) bool { return procs[i].res > procs[j].res })
	_, _ = fmt.Fprintf(stdio.Out, "%7s %-8s %8s %5s %s %s\n", "PID", "USER", "RES", "%MEM", "S", "COMMAND")
	memTotal := meminfo()["MemTotal"]
	for _, p := range procs {
		memPct := 0.0
		if memTotal > 0 {
			memPct = float64(p.res) * 100 / float64(memTotal)
		}
		_, _ = fmt.Fprintf(stdio.Out, "%7d %-8s %8d %5.1f %s %s\n", p.pid, p.user, p.res, memPct, p.state, p.comm)
	}
	return nil
}

type process struct {
	pid   int
	user  string
	res   int64 // resident KiB
	state string
	comm  string
}

// summary prints the top/Tasks/Mem header lines.
func (c *Command) summary(out io.Writer, procs []process) {
	running, sleeping, stopped, zombie := 0, 0, 0, 0
	for _, p := range procs {
		switch p.state {
		case "R":
			running++
		case "T", "t":
			stopped++
		case "Z":
			zombie++
		default:
			sleeping++
		}
	}
	_, _ = fmt.Fprintf(out, "top - %s up %s,  %d users,  load average: %s\n",
		now().Format("15:04:05"), uptimeStr(), countUsers(), loadavg())
	_, _ = fmt.Fprintf(out, "Tasks: %d total, %d running, %d sleeping, %d stopped, %d zombie\n",
		len(procs), running, sleeping, stopped, zombie)
	m := meminfo()
	used := m["MemTotal"] - m["MemFree"] - m["Buffers"] - m["Cached"]
	_, _ = fmt.Fprintf(out, "MiB Mem : %8.1f total, %8.1f free, %8.1f used, %8.1f buff/cache\n\n",
		mib(m["MemTotal"]), mib(m["MemFree"]), mib(used), mib(m["Buffers"]+m["Cached"]))
}

func mib(kb int64) float64 { return float64(kb) / 1024 }

// readProcs gathers the per-process fields top displays.
func readProcs() []process {
	entries, err := os.ReadDir(procDir)
	if err != nil {
		return nil
	}
	var procs []process
	for _, e := range entries {
		pid, err := strconv.Atoi(e.Name())
		if err != nil {
			continue
		}
		comm, state, ok := statComm(pid)
		if !ok {
			continue
		}
		procs = append(procs, process{
			pid: pid, user: owner(pid), res: residentKB(pid), state: state, comm: comm,
		})
	}
	return procs
}

func statComm(pid int) (comm, state string, ok bool) {
	data, err := os.ReadFile(filepath.Join(procDir, strconv.Itoa(pid), "stat")) //nolint:gosec // /proc path
	if err != nil {
		return "", "", false
	}
	line := string(data)
	o := strings.IndexByte(line, '(')
	cl := strings.LastIndexByte(line, ')')
	if o < 0 || cl < 0 || cl < o {
		return "", "", false
	}
	comm = line[o+1 : cl]
	f := strings.Fields(line[cl+1:])
	if len(f) < 1 {
		return "", "", false
	}
	return comm, f[0], true
}

// residentKB reads the resident set size (in KiB) from /proc/PID/statm.
func residentKB(pid int) int64 {
	data, err := os.ReadFile(filepath.Join(procDir, strconv.Itoa(pid), "statm")) //nolint:gosec // /proc path
	if err != nil {
		return 0
	}
	f := strings.Fields(string(data))
	if len(f) < 2 {
		return 0
	}
	pages, _ := strconv.ParseInt(f[1], 10, 64)
	return pages * pageKB
}

func owner(pid int) string {
	info, err := os.Stat(filepath.Join(procDir, strconv.Itoa(pid)))
	if err != nil {
		return "?"
	}
	st, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return "?"
	}
	uid := strconv.FormatUint(uint64(st.Uid), 10)
	if u, err := user.LookupId(uid); err == nil {
		return u.Username
	}
	return uid
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

func uptimeStr() string {
	data, err := os.ReadFile(uptimePath)
	if err != nil {
		return "0:00"
	}
	f := strings.Fields(string(data))
	if len(f) == 0 {
		return "0:00"
	}
	secs, _ := strconv.ParseFloat(f[0], 64)
	d := time.Duration(secs) * time.Second
	mins := int(d.Minutes())
	if days := mins / 1440; days > 0 {
		return fmt.Sprintf("%d days, %2d:%02d", days, (mins/60)%24, mins%60)
	}
	return fmt.Sprintf("%2d:%02d", (mins/60)%24, mins%60)
}

func loadavg() string {
	data, err := os.ReadFile(loadavgPath)
	if err != nil {
		return "0.00, 0.00, 0.00"
	}
	f := strings.Fields(string(data))
	if len(f) < 3 {
		return "0.00, 0.00, 0.00"
	}
	return fmt.Sprintf("%s, %s, %s", f[0], f[1], f[2])
}

func countUsers() int { return 0 }
