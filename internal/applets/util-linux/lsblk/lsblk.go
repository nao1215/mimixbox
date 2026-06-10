// Package lsblk implements the lsblk applet: list block devices and their
// partitions, read from /sys/block, with mount points from /proc/mounts.
package lsblk

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the lsblk applet.
type Command struct{}

// New returns an lsblk command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "lsblk" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "List information about block devices" }

// These are injectable so the listing can be tested against fixtures.
var (
	sysBlock   = "/sys/block"
	procMounts = "/proc/mounts"
)

type device struct {
	name   string
	majMin string
	rm     string
	size   int64
	ro     string
	typ    string
	mount  string
}

// Run executes lsblk.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-a]", stdio.Err).WithHelp(command.Help{
		Description: "List the block devices under /sys/block and their partitions as a tree, with " +
			"each device's MAJ:MIN, removable flag, size, read-only flag, type, and mount point. " +
			"Empty devices (size 0) are hidden unless -a is given.",
		Examples: []command.Example{
			{Command: "lsblk", Explain: "List the block devices."},
			{Command: "lsblk -a", Explain: "Include empty devices."},
		},
	})
	all := fs.BoolP("all", "a", false, "also list empty devices")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	mounts := readMounts(procMounts)
	names, err := os.ReadDir(sysBlock)
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "lsblk: %s\n", command.FileError(sysBlock, err))
		return command.SilentFailure()
	}

	var disks []string
	for _, n := range names {
		disks = append(disks, n.Name())
	}
	sort.Strings(disks)

	_, _ = fmt.Fprintf(stdio.Out, "%-12s %-7s %2s %6s %2s %-4s %s\n", "NAME", "MAJ:MIN", "RM", "SIZE", "RO", "TYPE", "MOUNTPOINTS")
	for _, name := range disks {
		d := readDevice(name, "disk", mounts)
		// Hide empty devices and RAM disks by default, as lsblk does.
		if (d.size == 0 || strings.HasPrefix(name, "ram")) && !*all {
			continue
		}
		printDevice(stdio.Out, d, "")

		parts := partitions(name)
		for i, p := range parts {
			pd := readDevice(filepath.Join(name, p), "part", mounts)
			pd.name = p
			prefix := "├─"
			if i == len(parts)-1 {
				prefix = "└─"
			}
			printDevice(stdio.Out, pd, prefix)
		}
	}
	return nil
}

// readDevice reads one device's attributes from /sys/block/<rel>.
func readDevice(rel, typ string, mounts map[string]string) device {
	dir := filepath.Join(sysBlock, rel)
	name := filepath.Base(rel)
	d := device{
		name:   name,
		majMin: readField(filepath.Join(dir, "dev")),
		rm:     readField(filepath.Join(dir, "removable")),
		size:   readInt(filepath.Join(dir, "size")) * 512,
		ro:     readField(filepath.Join(dir, "ro")),
		typ:    typ,
		mount:  mounts["/dev/"+name],
	}
	if d.rm == "" {
		d.rm = "0"
	}
	if d.ro == "" {
		d.ro = "0"
	}
	return d
}

// partitions returns the partition sub-device names of a disk, sorted.
func partitions(disk string) []string {
	entries, err := os.ReadDir(filepath.Join(sysBlock, disk))
	if err != nil {
		return nil
	}
	var parts []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if _, err := os.Stat(filepath.Join(sysBlock, disk, e.Name(), "partition")); err == nil {
			parts = append(parts, e.Name())
		}
	}
	sort.Strings(parts)
	return parts
}

func printDevice(out io.Writer, d device, prefix string) {
	_, _ = fmt.Fprintf(out, "%-12s %-7s %2s %6s %2s %-4s %s\n",
		prefix+d.name, d.majMin, d.rm, humanSize(d.size), d.ro, d.typ, d.mount)
}

// readMounts maps a device path to its mount point from /proc/mounts.
func readMounts(path string) map[string]string {
	m := map[string]string{}
	f, err := os.Open(path) //nolint:gosec // the mounts path
	if err != nil {
		return m
	}
	defer func() { _ = f.Close() }()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) >= 2 {
			if _, ok := m[fields[0]]; !ok {
				m[fields[0]] = fields[1]
			}
		}
	}
	return m
}

func readField(path string) string {
	data, err := os.ReadFile(path) //nolint:gosec // a /sys attribute path
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func readInt(path string) int64 {
	n, _ := strconv.ParseInt(readField(path), 10, 64)
	return n
}

// humanSize renders a byte count the way lsblk does: 1024-based with a single
// decimal that is dropped when it is whole (2G, 256G, 364.8M).
func humanSize(bytes int64) string {
	if bytes == 0 {
		return "0B"
	}
	const units = "BKMGTP"
	val := float64(bytes)
	i := 0
	for val >= 1024 && i < len(units)-1 {
		val /= 1024
		i++
	}
	// Format to one decimal, then drop a trailing ".0" as lsblk does (2G, not 2.0G).
	s := strings.TrimSuffix(fmt.Sprintf("%.1f", val), ".0")
	return s + string(units[i])
}
