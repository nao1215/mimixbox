// Package uptime implements the uptime applet: show how long the system has been
// running, the number of logged-in users, and the load averages.
package uptime

import (
	"context"
	"encoding/binary"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the uptime applet.
type Command struct{}

// New returns an uptime command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "uptime" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Tell how long the system has been running" }

// Injected so the output can be tested deterministically.
var (
	uptimePath  = "/proc/uptime"
	loadavgPath = "/proc/loadavg"
	utmpPath    = "/var/run/utmp"
	now         = time.Now
)

// utmp layout: 384-byte records; USER_PROCESS is type 7.
const (
	recordSize  = 384
	userProcess = 7
)

// Run executes uptime.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "", stdio.Err).WithHelp(command.Help{
		Description: "Print the current time, how long the system has been up, the number of users " +
			"currently logged in, and the load averages for the past 1, 5, and 15 minutes.",
		Examples: []command.Example{
			{Command: "uptime", Explain: "Show the uptime summary."},
		},
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	_, _ = fmt.Fprintln(stdio.Out, line(now(), readUptime(), readLoadavg(), countUsers()))
	return nil
}

// line builds the uptime summary, matching the procps layout.
func line(t time.Time, up time.Duration, load string, users int) string {
	word := "users"
	if users == 1 {
		word = "user"
	}
	return fmt.Sprintf(" %s up %s,  %d %s,  load average: %s",
		t.Format("15:04:05"), formatUptime(up), users, word, load)
}

// formatUptime renders the uptime like procps does.
func formatUptime(d time.Duration) string {
	totalMin := int(d.Minutes())
	days := totalMin / (60 * 24)
	hours := (totalMin / 60) % 24
	mins := totalMin % 60
	switch {
	case days > 1:
		return fmt.Sprintf("%d days, %2d:%02d", days, hours, mins)
	case days == 1:
		return fmt.Sprintf("1 day, %2d:%02d", hours, mins)
	default:
		return fmt.Sprintf("%2d:%02d", hours, mins)
	}
}

func readUptime() time.Duration {
	data, err := os.ReadFile(uptimePath)
	if err != nil {
		return 0
	}
	fields := strings.Fields(string(data))
	if len(fields) == 0 {
		return 0
	}
	secs, _ := strconv.ParseFloat(fields[0], 64)
	return time.Duration(secs) * time.Second
}

func readLoadavg() string {
	data, err := os.ReadFile(loadavgPath)
	if err != nil {
		return "0.00, 0.00, 0.00"
	}
	fields := strings.Fields(string(data))
	if len(fields) < 3 {
		return "0.00, 0.00, 0.00"
	}
	return fmt.Sprintf("%s, %s, %s", fields[0], fields[1], fields[2])
}

func countUsers() int {
	data, err := os.ReadFile(utmpPath)
	if err != nil {
		return 0
	}
	n := 0
	for off := 0; off+recordSize <= len(data); off += recordSize {
		if int16(binary.LittleEndian.Uint16(data[off:])) == userProcess {
			n++
		}
	}
	return n
}
