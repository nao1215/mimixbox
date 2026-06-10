// Package wall implements the wall applet: write a message to all logged-in
// users' terminals.
package wall

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"os/user"
	"strings"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the wall applet.
type Command struct{}

// New returns a wall command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "wall" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Write a message to all logged-in users" }

// Injected so the broadcast can be tested without a terminal or real logins.
var (
	utmpPath = "/var/run/utmp"
	now      = time.Now
	sender   = defaultSender
	writeTTY = func(line, text string) error {
		f, err := os.OpenFile("/dev/"+line, os.O_WRONLY, 0) //nolint:gosec // a terminal device
		if err != nil {
			return err
		}
		defer func() { _ = f.Close() }()
		_, err = io.WriteString(f, text)
		return err
	}
)

// Linux utmp layout: 384-byte records; USER_PROCESS is type 7.
const (
	recordSize  = 384
	typeOffset  = 0
	lineOffset  = 8
	fieldLen    = 32
	userProcess = 7
)

// Run executes wall.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[MESSAGE]", stdio.Err).WithHelp(command.Help{
		Description: "Write MESSAGE (or standard input when no MESSAGE is given) to the terminals of " +
			"all logged-in users, prefixed with a 'Broadcast message' banner. Terminals that cannot " +
			"be written are skipped.",
		Examples: []command.Example{
			{Command: "wall 'system going down in 5 minutes'", Explain: "Broadcast a warning."},
			{Command: "echo done | wall", Explain: "Broadcast a message from standard input."},
		},
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	var message string
	if rest := fs.Args(); len(rest) > 0 {
		message = strings.Join(rest, " ")
	} else {
		data, _ := io.ReadAll(stdio.In)
		message = strings.TrimRight(string(data), "\n")
	}

	banner := c.banner(message)
	for _, line := range loggedInLines(utmpPath) {
		_ = writeTTY(line, banner) // unwritable terminals are silently skipped
	}
	return nil
}

// banner builds the broadcast text.
func (c *Command) banner(message string) string {
	who, host := sender()
	stamp := now().Format("Mon Jan _2 15:04:05 2006")
	return fmt.Sprintf("\r\nBroadcast message from %s@%s (%s):\r\n\r\n%s\r\n", who, host, stamp, message)
}

// loggedInLines returns the terminal lines of every USER_PROCESS in utmp.
func loggedInLines(path string) []string {
	data, err := os.ReadFile(path) //nolint:gosec // the utmp path
	if err != nil {
		return nil
	}
	var lines []string
	for off := 0; off+recordSize <= len(data); off += recordSize {
		rec := data[off : off+recordSize]
		if int16(binary.LittleEndian.Uint16(rec[typeOffset:])) != userProcess {
			continue
		}
		if line := cstr(rec[lineOffset : lineOffset+fieldLen]); line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

func defaultSender() (string, string) {
	who := "root"
	if u, err := user.Current(); err == nil {
		who = u.Username
	}
	host, err := os.Hostname()
	if err != nil {
		host = "localhost"
	}
	return who, host
}

func cstr(b []byte) string {
	for i, x := range b {
		if x == 0 {
			return string(b[:i])
		}
	}
	return string(b)
}
