// Package w implements the w applet: show who is logged in, with a system
// summary header (time, uptime, user count, and load averages) followed by one
// row per active login.
package w

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

// Command is the w applet.
type Command struct{}

// New returns a w command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "w" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Show who is logged on and a system summary" }

// These are indirected so the output can be tested deterministically.
var (
	utmpPath    = "/var/run/utmp"
	loadavgPath = "/proc/loadavg"
	uptimePath  = "/proc/uptime"
	now         = time.Now
)

// Linux utmp layout (x86_64): 384-byte records; USER_PROCESS is type 7.
const (
	recordSize  = 384
	typeOffset  = 0
	lineOffset  = 8
	userOffset  = 44
	hostOffset  = 76
	tvSecOffset = 340
	fieldLen    = 32
	hostLen     = 256
	userProcess = 7
)

type entry struct {
	user  string
	line  string
	host  string
	login int64
}

// Run executes w.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "", stdio.Err).WithHelp(command.Help{
		Description: "Print a header with the current time, system uptime, the number of logged-in " +
			"users, and the load averages, then one line per login showing the user, terminal, " +
			"remote host, and login time.",
		Examples: []command.Example{
			{Command: "w", Explain: "Show who is logged in."},
		},
		Notes: []string{
			"The IDLE and WHAT columns are placeholders ('-'); per-process accounting is not implemented.",
		},
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	entries := readEntries(utmpPath)
	_, _ = fmt.Fprintln(stdio.Out, header(now(), readUptime(), readLoadavg(), len(entries)))
	_, _ = fmt.Fprintf(stdio.Out, "%-8s %-8s %-16s %-7s %-6s %s\n", "USER", "TTY", "FROM", "LOGIN@", "IDLE", "WHAT")
	for _, e := range entries {
		from := e.host
		if from == "" {
			from = "-"
		}
		login := time.Unix(e.login, 0).Format("15:04")
		_, _ = fmt.Fprintf(stdio.Out, "%-8s %-8s %-16s %-7s %-6s %s\n", e.user, e.line, from, login, "-", "-")
	}
	return nil
}

// header builds the summary line.
func header(t time.Time, uptime time.Duration, load string, users int) string {
	word := "users"
	if users == 1 {
		word = "user"
	}
	return fmt.Sprintf(" %s up %s, %d %s,  load average: %s",
		t.Format("15:04:05"), formatUptime(uptime), users, word, load)
}

// formatUptime renders a duration like the uptime/w tools do.
func formatUptime(d time.Duration) string {
	totalMin := int(d.Minutes())
	days := totalMin / (60 * 24)
	hours := (totalMin / 60) % 24
	mins := totalMin % 60
	if days > 0 {
		dayWord := "days"
		if days == 1 {
			dayWord = "day"
		}
		return fmt.Sprintf("%d %s, %2d:%02d", days, dayWord, hours, mins)
	}
	return fmt.Sprintf("%2d:%02d", hours, mins)
}

// readEntries parses the USER_PROCESS records from utmp.
func readEntries(path string) []entry {
	data, err := os.ReadFile(path) //nolint:gosec // the utmp database path
	if err != nil {
		return nil
	}
	var entries []entry
	for off := 0; off+recordSize <= len(data); off += recordSize {
		rec := data[off : off+recordSize]
		if int16(binary.LittleEndian.Uint16(rec[typeOffset:])) != userProcess {
			continue
		}
		entries = append(entries, entry{
			user:  cstr(rec[userOffset : userOffset+fieldLen]),
			line:  cstr(rec[lineOffset : lineOffset+fieldLen]),
			host:  cstr(rec[hostOffset : hostOffset+hostLen]),
			login: int64(int32(binary.LittleEndian.Uint32(rec[tvSecOffset:]))),
		})
	}
	return entries
}

// readUptime reads the system uptime from /proc/uptime, or 0 on failure.
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

// readLoadavg reads the three load averages from /proc/loadavg.
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

// cstr returns the NUL-terminated string at the start of b.
func cstr(b []byte) string {
	for i, x := range b {
		if x == 0 {
			return string(b[:i])
		}
	}
	return string(b)
}
