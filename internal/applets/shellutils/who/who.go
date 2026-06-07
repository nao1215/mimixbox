// Package who implements the who applet: show who is logged on by reading the
// Linux utmp login-record database (/var/run/utmp by default). The record
// layout matches Linux's struct utmp (384 bytes), and the parsing and
// formatting are split into pure functions so they can be exercised in memory
// without a real login session.
package who

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
)

// utmpFile is the login-record database read by default. It is a package
// variable so tests can point it at a fixture file.
var utmpFile = "/var/run/utmp"

// utmpRecordSize is the size in bytes of a Linux struct utmp record.
const utmpRecordSize = 384

// ut_type values from <utmp.h>. Only the ones who cares about are listed.
const (
	bootTime    = 2 // BOOT_TIME:    time of system boot
	userProcess = 7 // USER_PROCESS: a normal logged-in user process
)

// Field offsets within a struct utmp record (see the package doc).
const (
	offType = 0   // ut_type   int16
	offLine = 8   // ut_line   char[32]
	offUser = 44  // ut_user   char[32]
	offHost = 76  // ut_host   char[256]
	offSec  = 340 // ut_tv.tv_sec int32
)

// Entry is one parsed utmp record, holding only the fields who reports.
type Entry struct {
	Type int16     // ut_type
	User string    // ut_user, NUL-trimmed
	Line string    // ut_line, NUL-trimmed
	Host string    // ut_host, NUL-trimmed
	Time time.Time // login/boot time from ut_tv.tv_sec
}

// Command is the who applet.
type Command struct{}

// New returns a who command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "who" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Show who is logged on" }

type options struct {
	boot    bool // -b, --boot:    print the system boot time
	heading bool // -H, --heading: print a column heading line
	count   bool // -q, --count:   print only login names and a user count
	idle    bool // -u:            also report idle time and PID (best-effort)
}

// Run executes who. With no options it reads the utmp database and prints one
// line per logged-in user. An optional FILE operand overrides utmpFile.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... [FILE]", stdio.Err)
	boot := fs.BoolP("boot", "b", false, "time of last system boot")
	heading := fs.BoolP("heading", "H", false, "print line of column headings")
	count := fs.BoolP("count", "q", false, "all login names and number of users logged on")
	idle := fs.BoolP("idle", "u", false, "list users logged in")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	opts := options{boot: *boot, heading: *heading, count: *count, idle: *idle}

	path := utmpFile
	if operands := fs.Args(); len(operands) > 0 {
		path = operands[0]
	}

	data, rerr := os.ReadFile(path) //nolint:gosec // reading a user-named utmp file is the whole point
	if rerr != nil {
		_, _ = fmt.Fprintf(stdio.Err, "who: %s\n", command.FileError(path, rerr))
		return command.SilentFailure()
	}

	entries, perr := parseUtmp(bytes.NewReader(data))
	if perr != nil {
		_, _ = fmt.Fprintf(stdio.Err, "who: %v\n", perr)
		return command.SilentFailure()
	}

	_, _ = io.WriteString(stdio.Out, render(entries, opts))
	return nil
}

// parseUtmp reads fixed-size struct utmp records from r and returns them as
// Entry values. A trailing short read (fewer than utmpRecordSize bytes) is
// treated as end of data rather than an error, matching how the file may be
// truncated. It is a pure function so tests need no /var/run/utmp.
func parseUtmp(r io.Reader) ([]Entry, error) {
	var entries []Entry
	rec := make([]byte, utmpRecordSize)
	for {
		n, err := io.ReadFull(r, rec)
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if n < utmpRecordSize {
			break
		}
		entries = append(entries, Entry{
			Type: int16(binary.LittleEndian.Uint16(rec[offType:])),
			User: cString(rec[offUser : offUser+32]),
			Line: cString(rec[offLine : offLine+32]),
			Host: cString(rec[offHost : offHost+256]),
			Time: time.Unix(int64(int32(binary.LittleEndian.Uint32(rec[offSec:]))), 0),
		})
	}
	return entries, nil
}

// cString returns the bytes up to the first NUL as a string.
func cString(b []byte) string {
	if i := bytes.IndexByte(b, 0); i >= 0 {
		return string(b[:i])
	}
	return string(b)
}

// render produces who's whole output for the given entries and options.
func render(entries []Entry, opts options) string {
	if opts.count {
		return renderCount(entries)
	}

	var b bytes.Buffer
	if opts.heading {
		b.WriteString(heading())
	}
	if opts.boot {
		for _, e := range entries {
			if e.Type == bootTime {
				b.WriteString(formatBoot(e))
			}
		}
		return b.String()
	}
	for _, e := range entries {
		if e.Type == userProcess {
			b.WriteString(formatUser(e, opts))
		}
	}
	return b.String()
}

// heading returns the column heading line printed by -H.
func heading() string {
	return fmt.Sprintf("%-8s %-12s %s\n", "NAME", "LINE", "TIME")
}

// formatUser formats one USER_PROCESS record GNU-style: "NAME LINE TIME".
func formatUser(e Entry, opts options) string {
	line := fmt.Sprintf("%-8s %-12s %s", e.User, e.Line, formatTime(e.Time))
	if opts.idle {
		line += fmt.Sprintf("   %s", e.Host)
	}
	return line + "\n"
}

// formatBoot formats a BOOT_TIME record GNU-style as "system boot  TIME".
func formatBoot(e Entry) string {
	return fmt.Sprintf("%-12s %s\n", "system boot", formatTime(e.Time))
}

// renderCount implements -q: the login names on one line, then "# users=N".
func renderCount(entries []Entry) string {
	var b bytes.Buffer
	count := 0
	for _, e := range entries {
		if e.Type != userProcess {
			continue
		}
		if count > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(e.User)
		count++
	}
	if count > 0 {
		b.WriteByte('\n')
	}
	fmt.Fprintf(&b, "# users=%d\n", count)
	return b.String()
}

// formatTime renders a login/boot time the way GNU who does, as
// "YYYY-MM-DD HH:MM".
func formatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04")
}
