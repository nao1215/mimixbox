// Package passwd implements the passwd applet: change, lock, unlock, or clear a
// user's password in /etc/shadow.
package passwd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/GehirnInc/crypt"
	_ "github.com/GehirnInc/crypt/sha512_crypt" // register $6$

	"github.com/nao1215/mimixbox/internal/command"
	"golang.org/x/sys/unix"
)

// Command is the passwd applet.
type Command struct{}

// New returns a passwd command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "passwd" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Change a user's password" }

const secondsPerDay = 86400

// Injected so the database, the current user, and the clock are testable.
var (
	shadowPath    = "/etc/shadow"
	now           = time.Now
	currentUserFn = func() (string, error) {
		u, err := user.Current()
		if err != nil {
			return "", err
		}
		return u.Username, nil
	}
)

// Run executes passwd.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-l|-u|-d] [USER]", stdio.Err).WithHelp(command.Help{
		Description: "Change the password of USER (the current user by default) in /etc/shadow. With no " +
			"flag the new password is read from standard input (and, if a second line is present, must " +
			"match it) and stored as a sha-512 hash. -l locks the account, -u unlocks it, and -d clears " +
			"its password. Changing the real database requires privilege.",
		Examples: []command.Example{
			{Command: "echo newpass | passwd alice", Explain: "Set alice's password."},
			{Command: "passwd -l bob", Explain: "Lock bob's account."},
		},
		ExitStatus: "0  the password was changed.\n1  the user is unknown or the database is unwritable.",
	})
	lock := fs.BoolP("lock", "l", false, "lock the account")
	unlock := fs.BoolP("unlock", "u", false, "unlock the account")
	del := fs.BoolP("delete", "d", false, "clear the password")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	if countTrue(*lock, *unlock, *del) > 1 {
		return command.Failuref("-l, -u, and -d are mutually exclusive")
	}

	rest := fs.Args()
	if len(rest) > 1 {
		return command.Failuref("at most one user may be given")
	}
	target := ""
	if len(rest) == 1 {
		target = rest[0]
	} else if target, err = currentUserFn(); err != nil {
		return command.Failuref("cannot determine the current user: %v", err)
	}

	// Compute the replacement hash before taking the lock so an empty/mismatched
	// password fails without touching the database.
	var newPassword string
	if !*lock && !*unlock && !*del {
		if newPassword, err = newHash(stdio); err != nil {
			return command.Failuref("%v", err)
		}
	}

	return withLock(shadowPath, func() error {
		lines, idx, fields, err := findUser(target)
		if err != nil {
			return command.Failuref("%v", err)
		}
		switch {
		case *lock:
			if !strings.HasPrefix(fields[1], "!") {
				fields[1] = "!" + fields[1]
			}
		case *unlock:
			fields[1] = strings.TrimPrefix(fields[1], "!")
		case *del:
			fields[1] = ""
		default:
			fields[1] = newPassword
			setLastChange(fields)
		}
		lines[idx] = strings.Join(fields, ":")
		if err := writeLines(lines); err != nil {
			return command.Failuref("cannot write %s: %v", shadowPath, err)
		}
		_, _ = fmt.Fprintf(stdio.Err, "passwd: password for %q changed\n", target)
		return nil
	})
}

// setLastChange records today's date (days since the epoch) in the shadow
// last-change field, so password-aging consumers see the update.
func setLastChange(fields []string) {
	if len(fields) > 2 {
		fields[2] = strconv.FormatInt(now().Unix()/secondsPerDay, 10)
	}
}

func countTrue(bs ...bool) int {
	n := 0
	for _, b := range bs {
		if b {
			n++
		}
	}
	return n
}

// withLock runs fn while holding an exclusive advisory lock, serializing
// concurrent password changes. The lock is taken on a dedicated, stable lockfile
// ("<path>.lock") rather than on the shadow file itself: writeLines replaces the
// shadow file's inode via rename, so a lock held on that inode would not be seen
// by a concurrent run. If the lockfile cannot be created (e.g. an unprivileged
// run), fn proceeds without the lock.
func withLock(path string, fn func() error) error {
	f, err := os.OpenFile(path+".lock", os.O_CREATE|os.O_RDWR, 0o600) //nolint:gosec // adjacent to the shadow path
	if err != nil {
		return fn()
	}
	defer func() { _ = f.Close() }()
	if unix.Flock(int(f.Fd()), unix.LOCK_EX) == nil {
		defer func() { _ = unix.Flock(int(f.Fd()), unix.LOCK_UN) }()
	}
	return fn()
}

// findUser returns the shadow lines, the index of target's line, and that
// line's colon fields.
func findUser(target string) (lines []string, idx int, fields []string, err error) {
	data, err := os.ReadFile(shadowPath) //nolint:gosec // well-known shadow path
	if err != nil {
		return nil, 0, nil, fmt.Errorf("cannot read %s: %v", shadowPath, err)
	}
	trimmed := strings.TrimRight(string(data), "\n")
	if trimmed != "" {
		lines = strings.Split(trimmed, "\n")
	}
	for i, line := range lines {
		f := strings.Split(line, ":")
		if len(f) >= 2 && f[0] == target {
			return lines, i, f, nil
		}
	}
	return nil, 0, nil, fmt.Errorf("user %q does not exist", target)
}

// newHash reads the new password (and an optional confirmation) from stdin and
// returns its sha-512 crypt hash.
func newHash(stdio command.IO) (string, error) {
	sc := bufio.NewScanner(stdio.In)
	if !sc.Scan() {
		return "", fmt.Errorf("no password provided")
	}
	password := sc.Text()
	if password == "" {
		return "", fmt.Errorf("empty password is not allowed")
	}
	if sc.Scan() && sc.Text() != password {
		return "", fmt.Errorf("passwords do not match")
	}
	hash, err := crypt.New(crypt.SHA512).Generate([]byte(password), nil)
	if err != nil {
		return "", fmt.Errorf("cannot hash the password: %v", err)
	}
	return hash, nil
}

// writeLines atomically replaces the shadow file, preserving its existing mode
// and ownership.
func writeLines(lines []string) error {
	content := strings.Join(lines, "\n") + "\n"
	mode := os.FileMode(0o600)
	var uid, gid = -1, -1
	if info, err := os.Stat(shadowPath); err == nil {
		mode = info.Mode().Perm()
		if st, ok := info.Sys().(*syscall.Stat_t); ok {
			uid, gid = int(st.Uid), int(st.Gid)
		}
	}

	tmp, err := os.CreateTemp(filepath.Dir(shadowPath), ".passwd-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }()
	if _, err := tmp.WriteString(content); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Chmod(mode); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if uid >= 0 {
		_ = os.Chown(tmpName, uid, gid) // best effort: keep the original owner
	}
	return os.Rename(tmpName, shadowPath)
}
