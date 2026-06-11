// Package chpasswd implements the chpasswd applet: update user passwords in
// batch from "user:password" lines on standard input.
package chpasswd

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/GehirnInc/crypt"
	_ "github.com/GehirnInc/crypt/md5_crypt"    // register $1$
	_ "github.com/GehirnInc/crypt/sha256_crypt" // register $5$
	_ "github.com/GehirnInc/crypt/sha512_crypt" // register $6$

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the chpasswd applet.
type Command struct{}

// New returns a chpasswd command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "chpasswd" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Update passwords in batch" }

// shadowPath is the shadow database; tests point it at a fixture.
var shadowPath = "/etc/shadow"

var cryptMethods = map[string]crypt.Crypt{
	"sha-512": crypt.SHA512, "sha512": crypt.SHA512,
	"sha-256": crypt.SHA256, "sha256": crypt.SHA256,
	"md5": crypt.MD5,
}

// Run executes chpasswd.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-e] [-c METHOD]", stdio.Err).WithHelp(command.Help{
		Description: "Read 'user:password' lines from standard input and set each user's password in " +
			"/etc/shadow. Passwords are hashed with -c METHOD (sha-512 by default); with -e the " +
			"supplied values are already-encrypted hashes and are stored verbatim. Requires privilege " +
			"to write the real shadow database.",
		Examples: []command.Example{
			{Command: "echo 'alice:secret' | chpasswd", Explain: "Set alice's password."},
			{Command: "chpasswd -e < hashes.txt", Explain: "Load already-hashed passwords."},
		},
		ExitStatus: "0  all passwords were updated.\n1  a user was unknown or the database is unreadable.",
	})
	encrypted := fs.BoolP("encrypted", "e", false, "the supplied passwords are already hashed")
	methodName := fs.StringP("crypt-method", "c", "sha-512", "hashing method: sha-512, sha-256, or md5")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	method, ok := cryptMethods[strings.ToLower(*methodName)]
	if !ok {
		return command.Failuref("unknown method: %q (use sha-512, sha-256, or md5)", *methodName)
	}

	updates, err := readUpdates(stdio.In, *encrypted, method)
	if err != nil {
		return command.Failuref("%v", err)
	}

	lines, err := readLines(shadowPath)
	if err != nil {
		return command.Failuref("cannot read %s: %v", shadowPath, err)
	}

	applied := map[string]bool{}
	for i, line := range lines {
		fields := strings.Split(line, ":")
		if len(fields) < 2 {
			continue
		}
		if hash, ok := updates[fields[0]]; ok {
			fields[1] = hash
			lines[i] = strings.Join(fields, ":")
			applied[fields[0]] = true
		}
	}

	failed := false
	for user := range updates {
		if !applied[user] {
			_, _ = fmt.Fprintf(stdio.Err, "chpasswd: unknown user: %s\n", user)
			failed = true
		}
	}

	if err := writeLines(shadowPath, lines); err != nil {
		return command.Failuref("cannot write %s: %v", shadowPath, err)
	}
	if failed {
		return command.SilentFailure()
	}
	return nil
}

// readUpdates parses the "user:password" lines into a map of user to the hash
// (or verbatim value when already encrypted) to store.
func readUpdates(r io.Reader, encrypted bool, method crypt.Crypt) (map[string]string, error) {
	updates := map[string]string{}
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := strings.TrimRight(sc.Text(), "\r\n")
		if line == "" {
			continue
		}
		user, password, ok := strings.Cut(line, ":")
		if !ok || user == "" {
			return nil, fmt.Errorf("malformed input line (need user:password): %q", line)
		}
		if encrypted {
			updates[user] = password
			continue
		}
		hash, err := crypt.New(method).Generate([]byte(password), nil)
		if err != nil {
			return nil, fmt.Errorf("cannot hash password for %s: %v", user, err)
		}
		updates[user] = hash
	}
	return updates, sc.Err()
}

func readLines(path string) ([]string, error) {
	data, err := os.ReadFile(path) //nolint:gosec // well-known shadow path
	if err != nil {
		return nil, err
	}
	trimmed := strings.TrimRight(string(data), "\n")
	if trimmed == "" {
		return nil, nil
	}
	return strings.Split(trimmed, "\n"), nil
}

// writeLines atomically replaces path: it writes a temporary file in the same
// directory and renames it into place, so an interrupted write cannot leave the
// shadow database truncated or corrupted.
func writeLines(path string, lines []string) error {
	content := strings.Join(lines, "\n") + "\n"
	tmp, err := os.CreateTemp(filepath.Dir(path), ".chpasswd-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }() // no-op once the rename succeeds

	if _, err := tmp.WriteString(content); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Chmod(0o600); err != nil { //nolint:gosec // shadow is mode 0600
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}
