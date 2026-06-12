// Package chpst implements the chpst applet: run a program with a changed
// process state (uid/gid, environment directory, resource limits, niceness).
package chpst

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"

	"github.com/nao1215/mimixbox/internal/command"
	"golang.org/x/sys/unix"
)

// Command is the chpst applet.
type Command struct{}

// New returns a chpst command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "chpst" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Run a program with a changed process state" }

// procSpec captures the requested process changes.
type procSpec struct {
	env      []string
	uid, gid int
	setCreds bool
	nice     int
	setNice  bool
	prog     string
	args     []string
}

// Injected so the database, the rlimits, and the privileged exec are testable.
var (
	passwdPath  = "/etc/passwd"
	setRlimitFn = func(resource int, value uint64) error {
		return unix.Setrlimit(resource, &unix.Rlimit{Cur: value, Max: value})
	}
	runFn = func(ctx context.Context, stdio command.IO, spec procSpec) error {
		if spec.setNice {
			_ = unix.Setpriority(unix.PRIO_PROCESS, 0, spec.nice)
		}
		cmd := exec.CommandContext(ctx, spec.prog, spec.args...) //nolint:gosec // running the user's program is the point
		cmd.Env = spec.env
		cmd.Stdin, cmd.Stdout, cmd.Stderr = stdio.In, stdio.Out, stdio.Err
		if spec.setCreds {
			cmd.SysProcAttr = &syscall.SysProcAttr{
				Credential: &syscall.Credential{Uid: uint32(spec.uid), Gid: uint32(spec.gid)},
			}
		}
		return cmd.Run()
	}
)

var limitFlags = []struct {
	short    string
	resource int
}{
	{"m", unix.RLIMIT_AS}, {"d", unix.RLIMIT_DATA}, {"o", unix.RLIMIT_NOFILE},
	{"p", unix.RLIMIT_NPROC}, {"f", unix.RLIMIT_FSIZE}, {"c", unix.RLIMIT_CORE},
	{"t", unix.RLIMIT_CPU},
}

// Run executes chpst.
func (c *Command) Run(ctx context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-u USER] [-e DIR] [-n NICE] [limits] PROG [ARG...]", stdio.Err).WithHelp(command.Help{
		Description: "Change the process state, then run PROG: -u USER sets the uid and gid (dropping " +
			"privilege); -e DIR loads environment variables from a directory (as envdir); -n adds to " +
			"the niceness; and -m/-d/-o/-p/-f/-c/-t set resource limits (as softlimit). This is the " +
			"daemontools/runit chpst.",
		Examples: []command.Example{
			{Command: "chpst -u nobody -e ./env -o 64 mydaemon", Explain: "Drop to nobody, load env, cap files."},
		},
		ExitStatus: "PROG's exit status, or 1 on a usage error.",
	})
	userFlag := fs.StringP("user", "u", "", "set the uid/gid to this user's")
	envDir := fs.StringP("envdir", "e", "", "load environment variables from this directory")
	nice := fs.IntP("nice", "n", 0, "add this value to the niceness")
	limits := map[int]*int64{}
	for _, lf := range limitFlags {
		limits[lf.resource] = fs.Int64P("limit-"+lf.short, lf.short, -1, "resource limit")
	}
	fs.SetInterspersed(false)

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	rest := fs.Args()
	if len(rest) == 0 {
		return command.Failuref("a program is required")
	}

	spec := procSpec{env: os.Environ(), prog: rest[0], args: rest[1:]}

	if *envDir != "" {
		if spec.env, err = applyDir(spec.env, *envDir); err != nil {
			return command.Failuref("%s: %v", *envDir, err)
		}
	}
	if *userFlag != "" {
		if spec.uid, spec.gid, err = resolve(*userFlag); err != nil {
			return command.Failuref("%v", err)
		}
		spec.setCreds = true
	}
	if fs.Changed("nice") {
		spec.nice = *nice
		spec.setNice = true
	}
	for _, lf := range limitFlags {
		if v := *limits[lf.resource]; v >= 0 {
			if err := setRlimitFn(lf.resource, uint64(v)); err != nil {
				return command.Failuref("cannot set -%s limit: %v", lf.short, err)
			}
		}
	}

	if err := runFn(ctx, stdio, spec); err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			return &command.ExitError{Code: ee.ExitCode()}
		}
		return command.Failuref("%v", err)
	}
	return nil
}

// applyDir returns env with the variables from dir's files applied.
func applyDir(env []string, dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	vars := map[string]string{}
	for _, kv := range env {
		if k, v, ok := strings.Cut(kv, "="); ok {
			vars[k] = v
		}
	}
	for _, e := range entries {
		if e.IsDir() || strings.ContainsRune(e.Name(), '=') {
			continue
		}
		data, err := os.ReadFile(dir + "/" + e.Name()) //nolint:gosec // env directory file
		if err != nil {
			return nil, err
		}
		line := string(data)
		if i := strings.IndexByte(line, '\n'); i >= 0 {
			line = line[:i]
		}
		value := strings.TrimRight(line, " \t")
		if value == "" && len(data) == 0 {
			delete(vars, e.Name())
			continue
		}
		vars[e.Name()] = value
	}
	out := make([]string, 0, len(vars))
	for k, v := range vars {
		out = append(out, k+"="+v)
	}
	return out, nil
}

// resolve returns the uid and gid of the named user (USER or USER:GROUP form,
// where GROUP if present overrides the gid).
func resolve(spec string) (uid, gid int, err error) {
	name, group, _ := strings.Cut(spec, ":")
	data, err := os.ReadFile(passwdPath) //nolint:gosec // well-known passwd path
	if err != nil {
		return 0, 0, err
	}
	for _, line := range strings.Split(strings.TrimRight(string(data), "\n"), "\n") {
		f := strings.Split(line, ":")
		if len(f) < 4 || f[0] != name {
			continue
		}
		u, err1 := strconv.Atoi(f[2])
		g, err2 := strconv.Atoi(f[3])
		if err1 != nil || err2 != nil {
			return 0, 0, errors.New("user " + name + " has an invalid uid/gid")
		}
		if group != "" {
			if g, err = strconv.Atoi(group); err != nil {
				return 0, 0, errors.New("invalid group id: " + group)
			}
		}
		return u, g, nil
	}
	return 0, 0, errors.New("unknown user: " + name)
}
