// Package chsh implements the chsh applet: change a user's login shell by
// rewriting the seventh field of their /etc/passwd entry.
//
// MimixBox does not perform PAM authentication. chsh relies on filesystem
// permissions on the passwd database instead: rewriting another user's entry,
// or any entry at all, requires write access to /etc/passwd (normally root).
// This is the "Linux without PAM" model, and it is the only model MimixBox
// supports today; there is no separate PAM-backed path. On platforms without an
// /etc/passwd database the command fails explicitly rather than pretending to
// succeed.
package chsh

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
	mb "github.com/nao1215/mimixbox/internal/lib"
)

// passwdPath is the user database chsh rewrites; tests point it at a fixture.
var passwdPath = "/etc/passwd"

// shellsPath lists the shells a non-privileged user may choose; tests override it.
var shellsPath = mb.ShellsFilePath

// geteuid is indirected so tests can exercise the privileged and unprivileged
// validation paths deterministically.
var geteuid = os.Geteuid

// Command is the chsh applet.
type Command struct{}

// New returns a chsh command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "chsh" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Change a user's login shell" }

// Run executes chsh.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-s SHELL] [-l] [USER]", stdio.Err).WithHelp(command.Help{
		Description: "Change the login shell recorded for a user in /etc/passwd. With -s SHELL the " +
			"shell is set non-interactively; otherwise chsh reads the new shell from standard input. " +
			"With no USER the current user's shell is changed. Use -l to list the shells in /etc/shells " +
			"and exit. A non-privileged user may only select a shell listed in /etc/shells; the " +
			"superuser may set any path.",
		Examples: []command.Example{
			{Command: "chsh -s /bin/bash", Explain: "Set the current user's shell to /bin/bash."},
			{Command: "chsh -s /bin/sh alice", Explain: "Set alice's shell (needs write access to /etc/passwd)."},
			{Command: "chsh -l", Explain: "List the shells available in /etc/shells."},
		},
		ExitStatus: "0  the shell was changed (or listed) successfully.\n" +
			"1  the user was unknown, the shell was rejected, or the database could not be written.",
		Notes: []string{
			"MimixBox does not use PAM; authorization is by filesystem permission on /etc/passwd.",
			"Changing another user's shell, or running as a non-owner, requires privilege (usually root).",
		},
	})
	shell := fs.StringP("shell", "s", "", "login shell to set; if empty, read it from standard input")
	list := fs.BoolP("list-shells", "l", false, "list the shells in /etc/shells and exit")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	if *list {
		return listShells(stdio)
	}

	username, err := targetUser(fs.Args())
	if err != nil {
		return command.Failuref("%v", err)
	}

	newShell := *shell
	if newShell == "" {
		newShell, err = promptShell(stdio)
		if err != nil {
			return command.Failuref("%v", err)
		}
	}
	newShell = strings.TrimSpace(newShell)
	if newShell == "" {
		return command.Failuref("no shell given")
	}

	if err := validateShell(newShell); err != nil {
		return command.Failuref("%v", err)
	}

	if err := changeShell(username, newShell); err != nil {
		return command.Failuref("%v", err)
	}
	return nil
}

// targetUser resolves which user's shell to change: the operand if given,
// otherwise the user running the command.
func targetUser(args []string) (string, error) {
	if len(args) > 1 {
		return "", fmt.Errorf("too many arguments")
	}
	if len(args) == 1 {
		if args[0] == "" {
			return "", fmt.Errorf("empty user name")
		}
		return args[0], nil
	}
	u, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("cannot determine the current user: %v", err)
	}
	return u.Username, nil
}

// promptShell reads a single line (the new shell) from standard input.
func promptShell(stdio command.IO) (string, error) {
	_, _ = fmt.Fprint(stdio.Out, "New shell: ")
	sc := bufio.NewScanner(stdio.In)
	if !sc.Scan() {
		if err := sc.Err(); err != nil {
			return "", err
		}
		return "", fmt.Errorf("no shell given")
	}
	return sc.Text(), nil
}

// validateShell checks that newShell is acceptable: an absolute path with no
// characters that could corrupt the colon-separated, line-oriented passwd
// database. For non-privileged users it must additionally be listed in
// /etc/shells. The superuser may set any such absolute path so that recovery
// and unusual setups remain possible.
func validateShell(newShell string) error {
	if !filepath.IsAbs(newShell) {
		return fmt.Errorf("%q is not an absolute path", newShell)
	}
	// A ':' or a newline would let the value spill into other passwd fields or
	// forge an entire new line (e.g. a passwordless UID-0 account), so reject
	// any field separator or control character outright.
	if strings.ContainsAny(newShell, ":\n\r") {
		return fmt.Errorf("shell path must not contain ':' or a newline")
	}
	for _, r := range newShell {
		if r < 0x20 || r == 0x7f {
			return fmt.Errorf("shell path must not contain control characters")
		}
	}
	if geteuid() == 0 {
		return nil
	}
	shells, err := readShells()
	if err != nil {
		return fmt.Errorf("cannot read %s: %v", shellsPath, err)
	}
	for _, s := range shells {
		if s == newShell {
			return nil
		}
	}
	return fmt.Errorf("%q is not listed in %s", newShell, shellsPath)
}

// readShells returns the shell paths listed in shellsPath, ignoring blank and
// comment lines.
func readShells() ([]string, error) {
	f, err := os.Open(shellsPath) //nolint:gosec // operating on the named shells file is the point
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	var shells []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		shells = append(shells, line)
	}
	return shells, sc.Err()
}

// listShells writes every shell from /etc/shells to standard output.
func listShells(stdio command.IO) error {
	shells, err := readShells()
	if err != nil {
		return command.Failuref("cannot read %s: %v", shellsPath, err)
	}
	for _, s := range shells {
		_, _ = fmt.Fprintln(stdio.Out, s)
	}
	return nil
}

// changeShell rewrites the seventh field of username's /etc/passwd entry to
// newShell, leaving every other line and field untouched.
func changeShell(username, newShell string) error {
	lines, err := readLines(passwdPath)
	if err != nil {
		return fmt.Errorf("cannot read %s: %v", passwdPath, err)
	}

	found := false
	for i, line := range lines {
		fields := strings.Split(line, ":")
		if len(fields) != 7 || fields[0] != username {
			continue
		}
		if fields[6] == newShell {
			return nil // already the requested shell; nothing to write
		}
		fields[6] = newShell
		lines[i] = strings.Join(fields, ":")
		found = true
		break
	}
	if !found {
		return fmt.Errorf("unknown user: %s", username)
	}

	if err := writeLines(passwdPath, lines); err != nil {
		return fmt.Errorf("cannot write %s: %v", passwdPath, err)
	}
	return nil
}

func readLines(path string) ([]string, error) {
	data, err := os.ReadFile(path) //nolint:gosec // well-known passwd path
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
// passwd database truncated or corrupted.
func writeLines(path string, lines []string) error {
	content := strings.Join(lines, "\n") + "\n"
	tmp, err := os.CreateTemp(filepath.Dir(path), ".chsh-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }() // no-op once the rename succeeds

	if _, err := tmp.WriteString(content); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Chmod(0o644); err != nil { //nolint:gosec // passwd is world-readable mode 0644
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}
