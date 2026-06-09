// Package last implements the last applet: show a listing of recent logins and
// logouts read from the wtmp database, most recent first.
package last

import (
	"context"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the last applet.
type Command struct{}

// New returns a last command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "last" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Show a listing of last logged-in users" }

// wtmpPath is the login-history file; tests point it at a fixture.
var wtmpPath = "/var/log/wtmp"

// Linux utmp/wtmp layout (x86_64): 384-byte records.
const (
	recordSize  = 384
	typeOffset  = 0
	lineOffset  = 8
	userOffset  = 44
	hostOffset  = 76
	tvSecOffset = 340
	fieldLen    = 32
	hostLen     = 256

	bootTime    = 2
	userProcess = 7
	deadProcess = 8
)

type session struct {
	user   string
	line   string
	host   string
	login  int64
	logout int64
	active bool
	reboot bool
}

// Run executes last.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-n COUNT] [FILE]", stdio.Err).WithHelp(command.Help{
		Description: "List the most recent logins and logouts from the wtmp database (or FILE), newest " +
			"first. Each line shows the user, terminal, remote host, and the login span; an open " +
			"session reads 'still logged in'.",
		Examples: []command.Example{
			{Command: "last", Explain: "Show the login history."},
			{Command: "last -n 10", Explain: "Show only the 10 most recent entries."},
		},
	})
	count := fs.IntP("count", "n", 0, "show only the last COUNT entries (0 = all)")
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	path := wtmpPath
	if rest := fs.Args(); len(rest) > 0 {
		path = rest[0]
	}

	data, err := os.ReadFile(path) //nolint:gosec // the wtmp database (or a user-named override)
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "last: %s\n", command.FileError(path, err))
		return command.SilentFailure()
	}

	sessions, first := build(data)
	shown := 0
	for i := len(sessions) - 1; i >= 0; i-- {
		if *count > 0 && shown >= *count {
			break
		}
		_, _ = fmt.Fprintln(stdio.Out, format(sessions[i]))
		shown++
	}
	if first > 0 {
		_, _ = fmt.Fprintf(stdio.Out, "\nwtmp begins %s\n", time.Unix(first, 0).Format("Mon Jan _2 15:04:05 2006"))
	}
	return nil
}

// build pairs login and logout records into sessions, in chronological order,
// and returns the earliest record time seen.
func build(data []byte) ([]session, int64) {
	var sessions []session
	active := map[string]int{} // line -> index in sessions
	var first int64

	for off := 0; off+recordSize <= len(data); off += recordSize {
		rec := data[off : off+recordSize]
		typ := int16(binary.LittleEndian.Uint16(rec[typeOffset:]))
		tv := int64(int32(binary.LittleEndian.Uint32(rec[tvSecOffset:])))
		if tv > 0 && (first == 0 || tv < first) {
			first = tv
		}
		line := cstr(rec[lineOffset : lineOffset+fieldLen])

		switch typ {
		case bootTime:
			sessions = append(sessions, session{user: "reboot", line: "system boot", host: cstr(rec[hostOffset : hostOffset+hostLen]), login: tv, reboot: true})
		case userProcess:
			sessions = append(sessions, session{
				user:   cstr(rec[userOffset : userOffset+fieldLen]),
				line:   line,
				host:   cstr(rec[hostOffset : hostOffset+hostLen]),
				login:  tv,
				active: true,
			})
			active[line] = len(sessions) - 1
		case deadProcess:
			if idx, ok := active[line]; ok {
				sessions[idx].logout = tv
				sessions[idx].active = false
				delete(active, line)
			}
		}
	}
	return sessions, first
}

// format renders one session line.
func format(s session) string {
	const layout = "Mon Jan _2 15:04"
	login := time.Unix(s.login, 0).Format(layout)
	if s.reboot {
		return fmt.Sprintf("%-8s %-12s %-16s %s", s.user, s.line, s.host, login)
	}
	host := s.host
	if host == "" {
		host = "-"
	}
	if s.active {
		return fmt.Sprintf("%-8s %-12s %-16s %s   still logged in", s.user, s.line, host, login)
	}
	logout := time.Unix(s.logout, 0).Format("15:04")
	dur := time.Duration(s.logout-s.login) * time.Second
	return fmt.Sprintf("%-8s %-12s %-16s %s - %s  (%02d:%02d)",
		s.user, s.line, host, login, logout, int(dur.Hours()), int(dur.Minutes())%60)
}

func cstr(b []byte) string {
	for i, x := range b {
		if x == 0 {
			return string(b[:i])
		}
	}
	return strings.TrimRight(string(b), "\x00")
}
