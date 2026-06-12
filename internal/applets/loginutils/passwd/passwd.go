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
	"strings"

	"github.com/GehirnInc/crypt"
	_ "github.com/GehirnInc/crypt/sha512_crypt" // register $6$

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the passwd applet.
type Command struct{}

// New returns a passwd command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "passwd" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Change a user's password" }

// Injected so the database and the current user are testable.
var (
	shadowPath    = "/etc/shadow"
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

	target := ""
	if rest := fs.Args(); len(rest) > 0 {
		target = rest[0]
	} else if target, err = currentUserFn(); err != nil {
		return command.Failuref("cannot determine the current user: %v", err)
	}

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
		hash, err := newHash(stdio)
		if err != nil {
			return command.Failuref("%v", err)
		}
		fields[1] = hash
	}

	lines[idx] = strings.Join(fields, ":")
	if err := writeLines(lines); err != nil {
		return command.Failuref("cannot write %s: %v", shadowPath, err)
	}
	_, _ = fmt.Fprintf(stdio.Err, "passwd: password for %q changed\n", target)
	return nil
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

// writeLines atomically replaces the shadow file.
func writeLines(lines []string) error {
	content := strings.Join(lines, "\n") + "\n"
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
	if err := tmp.Chmod(0o600); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, shadowPath)
}
