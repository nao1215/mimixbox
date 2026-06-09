package mbsh

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/nao1215/mimixbox/internal/applets/shellutils/mbsh/builtin"
	"github.com/nao1215/mimixbox/internal/command"
)

// redirect is one I/O redirection on a simple command.
type redirect struct {
	op   string // "<", ">", or ">>"
	file string
}

// simpleCommand is a single command: temporary NAME=value assignments, the
// command word and its arguments, and any redirections.
type simpleCommand struct {
	assigns []string
	args    []string
	redirs  []redirect
}

// pipeline is one or more simple commands joined by "|".
type pipeline struct {
	cmds []simpleCommand
}

// commandList is one or more pipelines joined by ";". Its exit status is that of
// the last pipeline; a pipeline's exit status is that of its last command.
type commandList struct {
	pipelines []pipeline
}

// parse turns a token stream into a commandList, returning an error for malformed
// input (an operator with no command, a redirection with no target, ...).
func parse(toks []token) (commandList, error) {
	var list commandList
	pl := pipeline{}
	cmd := simpleCommand{}

	finishCmd := func() error {
		if len(cmd.args) == 0 && len(cmd.assigns) == 0 && len(cmd.redirs) == 0 {
			return errors.New("syntax error near unexpected operator")
		}
		pl.cmds = append(pl.cmds, cmd)
		cmd = simpleCommand{}
		return nil
	}
	finishPipeline := func() error {
		if err := finishCmd(); err != nil {
			return err
		}
		list.pipelines = append(list.pipelines, pl)
		pl = pipeline{}
		return nil
	}

	for i := 0; i < len(toks); i++ {
		t := toks[i]
		if t.kind == tokOp {
			switch t.value {
			case ";":
				if err := finishPipeline(); err != nil {
					return commandList{}, err
				}
			case "|":
				if err := finishCmd(); err != nil {
					return commandList{}, err
				}
			case "<", ">", ">>":
				if i+1 >= len(toks) || toks[i+1].kind != tokWord {
					return commandList{}, fmt.Errorf("syntax error: %s needs a filename", t.value)
				}
				i++
				cmd.redirs = append(cmd.redirs, redirect{op: t.value, file: toks[i].value})
			}
			continue
		}
		// A NAME=value word is an assignment only before the command word.
		if t.assignKey != "" && len(cmd.args) == 0 {
			cmd.assigns = append(cmd.assigns, t.assignKey+"="+t.assignVal)
			continue
		}
		cmd.args = append(cmd.args, t.value)
	}

	// Flush the trailing pipeline unless the line ended at a ";".
	if len(cmd.args) > 0 || len(cmd.assigns) > 0 || len(cmd.redirs) > 0 || len(pl.cmds) > 0 {
		if err := finishPipeline(); err != nil {
			return commandList{}, err
		}
	}
	return list, nil
}

// execList runs each pipeline in order and returns the last one's exit status.
func (sh *shell) execList(ctx context.Context, stdio command.IO, list commandList) int {
	status := 0
	for _, pl := range list.pipelines {
		status = sh.execPipeline(ctx, stdio, pl)
		if sh.stop {
			break
		}
	}
	return status
}

// execPipeline runs a pipeline, returning the exit status of its last command.
func (sh *shell) execPipeline(ctx context.Context, stdio command.IO, pl pipeline) int {
	if len(pl.cmds) == 1 {
		return sh.execSimple(ctx, stdio, pl.cmds[0])
	}

	n := len(pl.cmds)
	cmds := make([]*exec.Cmd, n)
	var pipeFiles []*os.File   // pipe ends to close in the parent after Start
	var redirFiles []io.Closer // redirection files to close after Wait
	defer func() {
		for _, f := range redirFiles {
			_ = f.Close()
		}
	}()

	for i, sc := range pl.cmds {
		if len(sc.args) == 0 {
			_, _ = fmt.Fprintln(stdio.Err, "mbsh: missing command in pipeline")
			return 2
		}
		path, argv, ok := resolveCommand(sc.args)
		if !ok {
			_, _ = fmt.Fprintf(stdio.Err, "mbsh: %s: command not found\n", sc.args[0])
			return 127
		}
		c := exec.CommandContext(ctx, path, argv...) //nolint:gosec // user-typed command
		if len(sc.assigns) > 0 {
			c.Env = append(os.Environ(), sc.assigns...)
		}
		c.Stderr = stdio.Err
		if i == 0 {
			c.Stdin = stdio.In
		}
		if i == n-1 {
			c.Stdout = stdio.Out
		}
		cmds[i] = c
	}

	// Wire one pipe between each adjacent pair of commands.
	for i := 0; i < n-1; i++ {
		pr, pw, err := os.Pipe()
		if err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "mbsh: %v\n", err)
			for _, pf := range pipeFiles {
				_ = pf.Close()
			}
			return 1
		}
		cmds[i].Stdout = pw
		cmds[i+1].Stdin = pr
		pipeFiles = append(pipeFiles, pr, pw)
	}

	// Redirections override the pipe/standard wiring for the affected stage.
	for i, sc := range pl.cmds {
		f, err := applyRedirs(cmds[i], sc.redirs)
		if err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "mbsh: %v\n", err)
			for _, pf := range pipeFiles {
				_ = pf.Close()
			}
			return 1
		}
		redirFiles = append(redirFiles, f...)
	}

	for _, c := range cmds {
		if err := c.Start(); err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "mbsh: %v\n", err)
		}
	}
	// Close the parent's copies of the pipe fds so EOF propagates between stages.
	for _, pf := range pipeFiles {
		_ = pf.Close()
	}

	status := 0
	for i, c := range cmds {
		err := c.Wait()
		if i == n-1 {
			status = waitStatus(err)
		}
	}
	return status
}

// execSimple runs a single command with its redirections, dispatching to a
// builtin or an external/applet command.
func (sh *shell) execSimple(ctx context.Context, stdio command.IO, sc simpleCommand) int {
	io := stdio
	var closers []interface{ Close() error }
	defer func() {
		for _, c := range closers {
			_ = c.Close()
		}
	}()
	for _, r := range sc.redirs {
		f, err := openRedir(r)
		if err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "mbsh: %v\n", err)
			return 1
		}
		closers = append(closers, f)
		if r.op == "<" {
			io.In = f
		} else {
			io.Out = f
		}
	}

	if len(sc.args) == 0 {
		// A line of only assignments sets them for the session.
		for _, kv := range sc.assigns {
			if k, v, ok := cut(kv); ok {
				_ = os.Setenv(k, v)
			}
		}
		return 0
	}

	switch sc.args[0] {
	case "exit", "quit":
		sh.stop = true
		return 0
	}

	if builtin.IsBuiltinCmd(sc.args[0]) {
		if err := builtin.Run(io, sc.args[0], sc.args[1:]); err != nil {
			_, _ = fmt.Fprintln(io.Err, err)
			return 1
		}
		return 0
	}

	path, argv, ok := resolveCommand(sc.args)
	if !ok {
		_, _ = fmt.Fprintf(stdio.Err, "mbsh: %s: command not found\n", sc.args[0])
		return 127
	}
	c := exec.CommandContext(ctx, path, argv...) //nolint:gosec // user-typed command
	c.Stdin, c.Stdout, c.Stderr = io.In, io.Out, io.Err
	if len(sc.assigns) > 0 {
		c.Env = append(os.Environ(), sc.assigns...)
	}
	if err := c.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return exitErr.ExitCode()
		}
		_, _ = fmt.Fprintf(stdio.Err, "mbsh: %v\n", err)
		return 127
	}
	return 0
}

// resolveCommand maps a command and its args to the executable path and argv to
// run: the PATH binary when present, otherwise this MimixBox binary re-invoked
// as the applet. ok is false only when the running binary cannot be located.
func resolveCommand(args []string) (path string, argv []string, ok bool) {
	name := args[0]
	if _, err := exec.LookPath(name); err == nil {
		return name, args[1:], true
	}
	if self, err := os.Executable(); err == nil {
		return self, args, true
	}
	return "", nil, false
}

// openRedir opens the file for one redirection.
func openRedir(r redirect) (*os.File, error) {
	switch r.op {
	case "<":
		return os.Open(r.file) //nolint:gosec // user-named file
	case ">":
		return os.Create(r.file) //nolint:gosec // user-named file
	case ">>":
		return os.OpenFile(r.file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644) //nolint:gosec // user-named file
	default:
		return nil, fmt.Errorf("unknown redirection %q", r.op)
	}
}

// applyRedirs opens and applies a stage's redirections to an exec.Cmd, returning
// the opened files for the caller to close after the command finishes.
func applyRedirs(c *exec.Cmd, redirs []redirect) ([]io.Closer, error) {
	var files []io.Closer
	for _, r := range redirs {
		f, err := openRedir(r)
		if err != nil {
			for _, of := range files {
				_ = of.Close()
			}
			return nil, err
		}
		files = append(files, f)
		if r.op == "<" {
			c.Stdin = f
		} else {
			c.Stdout = f
		}
	}
	return files, nil
}

// waitStatus extracts a process exit status from an exec Wait error.
func waitStatus(err error) int {
	if err == nil {
		return 0
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	return 127
}

// cut splits "k=v" into k and v.
func cut(s string) (string, string, bool) {
	for i := 0; i < len(s); i++ {
		if s[i] == '=' {
			return s[:i], s[i+1:], true
		}
	}
	return s, "", false
}
