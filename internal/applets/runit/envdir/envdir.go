// Package envdir implements the envdir applet: run a program with environment
// variables set from the files of a directory.
package envdir

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the envdir applet.
type Command struct{}

// New returns an envdir command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "envdir" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Run a program with env from a directory" }

// Run executes envdir.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "DIR PROG [ARG...]", stdio.Err).WithHelp(command.Help{
		Description: "Run PROG with the environment modified by the files in DIR: each file sets the " +
			"variable named after it to its first line (with trailing whitespace removed), and an " +
			"empty file removes that variable. This is the daemontools/runit envdir.",
		Examples: []command.Example{
			{Command: "envdir ./env printenv FOO", Explain: "Run printenv with FOO from ./env/FOO."},
		},
		ExitStatus: "PROG's exit status, or 1 on a usage or directory error.",
	})
	// Stop at the first operand so PROG's own flags are not parsed by envdir.
	fs.SetInterspersed(false)

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	rest := fs.Args()
	if len(rest) < 2 {
		return command.Failuref("a directory and a program are required")
	}
	dir, prog, progArgs := rest[0], rest[1], rest[2:]

	env, err := applyDir(os.Environ(), dir)
	if err != nil {
		return command.Failuref("%s: %v", dir, err)
	}

	cmd := exec.Command(prog, progArgs...) //nolint:gosec // running the user's program is the point
	cmd.Env = env
	cmd.Stdin, cmd.Stdout, cmd.Stderr = stdio.In, stdio.Out, stdio.Err
	if err := cmd.Run(); err != nil {
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

	vars := toMap(env)
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		data, err := os.ReadFile(dir + "/" + e.Name()) //nolint:gosec // env directory file
		if err != nil {
			return nil, err
		}
		value := strings.TrimRight(firstLine(string(data)), " \t")
		if value == "" && len(data) == 0 {
			delete(vars, e.Name()) // an empty file removes the variable
			continue
		}
		vars[e.Name()] = value
	}
	return fromMap(vars), nil
}

// firstLine returns the content up to the first newline.
func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}

func toMap(env []string) map[string]string {
	m := make(map[string]string, len(env))
	for _, kv := range env {
		if k, v, ok := strings.Cut(kv, "="); ok {
			m[k] = v
		}
	}
	return m
}

func fromMap(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k, v := range m {
		out = append(out, k+"="+v)
	}
	return out
}
