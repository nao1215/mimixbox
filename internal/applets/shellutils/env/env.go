// Package env implements the env applet: run a program in a modified
// environment, or print the current environment.
package env

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the env applet.
type Command struct{}

// New returns an env command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "env" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string {
	return "Run a program in a modified environment / print the environment"
}

// Run executes env.
func (c *Command) Run(ctx context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... [NAME=VALUE]... [COMMAND [ARG]...]", stdio.Err)
	// Options stop at the first operand so that flags meant for COMMAND (such as
	// "sh -c") are passed through untouched instead of parsed by env.
	fs.SetInterspersed(false)
	ignore := fs.BoolP("ignore-environment", "i", false, "start with an empty environment")
	unset := fs.StringArrayP("unset", "u", nil, "remove variable from the environment")
	null := fs.BoolP("null", "0", false, "end each output line with NUL, not newline")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	// Build the base environment: empty for -i, otherwise the inherited one.
	var environ []string
	if !*ignore {
		environ = os.Environ()
	}

	// Leading NAME=VALUE operands set variables; the first operand that is not
	// an assignment begins the command and its arguments.
	rest := fs.Args()
	var idx int
	for idx = 0; idx < len(rest); idx++ {
		name, _, ok := splitAssignment(rest[idx])
		if !ok {
			break
		}
		environ = setEnv(environ, name, rest[idx])
	}

	// Apply -u removals after the assignments so an explicit assignment can be
	// unset and a later assignment is honored.
	for _, name := range *unset {
		environ = unsetEnv(environ, name)
	}

	argv := rest[idx:]
	if len(argv) == 0 {
		printEnviron(stdio, environ, *null)
		return nil
	}

	return runCommand(ctx, stdio, environ, argv)
}

// splitAssignment reports whether operand has the form NAME=VALUE and, if so,
// returns its NAME and VALUE. An operand with an empty name (e.g. "=x") is not
// treated as an assignment.
func splitAssignment(operand string) (name, value string, ok bool) {
	i := strings.IndexByte(operand, '=')
	if i <= 0 {
		return "", "", false
	}
	return operand[:i], operand[i+1:], true
}

// setEnv replaces the entry whose name matches assignment ("NAME=VALUE") or
// appends it when the name is not present.
func setEnv(environ []string, name, assignment string) []string {
	prefix := name + "="
	for i, e := range environ {
		if strings.HasPrefix(e, prefix) {
			environ[i] = assignment
			return environ
		}
	}
	return append(environ, assignment)
}

// unsetEnv removes every entry whose name matches.
func unsetEnv(environ []string, name string) []string {
	prefix := name + "="
	out := environ[:0]
	for _, e := range environ {
		if strings.HasPrefix(e, prefix) {
			continue
		}
		out = append(out, e)
	}
	return out
}

// printEnviron writes each NAME=VALUE entry, terminated by newline (or NUL when
// null is set).
func printEnviron(stdio command.IO, environ []string, null bool) {
	end := byte('\n')
	if null {
		end = 0
	}
	for _, e := range environ {
		_, _ = fmt.Fprintf(stdio.Out, "%s%c", e, end)
	}
}

// runCommand execs argv[0] with the remaining args and the modified
// environment, mirroring its exit status. A command that cannot be found is
// reported GNU-style and maps to exit status 127.
func runCommand(ctx context.Context, stdio command.IO, environ, argv []string) error {
	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...) //nolint:gosec // running a user-named command is the whole point
	cmd.Env = environ
	cmd.Stdin = stdio.In
	cmd.Stdout = stdio.Out
	cmd.Stderr = stdio.Err

	err := cmd.Run()
	if err == nil {
		return nil
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return &command.ExitError{Code: exitErr.ExitCode()}
	}

	// Anything else (most commonly "executable file not found") means the
	// command could not be started.
	_, _ = fmt.Fprintf(stdio.Err, "env: '%s': No such file or directory\n", argv[0])
	return &command.ExitError{Code: 127}
}
