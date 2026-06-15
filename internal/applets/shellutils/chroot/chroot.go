// Package chroot implements the chroot applet: run a command or an interactive
// shell with a special root directory.
package chroot

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the chroot applet.
type Command struct{}

// New returns a chroot command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "chroot" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string {
	return "Run command or interactive shell with special root directory"
}

// indirected so tests can drive the failure-mode logic without privileges.
var (
	setgid    = syscall.Setgid
	setuid    = syscall.Setuid
	setgroups = syscall.Setgroups
)

// identity is the resolved UID/GID/supplementary-group set to apply inside the
// jail. It is produced by resolveIdentity from the jail's account databases and
// applied by apply after syscall.Chroot has succeeded.
type identity struct {
	uid    int
	gid    int
	groups []int
}

// Run executes chroot. It changes the root directory to NEWROOT and runs
// COMMAND (defaulting to the shell) inside it. This requires root privileges;
// when it cannot change the root directory it prints a GNU-style error and
// returns command.SilentFailure().
//
// Identity model: by default chroot keeps the caller's UID/GID, matching GNU
// and BusyBox. When --userspec USER[:GROUP] (and optionally --groups
// G[,G]...) is given, the names/IDs are resolved against the JAIL's
// /etc/passwd and /etc/group (read after entering the jail, never the host's)
// and privileges are dropped in the order setgroups -> setgid -> setuid. If a
// requested name cannot be resolved in the jail, or a privilege drop fails,
// chroot reports a deterministic error and exits non-zero rather than running
// the command with mismatched host identity.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION] NEWROOT [COMMAND [ARG]...]", stdio.Err).WithHelp(command.Help{
		Description: "Run COMMAND (or $SHELL) with the root directory set to NEWROOT. Requires root " +
			"privileges. With --userspec, the user/group is resolved against the JAIL's " +
			"/etc/passwd and /etc/group and privileges are dropped after entering the jail.",
		Examples: []command.Example{
			{Command: "chroot /jail /bin/sh", Explain: "Run a shell inside /jail as the current user."},
			{Command: "chroot --userspec=1000:1000 /jail /bin/sh", Explain: "Drop to uid/gid 1000 in the jail."},
			{Command: "chroot --userspec=alice:devs /jail id", Explain: "Resolve alice/devs against the jail's databases."},
		},
		ExitStatus: "0    COMMAND ran successfully.\n" +
			"1    chroot failed, an identity could not be resolved, or COMMAND failed.\n",
		Notes: []string{
			"USER and GROUP in --userspec may be names or numeric IDs; names are resolved against " +
				"the jail's /etc/passwd and /etc/group, not the host's.",
			"When GROUP is omitted from --userspec and USER is a name, the user's primary group from " +
				"the jail's /etc/passwd is used; if USER is numeric, the gid defaults to the uid.",
			"--groups sets the supplementary group list; with --userspec but no --groups, supplementary " +
				"groups are cleared (matching coreutils).",
			"Without --userspec the caller's UID/GID are kept unchanged.",
		},
	})

	userSpec := fs.String("userspec", "", "USER[:GROUP] to switch to inside the jail (names or IDs)")
	groupsCSV := fs.String("groups", "", "comma-separated supplementary GROUP list for inside the jail")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	operands := fs.Args()
	if len(operands) == 0 {
		_, _ = fmt.Fprintln(stdio.Err, "chroot: missing operand")
		return command.SilentFailure()
	}

	newRoot := os.ExpandEnv(operands[0])
	if err := syscall.Chroot(newRoot); err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "chroot: cannot change root directory to '%s': %s\n",
			newRoot, reason(err))
		return command.SilentFailure()
	}

	//----------------From here, in the prison-------------------
	if err := os.Chdir("/"); err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "chroot: cannot change root directory to '%s': %s\n",
			newRoot, reason(err))
		return command.SilentFailure()
	}

	// Resolve and apply the requested identity against the JAIL's account
	// databases. /etc/passwd and /etc/group are now the jail's files because
	// the root directory has already changed.
	if *userSpec != "" || *groupsCSV != "" {
		passwd, _ := os.Open("/etc/passwd")
		group, _ := os.Open("/etc/group")
		id, ierr := resolveIdentity(*userSpec, *groupsCSV, passwd, group)
		if passwd != nil {
			_ = passwd.Close()
		}
		if group != nil {
			_ = group.Close()
		}
		if ierr != nil {
			_, _ = fmt.Fprintf(stdio.Err, "chroot: %s\n", ierr)
			return command.SilentFailure()
		}
		if aerr := id.apply(); aerr != nil {
			_, _ = fmt.Fprintf(stdio.Err, "chroot: %s\n", aerr)
			return command.SilentFailure()
		}
	}

	name, argv := decideExecCommand(operands[1:])
	// Reset the environment variable SHELL for the jail environment.
	if err := os.Setenv("SHELL", name); err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "chroot: %v\n", err)
		return command.SilentFailure()
	}

	cmd := exec.Command(name, argv...) //nolint:gosec // running a user-named command is the whole point
	cmd.Stdin = stdio.In
	cmd.Stdout = stdio.Out
	cmd.Stderr = stdio.Err
	cmd.Env = os.Environ()

	if err := cmd.Run(); err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "chroot: %v\n", err)
		return command.SilentFailure()
	}
	return nil
}

// apply drops privileges to the resolved identity. The order matters:
// supplementary groups and the primary GID must be set while still privileged,
// before the UID is dropped (after setuid the process can no longer change its
// groups or gid).
func (id identity) apply() error {
	if id.groups != nil {
		if err := setgroups(id.groups); err != nil {
			return fmt.Errorf("cannot set supplementary groups: %w", err)
		}
	}
	if err := setgid(id.gid); err != nil {
		return fmt.Errorf("cannot set gid to %d: %w", id.gid, err)
	}
	if err := setuid(id.uid); err != nil {
		return fmt.Errorf("cannot set uid to %d: %w", id.uid, err)
	}
	return nil
}

// resolveIdentity turns a --userspec value and a --groups CSV into a concrete
// uid/gid/supplementary-group set, resolving names against the passwd and group
// sources (the jail's /etc/passwd and /etc/group). It is pure with respect to
// the process: it performs no syscalls, so it is unit-testable without root.
//
// passwd or group may be nil (e.g. the file is absent in the jail); in that
// case only numeric IDs can be resolved and a name lookup returns an error.
func resolveIdentity(userSpec, groupsCSV string, passwd, group io.Reader) (identity, error) {
	if userSpec == "" && groupsCSV != "" {
		// --groups without --userspec has no deterministic identity to keep, so
		// reject it rather than guessing the caller's uid/gid.
		return identity{}, fmt.Errorf("--groups requires --userspec")
	}

	users := parsePasswd(passwd)
	groups := parseGroup(group)

	var id identity

	if userSpec != "" {
		userPart, groupPart, hasGroup := strings.Cut(userSpec, ":")
		if userPart == "" {
			return identity{}, fmt.Errorf("invalid userspec: %q", userSpec)
		}

		uid, primaryGID, uerr := lookupUser(userPart, users)
		if uerr != nil {
			return identity{}, uerr
		}
		id.uid = uid

		switch {
		case hasGroup && groupPart != "":
			gid, gerr := lookupGroup(groupPart, groups)
			if gerr != nil {
				return identity{}, gerr
			}
			id.gid = gid
		case hasGroup && groupPart == "":
			return identity{}, fmt.Errorf("invalid userspec: %q", userSpec)
		case primaryGID >= 0:
			id.gid = primaryGID
		default:
			// Numeric user with no passwd entry and no group given: gid
			// defaults to the uid, matching coreutils.
			id.gid = uid
		}
	}

	if groupsCSV != "" {
		list, gerr := lookupGroups(groupsCSV, groups)
		if gerr != nil {
			return identity{}, gerr
		}
		id.groups = list
	} else if userSpec != "" {
		// coreutils clears supplementary groups when --userspec is given
		// without --groups. Use an empty (non-nil) slice so apply calls
		// setgroups with no groups rather than skipping it.
		id.groups = []int{}
	}

	return id, nil
}

// lookupUser resolves name (a login name or numeric uid) to a uid and the
// user's primary gid. primaryGID is -1 when it cannot be determined (numeric
// uid with no matching passwd entry).
func lookupUser(name string, users map[string]passwdEntry) (uid, primaryGID int, err error) {
	if e, ok := users[name]; ok {
		return e.uid, e.gid, nil
	}
	if n, perr := strconv.Atoi(name); perr == nil && n >= 0 {
		return n, -1, nil
	}
	return 0, -1, fmt.Errorf("invalid user: %q", name)
}

// lookupGroup resolves name (a group name or numeric gid) to a gid.
func lookupGroup(name string, groups map[string]int) (int, error) {
	if gid, ok := groups[name]; ok {
		return gid, nil
	}
	if n, perr := strconv.Atoi(name); perr == nil && n >= 0 {
		return n, nil
	}
	return 0, fmt.Errorf("invalid group: %q", name)
}

// lookupGroups resolves a comma-separated list of group names/IDs.
func lookupGroups(csv string, groups map[string]int) ([]int, error) {
	parts := strings.Split(csv, ",")
	out := make([]int, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		gid, err := lookupGroup(p, groups)
		if err != nil {
			return nil, err
		}
		out = append(out, gid)
	}
	return out, nil
}

// passwdEntry holds the fields of /etc/passwd needed for identity resolution.
type passwdEntry struct {
	uid int
	gid int
}

// parsePasswd reads /etc/passwd-formatted data into a name->entry map. Numeric
// uid/gid fields that do not parse cause the line to be skipped. A nil reader
// yields an empty map.
func parsePasswd(r io.Reader) map[string]passwdEntry {
	out := map[string]passwdEntry{}
	if r == nil {
		return out
	}
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		f := strings.Split(line, ":")
		if len(f) < 4 {
			continue
		}
		uid, uerr := strconv.Atoi(f[2])
		gid, gerr := strconv.Atoi(f[3])
		if uerr != nil || gerr != nil {
			continue
		}
		out[f[0]] = passwdEntry{uid: uid, gid: gid}
	}
	return out
}

// parseGroup reads /etc/group-formatted data into a name->gid map. A nil reader
// yields an empty map.
func parseGroup(r io.Reader) map[string]int {
	out := map[string]int{}
	if r == nil {
		return out
	}
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		f := strings.Split(line, ":")
		if len(f) < 3 {
			continue
		}
		gid, err := strconv.Atoi(f[2])
		if err != nil {
			continue
		}
		out[f[0]] = gid
	}
	return out
}

// decideExecCommand resolves the command to run inside the jail. extra are the
// operands after NEWROOT. When none are given, the command is the shell taken
// from $SHELL (falling back to /bin/sh) run interactively.
func decideExecCommand(extra []string) (name string, argv []string) {
	if len(extra) == 0 {
		shell := os.Getenv("SHELL")
		if shell == "" {
			shell = "/bin/sh"
		}
		return shell, []string{"-i"}
	}
	return extra[0], extra[1:]
}

// reason maps a chroot/chdir failure to the GNU-style trailing message.
func reason(err error) string {
	if os.IsNotExist(err) {
		return "No such file or directory"
	}
	if os.IsPermission(err) {
		return "Operation not permitted"
	}
	return err.Error()
}
