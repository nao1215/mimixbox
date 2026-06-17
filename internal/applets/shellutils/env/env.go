// Package env implements the env applet: run a program in a modified
// environment, or print the current environment.
package env

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"

	"github.com/nao1215/mimixbox/internal/command"
	"github.com/nao1215/mimixbox/internal/signal"
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
	fs := command.NewFlagSet(c.Name(), "[OPTION]... [NAME=VALUE]... [COMMAND [ARG]...]", stdio.Err).WithHelp(command.Help{
		Description: "Set each NAME=VALUE in the environment and run COMMAND with the resulting environment. " +
			"With no COMMAND, print the resulting environment instead.",
		Examples: []command.Example{
			{Command: "env", Explain: "Print the current environment, one variable per line."},
			{Command: "env FOO=bar printenv FOO", Explain: "Run printenv with FOO set to bar."},
			{Command: "env -i sh -c 'echo $PATH'", Explain: "Run a command with an empty environment."},
			{Command: "env --chdir=/tmp pwd", Explain: "Change to /tmp before running pwd."},
			{Command: "env -S 'printf %s\\n hi' ", Explain: "Split the single string into arguments before running it."},
		},
		ExitStatus: "0    success.\n127  COMMAND could not be started (e.g. not found).\nN    otherwise, the exit status of COMMAND.",
	})
	// Options stop at the first operand so that flags meant for COMMAND (such as
	// "sh -c") are passed through untouched instead of parsed by env.
	fs.SetInterspersed(false)
	ignore := fs.BoolP("ignore-environment", "i", false, "start with an empty environment")
	unset := fs.StringArrayP("unset", "u", nil, "remove variable from the environment")
	null := fs.BoolP("null", "0", false, "end each output line with NUL, not newline")
	chdir := fs.StringP("chdir", "C", "", "change working directory to DIR before running COMMAND")
	splitString := fs.StringP("split-string", "S", "", "process and split S into separate arguments; used to pass multiple arguments on shebang lines")
	ignoreSignal := fs.StringP("ignore-signal", "", "", "set handling of SIGLIST signals (comma/space separated) to do nothing in the child; with no list, ignore all catchable signals")
	fs.Lookup("ignore-signal").NoOptDefVal = allSignalsSentinel

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	// Parse and validate the ignore-signal list up front so a bad name fails
	// before any command is started.
	var ignoreSignals []os.Signal
	if fs.Changed("ignore-signal") {
		ignoreSignals, err = parseIgnoreSignals(*ignoreSignal)
		if err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "env: %v\n", err)
			return &command.ExitError{Code: 125}
		}
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

	// --split-string expands its single argument into multiple arguments that
	// are prepended before any command operands. This is what lets a shebang
	// line carry several arguments through env.
	if *splitString != "" {
		split, splitErr := splitArgs(*splitString)
		if splitErr != nil {
			_, _ = fmt.Fprintf(stdio.Err, "env: %v\n", splitErr)
			return &command.ExitError{Code: 125}
		}
		argv = append(split, argv...)
	}

	if len(argv) == 0 {
		printEnviron(stdio, environ, *null)
		return nil
	}

	return runCommand(ctx, stdio, environ, argv, *chdir, ignoreSignals)
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
// reported GNU-style and maps to exit status 127. When chdir is non-empty the
// command is run with that working directory (an unusable directory is a
// fatal error). The signals in ignoreSignals are set to SIG_IGN so the
// launched child inherits that disposition.
func runCommand(ctx context.Context, stdio command.IO, environ, argv []string, chdir string, ignoreSignals []os.Signal) error {
	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...) //nolint:gosec // running a user-named command is the whole point
	cmd.Env = environ
	cmd.Stdin = stdio.In
	cmd.Stdout = stdio.Out
	cmd.Stderr = stdio.Err

	if chdir != "" {
		info, statErr := os.Stat(chdir)
		if statErr != nil {
			_, _ = fmt.Fprintf(stdio.Err, "env: cannot change directory to '%s': %v\n", chdir, statErr)
			return &command.ExitError{Code: 125}
		}
		if !info.IsDir() {
			_, _ = fmt.Fprintf(stdio.Err, "env: cannot change directory to '%s': Not a directory\n", chdir)
			return &command.ExitError{Code: 125}
		}
		cmd.Dir = chdir
	}

	// Ignore the requested signals for the duration of the child so the child,
	// which inherits the disposition, runs with those signals set to SIG_IGN.
	// The prior disposition is restored afterwards so env's own process state
	// is left unchanged (keeps repeated in-process runs deterministic and
	// testable).
	return signal.IgnoreDuring(ignoreSignals, func() error {
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
	})
}

// allSignalsSentinel is the value the ignore-signal flag takes when it is given
// with no argument (--ignore-signal), meaning "ignore all catchable signals".
const allSignalsSentinel = "\x00ALL\x00"

// splitArgs splits s the way GNU env's --split-string does: tokens are
// separated by unescaped whitespace, and a backslash introduces an escape.
// The supported escapes are \t (tab), \n (newline), \r (carriage return),
// \f (form feed), \v (vertical tab), \\ (backslash), \# (literal #), and the
// special \_ which inserts a space without ending the current argument. A
// leading '#' begins a comment that runs to end of input.
func splitArgs(s string) ([]string, error) {
	var args []string
	var cur strings.Builder
	inWord := false
	flush := func() {
		if inWord {
			args = append(args, cur.String())
			cur.Reset()
			inWord = false
		}
	}

	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		c := runes[i]
		switch {
		case c == '#' && !inWord:
			// A '#' at the start of a token begins a comment to end of input.
			return args, nil
		case c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '\f' || c == '\v':
			flush()
		case c == '\\':
			if i+1 >= len(runes) {
				return nil, fmt.Errorf("no terminating quote in -S string")
			}
			i++
			inWord = true
			switch runes[i] {
			case 't':
				cur.WriteByte('\t')
			case 'n':
				cur.WriteByte('\n')
			case 'r':
				cur.WriteByte('\r')
			case 'f':
				cur.WriteByte('\f')
			case 'v':
				cur.WriteByte('\v')
			case '\\':
				cur.WriteByte('\\')
			case '#':
				cur.WriteByte('#')
			case '_':
				cur.WriteByte(' ')
			case 'c':
				// \c ends argument processing entirely (GNU behavior).
				flush()
				return args, nil
			default:
				return nil, fmt.Errorf("invalid sequence '\\%c' in -S", runes[i])
			}
		default:
			inWord = true
			cur.WriteRune(c)
		}
	}
	flush()
	return args, nil
}

// signalNames maps signal names (without the SIG prefix) to their values.
var signalNames = map[string]syscall.Signal{
	"HUP": syscall.SIGHUP, "INT": syscall.SIGINT, "QUIT": syscall.SIGQUIT,
	"ILL": syscall.SIGILL, "TRAP": syscall.SIGTRAP, "ABRT": syscall.SIGABRT,
	"BUS": syscall.SIGBUS, "FPE": syscall.SIGFPE, "KILL": syscall.SIGKILL,
	"USR1": syscall.SIGUSR1, "SEGV": syscall.SIGSEGV, "USR2": syscall.SIGUSR2,
	"PIPE": syscall.SIGPIPE, "ALRM": syscall.SIGALRM, "TERM": syscall.SIGTERM,
	"CHLD": syscall.SIGCHLD, "CONT": syscall.SIGCONT, "STOP": syscall.SIGSTOP,
	"TSTP": syscall.SIGTSTP, "TTIN": syscall.SIGTTIN, "TTOU": syscall.SIGTTOU,
	"URG": syscall.SIGURG, "XCPU": syscall.SIGXCPU, "XFSZ": syscall.SIGXFSZ,
	"VTALRM": syscall.SIGVTALRM, "PROF": syscall.SIGPROF, "WINCH": syscall.SIGWINCH,
	"IO": syscall.SIGIO, "SYS": syscall.SIGSYS,
}

// catchableSignals lists the signals env ignores when --ignore-signal is given
// with no list. SIGKILL and SIGSTOP cannot be caught and are deliberately
// omitted.
var catchableSignals = []os.Signal{
	syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM,
	syscall.SIGUSR1, syscall.SIGUSR2, syscall.SIGPIPE, syscall.SIGALRM,
	syscall.SIGTSTP, syscall.SIGTTIN, syscall.SIGTTOU,
}

// parseSignal resolves a single signal token (a name with or without the SIG
// prefix, or a positive number) to its syscall.Signal.
func parseSignal(token string) (syscall.Signal, error) {
	t := strings.ToUpper(strings.TrimSpace(token))
	if t == "" {
		return 0, fmt.Errorf("invalid signal: empty name")
	}
	name := strings.TrimPrefix(t, "SIG")
	if sig, ok := signalNames[name]; ok {
		return sig, nil
	}
	if n, err := strconv.Atoi(t); err == nil && n > 0 {
		return syscall.Signal(n), nil
	}
	return 0, fmt.Errorf("%s: invalid signal", token)
}

// parseIgnoreSignals turns the value of --ignore-signal into the list of
// signals to ignore. The sentinel means "all catchable signals"; otherwise the
// value is a comma- and/or whitespace-separated list of signal names/numbers.
// A bad name is an error.
func parseIgnoreSignals(value string) ([]os.Signal, error) {
	if value == allSignalsSentinel || value == "" {
		out := make([]os.Signal, len(catchableSignals))
		copy(out, catchableSignals)
		return out, nil
	}
	fields := strings.FieldsFunc(value, func(r rune) bool {
		return r == ',' || r == ' ' || r == '\t'
	})
	var out []os.Signal
	for _, f := range fields {
		sig, err := parseSignal(f)
		if err != nil {
			return nil, err
		}
		out = append(out, sig)
	}
	return out, nil
}
