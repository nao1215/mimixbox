// Package crond implements the crond applet: a foreground cron daemon that runs
// the jobs in the cron spool directory at their scheduled times.
package crond

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the crond applet.
type Command struct{}

// New returns a crond command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "crond" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Run scheduled cron jobs (foreground)" }

// Injected so the spool, the clock, and job execution are testable.
var (
	spoolDir = "/var/spool/cron/crontabs"
	now      = time.Now
	runFn    = func(cmdline string) {
		cmd := exec.Command("sh", "-c", cmdline) //nolint:gosec // running scheduled jobs is the point
		_ = cmd.Start()
	}
)

// Run executes crond.
func (c *Command) Run(ctx context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "-f", stdio.Err).WithHelp(command.Help{
		Description: "Run cron jobs from the spool directory at their scheduled times. Only the " +
			"foreground mode (-f) is supported: crond stays in the foreground, checks every minute, " +
			"and runs each crontab entry whose schedule matches the current time, until interrupted. " +
			"Each crontab line is 'minute hour day-of-month month day-of-week command'.",
		Examples: []command.Example{
			{Command: "crond -f", Explain: "Run the cron daemon in the foreground."},
		},
		ExitStatus: "0  the daemon stopped cleanly.\n1  -f was not given.",
	})
	foreground := fs.BoolP("foreground", "f", false, "run in the foreground (the only supported mode)")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	if !*foreground {
		_, _ = fmt.Fprintln(stdio.Err, "crond: only the foreground mode (-f) is supported by this build")
		return command.SilentFailure()
	}

	for {
		next := now().Truncate(time.Minute).Add(time.Minute)
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(time.Until(next)):
			runDue(loadEntries(), now(), runFn)
		}
	}
}

// entry is one parsed crontab schedule and its command.
type entry struct {
	minute, hour, dom, month, dow string
	command                       string
}

// loadEntries reads and parses every crontab in the spool directory.
func loadEntries() []entry {
	files, err := os.ReadDir(spoolDir)
	if err != nil {
		return nil
	}
	var entries []entry
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		data, err := os.ReadFile(filepath.Join(spoolDir, f.Name())) //nolint:gosec // spool path
		if err != nil {
			continue
		}
		entries = append(entries, parseEntries(string(data))...)
	}
	return entries
}

// parseEntries parses crontab text into schedule entries, skipping blank lines,
// comments, and environment assignments.
func parseEntries(text string) []entry {
	var entries []entry
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 6 || strings.Contains(fields[0], "=") {
			continue
		}
		entries = append(entries, entry{
			minute:  fields[0],
			hour:    fields[1],
			dom:     fields[2],
			month:   fields[3],
			dow:     fields[4],
			command: strings.Join(fields[5:], " "),
		})
	}
	return entries
}

// runDue runs every entry whose schedule matches t.
func runDue(entries []entry, t time.Time, run func(string)) {
	for _, e := range entries {
		if e.matches(t) {
			run(e.command)
		}
	}
}

// matches reports whether the entry's schedule fires at time t. Following Vixie
// cron, when both day-of-month and day-of-week are restricted the entry fires if
// either matches; otherwise all fields must match.
func (e entry) matches(t time.Time) bool {
	dowVal := int(t.Weekday()) // 0 = Sunday
	if !matchField(e.minute, t.Minute(), 0, 59) ||
		!matchField(e.hour, t.Hour(), 0, 23) ||
		!matchField(e.month, int(t.Month()), 1, 12) {
		return false
	}
	domRestricted := e.dom != "*"
	dowRestricted := e.dow != "*"
	domOK := matchField(e.dom, t.Day(), 1, 31)
	dowOK := matchDow(e.dow, dowVal)
	if domRestricted && dowRestricted {
		return domOK || dowOK
	}
	return domOK && dowOK
}

// matchDow matches a day-of-week field, treating 7 as Sunday (0).
func matchDow(spec string, value int) bool {
	if matchField(spec, value, 0, 7) {
		return true
	}
	if value == 0 { // Sunday may also be written as 7
		return matchField(spec, 7, 0, 7)
	}
	return false
}

// matchField reports whether value satisfies a single cron field (supporting
// '*', 'N', 'N-M', 'STEP' forms, and comma-separated lists).
func matchField(spec string, value, min, max int) bool {
	for _, part := range strings.Split(spec, ",") {
		if matchPart(part, value, min, max) {
			return true
		}
	}
	return false
}

func matchPart(part string, value, min, max int) bool {
	step := 1
	if slash := strings.IndexByte(part, '/'); slash >= 0 {
		s, err := strconv.Atoi(part[slash+1:])
		if err != nil || s <= 0 {
			return false
		}
		step = s
		part = part[:slash]
	}

	lo, hi := min, max
	switch {
	case part == "*":
		// full range
	case strings.ContainsRune(part, '-'):
		bounds := strings.SplitN(part, "-", 2)
		a, err1 := strconv.Atoi(bounds[0])
		b, err2 := strconv.Atoi(bounds[1])
		if err1 != nil || err2 != nil {
			return false
		}
		lo, hi = a, b
	default:
		n, err := strconv.Atoi(part)
		if err != nil {
			return false
		}
		lo, hi = n, n
	}

	if value < lo || value > hi {
		return false
	}
	return (value-lo)%step == 0
}
