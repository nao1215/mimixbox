// Package bootchartd implements the bootchartd applet: collect a snapshot of the
// kernel performance counters into the bootchart log files.
package bootchartd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the bootchartd applet.
type Command struct{}

// New returns a bootchartd command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "bootchartd" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Collect a bootchart performance sample" }

// Injected so the kernel counters and the output location are testable.
var (
	statPath      = "/proc/stat"
	diskstatsPath = "/proc/diskstats"
	uptimePath    = "/proc/uptime"
	outputDir     = "/var/log/bootchart"
)

// sample describes one bootchart log file and its kernel source.
var samples = []struct {
	logName string
	source  func() string
}{
	{"proc_stat.log", func() string { return read(statPath) }},
	{"proc_diskstats.log", func() string { return read(diskstatsPath) }},
}

// Run executes bootchartd.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-o DIR]", stdio.Err).WithHelp(command.Help{
		Description: "Append a timestamped snapshot of the kernel performance counters (/proc/stat and " +
			"/proc/diskstats) to the bootchart log files in DIR (default /var/log/bootchart), in the " +
			"format the bootchart renderer reads. Run repeatedly during boot to build a profile; the " +
			"continuous sampling loop and tarball packaging of the full bootchartd are not implemented.",
		Examples: []command.Example{
			{Command: "bootchartd -o /tmp/bootlog", Explain: "Record one sample into /tmp/bootlog."},
		},
		ExitStatus: "0  the sample was recorded.\n1  the output directory could not be written.",
	})
	out := fs.StringP("output", "o", "", "directory for the log files (default /var/log/bootchart)")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	if *out != "" {
		outputDir = *out
	}

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return command.Failuref("cannot create %s: %v", outputDir, err)
	}

	stamp := uptimeJiffies()
	for _, s := range samples {
		entry := fmt.Sprintf("%d\n%s\n", stamp, strings.TrimRight(s.source(), "\n"))
		if err := appendFile(filepath.Join(outputDir, s.logName), entry); err != nil {
			return command.Failuref("cannot write %s: %v", s.logName, err)
		}
	}
	_, _ = fmt.Fprintf(stdio.Out, "bootchartd: sample recorded in %s\n", outputDir)
	return nil
}

// uptimeJiffies returns the system uptime in 1/100-second jiffies, the timestamp
// bootchart uses to separate samples.
func uptimeJiffies() int64 {
	fields := strings.Fields(read(uptimePath))
	if len(fields) == 0 {
		return 0
	}
	secs, _ := strconv.ParseFloat(fields[0], 64)
	return int64(secs * 100)
}

func read(path string) string {
	data, err := os.ReadFile(path) //nolint:gosec // /proc or test fixture path
	if err != nil {
		return ""
	}
	return string(data)
}

func appendFile(path, content string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644) //nolint:gosec // log file
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	_, err = f.WriteString(content)
	return err
}
