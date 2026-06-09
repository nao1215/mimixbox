// Package getopt implements the getopt applet: parse command-line options for
// shell scripts the way util-linux getopt(1) does in its enhanced (GNU) mode.
// It normalizes a script's arguments into a quoted, permuted list that the
// script then re-reads with `set -- "$@"` / eval.
package getopt

import (
	"context"
	"fmt"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the getopt applet.
type Command struct{}

// New returns a getopt command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "getopt" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Parse command options (enhanced, like util-linux getopt)" }

// argSpec records whether an option takes an argument.
type argSpec int

const (
	noArg argSpec = iota
	requiredArg
	optionalArg
)

// Run executes getopt.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	if len(args) == 1 && (args[0] == "--help" || args[0] == "--version") {
		fs := command.NewFlagSet(c.Name(), "-o OPTSTRING [--long LONGOPTS] [-n NAME] -- ARGS...", stdio.Err).WithHelp(c.help())
		_, _ = fs.Parse(stdio, args)
		return nil
	}

	optstring, longopts, name, quiet, legacy, params, perr := splitArgs(args)
	if perr != nil {
		_, _ = fmt.Fprintf(stdio.Err, "getopt: %v\n", perr)
		return &command.ExitError{Code: 2}
	}
	shorts := parseOptstring(optstring)
	longs := parseLongopts(longopts)

	out, errs := parse(params, shorts, longs, !legacy)
	if len(errs) > 0 {
		if !quiet {
			for _, e := range errs {
				_, _ = fmt.Fprintf(stdio.Err, "%s: %s\n", name, e)
			}
		}
		// getopt still prints the best-effort normalized line, but exits 1.
		_, _ = fmt.Fprintln(stdio.Out, out)
		return &command.ExitError{Code: 1}
	}
	_, _ = fmt.Fprintln(stdio.Out, out)
	return nil
}

// splitArgs separates getopt's own options from the ARGS to be parsed. It
// supports the enhanced form (-o/-l/-n/-q ... -- ARGS) and the legacy form
// (OPTSTRING ARGS).
func splitArgs(args []string) (optstring, longopts, name string, quiet, legacy bool, params []string, err error) {
	name = "getopt"
	if len(args) == 0 {
		return "", "", name, false, false, nil, fmt.Errorf("missing optstring argument")
	}
	// Legacy form: first argument is the optstring (does not start with '-').
	if !strings.HasPrefix(args[0], "-") {
		return args[0], "", name, false, true, args[1:], nil
	}

	i := 0
	for i < len(args) {
		a := args[i]
		switch {
		case a == "--":
			i++
			return optstring, longopts, name, quiet, false, args[i:], nil
		case a == "-o" || a == "--options":
			i++
			if i >= len(args) {
				return "", "", name, false, false, nil, fmt.Errorf("option '%s' requires an argument", a)
			}
			optstring = args[i]
		case strings.HasPrefix(a, "-o"):
			optstring = a[2:]
		case a == "-l" || a == "--longoptions" || a == "--long":
			i++
			if i >= len(args) {
				return "", "", name, false, false, nil, fmt.Errorf("option '%s' requires an argument", a)
			}
			longopts = args[i]
		case a == "-n" || a == "--name":
			i++
			if i >= len(args) {
				return "", "", name, false, false, nil, fmt.Errorf("option '%s' requires an argument", a)
			}
			name = args[i]
		case a == "-q" || a == "--quiet":
			quiet = true
		case a == "-a" || a == "--alternative" || a == "-u" || a == "--unquoted" || a == "-Q" || a == "--quiet-output" || a == "-T" || a == "--test":
			// Accepted for compatibility; behavior is the enhanced default.
		default:
			return "", "", name, false, false, nil, fmt.Errorf("unknown getopt option %q", a)
		}
		i++
	}
	return optstring, longopts, name, quiet, false, nil, nil
}

// parseOptstring maps each short option to whether it takes an argument.
func parseOptstring(s string) map[byte]argSpec {
	m := map[byte]argSpec{}
	s = strings.TrimPrefix(s, "+")
	s = strings.TrimPrefix(s, "-")
	for i := 0; i < len(s); i++ {
		c := s[i]
		spec := noArg
		if i+1 < len(s) && s[i+1] == ':' {
			spec = requiredArg
			i++
			if i+1 < len(s) && s[i+1] == ':' {
				spec = optionalArg
				i++
			}
		}
		m[c] = spec
	}
	return m
}

// parseLongopts maps each long option name to whether it takes an argument.
func parseLongopts(s string) map[string]argSpec {
	m := map[string]argSpec{}
	if s == "" {
		return m
	}
	for _, part := range strings.Split(s, ",") {
		if part == "" {
			continue
		}
		spec := noArg
		switch {
		case strings.HasSuffix(part, "::"):
			spec = optionalArg
			part = strings.TrimSuffix(part, "::")
		case strings.HasSuffix(part, ":"):
			spec = requiredArg
			part = strings.TrimSuffix(part, ":")
		}
		m[part] = spec
	}
	return m
}

// parse normalizes params per the option specs, permuting options before
// operands like GNU getopt.
func parse(params []string, shorts map[byte]argSpec, longs map[string]argSpec, quoted bool) (string, []string) {
	q := func(x string) string {
		if quoted {
			return quote(x)
		}
		return x
	}
	var opts, operands []string
	var errs []string
	endOfOpts := false

	for i := 0; i < len(params); i++ {
		p := params[i]
		switch {
		case endOfOpts || p == "-" || !strings.HasPrefix(p, "-"):
			operands = append(operands, q(p))
		case p == "--":
			endOfOpts = true
		case strings.HasPrefix(p, "--"):
			name, val, hasVal := strings.Cut(strings.TrimPrefix(p, "--"), "=")
			spec, ok := longs[name]
			if !ok {
				errs = append(errs, fmt.Sprintf("unrecognized option '--%s'", name))
				continue
			}
			opts = append(opts, "--"+name)
			switch spec {
			case requiredArg:
				switch {
				case hasVal:
					opts = append(opts, q(val))
				case i+1 < len(params):
					i++
					opts = append(opts, q(params[i]))
				default:
					errs = append(errs, fmt.Sprintf("option '--%s' requires an argument", name))
				}
			case optionalArg:
				if hasVal {
					opts = append(opts, q(val))
				} else {
					opts = append(opts, q(""))
				}
			case noArg:
			}
		default: // a -short cluster
			j := 1
			for j < len(p) {
				c := p[j]
				spec, ok := shorts[c]
				if !ok {
					errs = append(errs, fmt.Sprintf("invalid option -- '%c'", c))
					j++
					continue
				}
				opts = append(opts, "-"+string(c))
				if spec == requiredArg {
					if j+1 < len(p) {
						opts = append(opts, q(p[j+1:]))
					} else if i+1 < len(params) {
						i++
						opts = append(opts, q(params[i]))
					} else {
						errs = append(errs, fmt.Sprintf("option requires an argument -- '%c'", c))
					}
					j = len(p)
					continue
				}
				if spec == optionalArg {
					arg := ""
					if j+1 < len(p) {
						arg = p[j+1:]
					}
					opts = append(opts, q(arg))
					j = len(p)
					continue
				}
				j++
			}
		}
	}

	parts := append(opts, "--")
	parts = append(parts, operands...)
	return " " + strings.Join(parts, " "), errs
}

// quote single-quotes a token the way getopt does, escaping embedded quotes.
func quote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

func (c *Command) help() command.Help {
	return command.Help{
		Description: "Parse the ARGS after -- against the short options in OPTSTRING and the comma-" +
			"separated --long options, then print a normalized, single-quoted argument list with " +
			"options permuted before operands. A script re-reads it with: eval set -- \"$(getopt ...)\".",
		Examples: []command.Example{
			{Command: `getopt -o ab: --long alpha,beta: -- -a -b x file`, Explain: "Normalize the script's arguments."},
		},
		ExitStatus: "0  parsing succeeded.\n1  an argument could not be parsed.\n2  getopt's own options were wrong.",
		Notes: []string{
			"Enhanced mode only: output is always single-quoted; -u/-a/-Q/-T are accepted but do not change behavior.",
		},
	}
}
