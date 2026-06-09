// Package ipcs implements the ipcs applet: report status of System V
// inter-process communication facilities (message queues, shared memory, and
// semaphore arrays), read from /proc/sysvipc.
package ipcs

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the ipcs applet.
type Command struct{}

// New returns an ipcs command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "ipcs" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Show System V IPC facilities status" }

// procDir is the sysvipc directory; tests point it at a fixture.
var procDir = "/proc/sysvipc"

// Run executes ipcs.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-q] [-m] [-s]", stdio.Err).WithHelp(command.Help{
		Description: "Show information about the active System V IPC objects. With no option all three " +
			"facilities are shown; -q limits to message queues, -m to shared memory, and -s to " +
			"semaphore arrays.",
		Examples: []command.Example{
			{Command: "ipcs", Explain: "Show all IPC objects."},
			{Command: "ipcs -m", Explain: "Show only shared memory segments."},
		},
	})
	queues := fs.BoolP("queues", "q", false, "message queues")
	shmem := fs.BoolP("shmems", "m", false, "shared memory segments")
	sems := fs.BoolP("semaphores", "s", false, "semaphore arrays")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	all := !*queues && !*shmem && !*sems
	if all || *queues {
		c.messageQueues(stdio.Out)
	}
	if all || *shmem {
		c.sharedMemory(stdio.Out)
	}
	if all || *sems {
		c.semaphores(stdio.Out)
	}
	return nil
}

func (c *Command) messageQueues(out io.Writer) {
	_, _ = fmt.Fprintln(out, "\n------ Message Queues --------")
	_, _ = fmt.Fprintf(out, "%-12s %-10s %-10s %-6s %-12s %-10s\n", "key", "msqid", "owner", "perms", "used-bytes", "messages")
	for _, f := range records(filepath.Join(procDir, "msg")) {
		if len(f) < 8 {
			continue
		}
		_, _ = fmt.Fprintf(out, "%-12s %-10s %-10s %-6s %-12s %-10s\n",
			key(f[0]), f[1], owner(f[7]), f[2], f[3], f[4])
	}
}

func (c *Command) sharedMemory(out io.Writer) {
	_, _ = fmt.Fprintln(out, "\n------ Shared Memory Segments --------")
	_, _ = fmt.Fprintf(out, "%-12s %-10s %-10s %-6s %-12s %-8s\n", "key", "shmid", "owner", "perms", "bytes", "nattch")
	for _, f := range records(filepath.Join(procDir, "shm")) {
		if len(f) < 8 {
			continue
		}
		_, _ = fmt.Fprintf(out, "%-12s %-10s %-10s %-6s %-12s %-8s\n",
			key(f[0]), f[1], owner(f[7]), f[2], f[3], f[6])
	}
}

func (c *Command) semaphores(out io.Writer) {
	_, _ = fmt.Fprintln(out, "\n------ Semaphore Arrays --------")
	_, _ = fmt.Fprintf(out, "%-12s %-10s %-10s %-6s %-8s\n", "key", "semid", "owner", "perms", "nsems")
	for _, f := range records(filepath.Join(procDir, "sem")) {
		if len(f) < 5 {
			continue
		}
		_, _ = fmt.Fprintf(out, "%-12s %-10s %-10s %-6s %-8s\n",
			key(f[0]), f[1], owner(f[4]), f[2], f[3])
	}
}

// records reads a /proc/sysvipc file, skips its header line, and returns the
// whitespace-split fields of each remaining line.
func records(path string) [][]string {
	f, err := os.Open(path) //nolint:gosec // the sysvipc path
	if err != nil {
		return nil
	}
	defer func() { _ = f.Close() }()

	var out [][]string
	sc := bufio.NewScanner(f)
	first := true
	for sc.Scan() {
		if first { // header row
			first = false
			continue
		}
		fields := strings.Fields(sc.Text())
		if len(fields) > 0 {
			out = append(out, fields)
		}
	}
	return out
}

// key formats a decimal key field as ipcs does (0x%08x).
func key(s string) string {
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return s
	}
	return fmt.Sprintf("0x%08x", uint32(n))
}

var userCache = map[string]string{}

// owner resolves a uid string to a user name.
func owner(uid string) string {
	if name, ok := userCache[uid]; ok {
		return name
	}
	name := uid
	if u, err := user.LookupId(uid); err == nil {
		name = u.Username
	}
	userCache[uid] = name
	return name
}
