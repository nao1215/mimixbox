// Package realpath implements the realpath applet: print the resolved absolute
// file name for each operand.
package realpath

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the realpath applet.
type Command struct{}

// New returns a realpath command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "realpath" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Print the resolved absolute file name" }

// Run executes realpath.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... FILE...", stdio.Err).WithHelp(command.Help{
		Description: "Print the resolved, absolute file name of each FILE, expanding all symbolic " +
			"links and removing '.' and '..' components. By default every path component must " +
			"exist; -m allows missing components, -s leaves symbolic links unexpanded, and -z " +
			"separates results with a NUL byte instead of a newline.",
		Examples: []command.Example{
			{Command: "realpath ./foo/../bar", Explain: "Print the canonical absolute path of bar."},
			{Command: "realpath -s /usr/bin/awk", Explain: "Resolve the path without expanding symlinks."},
			{Command: "realpath -m /no/such/dir/file", Explain: "Resolve a path whose components need not exist."},
		},
		ExitStatus: "0  all paths were resolved successfully.\n1  a path could not be resolved or no operand was given.",
	})
	existing := fs.BoolP("canonicalize-existing", "e", false, "all components of the path must exist")
	missing := fs.BoolP("canonicalize-missing", "m", false, "no path components need exist or be a directory")
	noSymlinks := fs.BoolP("no-symlinks", "s", false, "don't expand symlinks")
	zero := fs.BoolP("zero", "z", false, "end each output line with NUL, not newline")
	quiet := fs.BoolP("quiet", "q", false, "suppress most error messages")
	relativeTo := fs.String("relative-to", "", "print the resolved path relative to DIR")
	relativeBase := fs.String("relative-base", "", "print absolute paths unless they are below DIR")
	logical := fs.BoolP("logical", "L", false, "resolve '..' components lexically, do not expand symlinks")
	physical := fs.BoolP("physical", "P", false, "resolve all symlinks (the default)")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	names := fs.Args()
	if len(names) == 0 {
		_, _ = fmt.Fprintln(stdio.Err, "realpath: missing operand")
		return command.SilentFailure()
	}

	end := byte('\n')
	if *zero {
		end = 0
	}

	// -L (logical) resolves '..' lexically without expanding symlinks, the same
	// shape as -s here; -P (physical) is the default full resolution.
	lexical := *noSymlinks || *logical
	_ = physical

	// --relative-base without --relative-to implies --relative-to=BASE (GNU).
	relTo, relBase := *relativeTo, *relativeBase
	if relBase != "" && relTo == "" {
		relTo = relBase
	}
	var canonRelTo, canonRelBase string
	if relTo != "" {
		if canonRelTo, err = resolve(relTo, *missing, lexical); err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "realpath: %s\n", command.FileError(relTo, err))
			return command.SilentFailure()
		}
	}
	if relBase != "" {
		if canonRelBase, err = resolve(relBase, *missing, lexical); err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "realpath: %s\n", command.FileError(relBase, err))
			return command.SilentFailure()
		}
	}

	var failed bool
	for _, name := range names {
		resolved, rerr := resolve(name, *missing, lexical)
		if rerr != nil {
			if !*quiet {
				_, _ = fmt.Fprintf(stdio.Err, "realpath: %s\n", command.FileError(name, rerr))
			}
			failed = true
			continue
		}
		resolved = applyRelative(resolved, relTo, canonRelTo, relBase, canonRelBase)
		_, _ = fmt.Fprintf(stdio.Out, "%s%c", resolved, end)
	}

	// existing is accepted for GNU compatibility; it matches the default
	// behavior of requiring every path component to exist.
	_ = existing

	if failed {
		return command.SilentFailure()
	}
	return nil
}

// applyRelative renders resolved relative to canonRelTo when --relative-to is in
// effect, but only when --relative-base is unset or resolved lies below
// canonRelBase; otherwise the absolute resolved path is returned (GNU semantics).
func applyRelative(resolved, relTo, canonRelTo, relBase, canonRelBase string) string {
	if relTo == "" {
		return resolved
	}
	if relBase != "" && !under(resolved, canonRelBase) {
		return resolved
	}
	rel, err := filepath.Rel(canonRelTo, resolved)
	if err != nil {
		return resolved
	}
	return rel
}

// under reports whether path is canonRelBase itself or lies beneath it.
func under(path, base string) bool {
	return path == base || strings.HasPrefix(path, base+string(filepath.Separator))
}

// resolve canonicalizes name to an absolute path. With noSymlinks it only
// cleans the path (filepath.Abs + filepath.Clean) without expanding symlinks.
// With missing it does not require the path to exist. Otherwise every path
// component must exist and all symlinks are resolved.
func resolve(name string, missing, noSymlinks bool) (string, error) {
	abs, err := filepath.Abs(name)
	if err != nil {
		return "", err
	}

	if noSymlinks {
		return filepath.Clean(abs), nil
	}

	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		if missing {
			return filepath.Clean(abs), nil
		}
		return "", err
	}
	return resolved, nil
}
