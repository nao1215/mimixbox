// Package free implements the free applet: display the amount of free and used
// memory in the system by reading /proc/meminfo.
package free

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the free applet.
type Command struct{}

// New returns a free command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "free" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Display amount of free and used memory in the system" }

// meminfoSource returns the contents of /proc/meminfo; tests replace it.
var meminfoSource = func() (io.Reader, error) {
	return os.Open("/proc/meminfo")
}

// Run executes free.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]...", stdio.Err).WithHelp(command.Help{
		Description: "Display the total, used, and free physical and swap memory in the system, reading the figures " +
			"from /proc/meminfo. By default the amounts are shown in kibibytes.",
		Examples: []command.Example{
			{Command: "free", Explain: "Show memory usage in kibibytes."},
			{Command: "free -m", Explain: "Show memory usage in mebibytes."},
			{Command: "free -h", Explain: "Show memory usage with human-readable binary suffixes."},
		},
		ExitStatus: "0  the report was printed successfully.\n1  memory information could not be read.",
	})
	bytesUnit := fs.BoolP("bytes", "b", false, "show output in bytes")
	kibi := fs.BoolP("kibi", "k", false, "show output in kibibytes (default)")
	mebi := fs.BoolP("mebi", "m", false, "show output in mebibytes")
	gibi := fs.BoolP("gibi", "g", false, "show output in gibibytes")
	human := fs.BoolP("human", "h", false, "show human-readable output")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	r, err := meminfoSource()
	if err != nil {
		return command.Failuref("cannot read memory information: %v", err)
	}
	if rc, ok := r.(io.Closer); ok {
		defer func() { _ = rc.Close() }()
	}

	info, err := parseMeminfo(r)
	if err != nil {
		return command.Failuref("cannot parse memory information: %v", err)
	}

	unit := chooseUnit(*bytesUnit, *mebi, *gibi, *human, *kibi)
	return c.render(stdio, info, unit)
}

// meminfo holds the kibibyte values free needs from /proc/meminfo.
type meminfo struct {
	memTotal, memFree, memAvailable, buffers, cached, sReclaimable, shmem int64
	swapTotal, swapFree                                                   int64
}

// parseMeminfo reads "Key: value kB" lines into a meminfo.
func parseMeminfo(r io.Reader) (meminfo, error) {
	fields := map[string]*int64{}
	var m meminfo
	for k, p := range map[string]*int64{
		"MemTotal": &m.memTotal, "MemFree": &m.memFree, "MemAvailable": &m.memAvailable,
		"Buffers": &m.buffers, "Cached": &m.cached, "SReclaimable": &m.sReclaimable,
		"Shmem": &m.shmem, "SwapTotal": &m.swapTotal, "SwapFree": &m.swapFree,
	} {
		fields[k] = p
	}

	sc := bufio.NewScanner(r)
	for sc.Scan() {
		key, val, ok := strings.Cut(sc.Text(), ":")
		if !ok {
			continue
		}
		ptr, want := fields[strings.TrimSpace(key)]
		if !want {
			continue
		}
		num := strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(val), "kB"))
		if n, err := strconv.ParseInt(num, 10, 64); err == nil {
			*ptr = n
		}
	}
	return m, sc.Err()
}

// unit describes how to scale and render a kibibyte value.
type unit struct {
	human bool
	div   float64
}

// chooseUnit maps the flags to the output unit, defaulting to kibibytes. All
// raw /proc/meminfo values are in kibibytes, so div scales from there.
func chooseUnit(b, m, g, h, _ bool) unit {
	switch {
	case h:
		return unit{human: true}
	case b:
		return unit{div: 1.0 / 1024}
	case m:
		return unit{div: 1024}
	case g:
		return unit{div: 1024 * 1024}
	default:
		return unit{div: 1}
	}
}

// render writes the Mem and Swap rows scaled to the chosen unit.
func (c *Command) render(stdio command.IO, m meminfo, u unit) error {
	cache := m.cached + m.sReclaimable
	buffCache := m.buffers + cache
	used := m.memTotal - m.memFree - buffCache
	if used < 0 {
		used = 0
	}

	var b strings.Builder
	fmt.Fprintf(&b, "%15s %11s %11s %11s %11s %11s %11s\n",
		"total", "used", "free", "shared", "buff/cache", "available", "")
	fmt.Fprintf(&b, "Mem:    %11s %11s %11s %11s %11s %11s\n",
		fmtVal(m.memTotal, u), fmtVal(used, u), fmtVal(m.memFree, u),
		fmtVal(m.shmem, u), fmtVal(buffCache, u), fmtVal(m.memAvailable, u))
	fmt.Fprintf(&b, "Swap:   %11s %11s %11s\n",
		fmtVal(m.swapTotal, u), fmtVal(m.swapTotal-m.swapFree, u), fmtVal(m.swapFree, u))

	if _, err := io.WriteString(stdio.Out, b.String()); err != nil {
		return command.Failure(err)
	}
	return nil
}

// fmtVal scales a kibibyte value to the unit and formats it, using a
// human-readable suffix when requested.
func fmtVal(kib int64, u unit) string {
	if u.human {
		return human(kib * 1024)
	}
	return strconv.FormatInt(int64(float64(kib)/u.div), 10)
}

// human renders a byte count with a binary suffix (Ki, Mi, Gi, ...).
func human(bytes int64) string {
	const step = 1024.0
	val := float64(bytes)
	units := []string{"B", "Ki", "Mi", "Gi", "Ti"}
	i := 0
	for val >= step && i < len(units)-1 {
		val /= step
		i++
	}
	if i == 0 {
		return fmt.Sprintf("%d%s", int64(val), units[i])
	}
	return fmt.Sprintf("%.1f%s", val, units[i])
}
