// Package users implements the users applet: print the login names of the users
// currently on the system, one space-separated line, read from the utmp
// database.
package users

import (
	"context"
	"encoding/binary"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the users applet.
type Command struct{}

// New returns a users command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "users" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Print the user names of those currently logged in" }

// utmpPath is the login record file; tests point it at a fixture.
var utmpPath = "/var/run/utmp"

// Linux struct utmp layout (x86_64): 384 bytes per record, ut_type at offset 0,
// ut_user (32 bytes) at offset 44. USER_PROCESS is type 7.
const (
	recordSize  = 384
	typeOffset  = 0
	userOffset  = 44
	userLen     = 32
	userProcess = 7
)

// Run executes users.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[FILE]", stdio.Err).WithHelp(command.Help{
		Description: "Print, on a single space-separated line, the login name of every user currently " +
			"logged in (one entry per active login). FILE overrides the default utmp database.",
		Examples: []command.Example{
			{Command: "users", Explain: "List the currently logged-in users."},
		},
		ExitStatus: "0  success.\n1  an error occurred.",
		Notes: []string{
			"Reads the Linux utmp format; a login appears once per session.",
		},
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	path := utmpPath
	if rest := fs.Args(); len(rest) > 0 {
		path = rest[0]
	}

	data, err := os.ReadFile(path) //nolint:gosec // the utmp path (or a user-named override)
	if err != nil {
		// A missing utmp simply means nobody is recorded as logged in.
		_, _ = fmt.Fprintln(stdio.Out)
		return nil
	}

	names := parse(data)
	sort.Strings(names)
	_, _ = fmt.Fprintln(stdio.Out, strings.Join(names, " "))
	return nil
}

// parse extracts the user name of every USER_PROCESS record in utmp data.
func parse(data []byte) []string {
	var names []string
	for off := 0; off+recordSize <= len(data); off += recordSize {
		rec := data[off : off+recordSize]
		if int16(binary.LittleEndian.Uint16(rec[typeOffset:])) != userProcess {
			continue
		}
		if name := cstr(rec[userOffset : userOffset+userLen]); name != "" {
			names = append(names, name)
		}
	}
	return names
}

// cstr returns the NUL-terminated string at the start of b.
func cstr(b []byte) string {
	if i := indexByte(b, 0); i >= 0 {
		return string(b[:i])
	}
	return string(b)
}

func indexByte(b []byte, c byte) int {
	for i, x := range b {
		if x == c {
			return i
		}
	}
	return -1
}
