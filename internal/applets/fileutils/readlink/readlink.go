// Package readlink implements the readlink applet: print the target of a
// symbolic link, or with -f the canonicalized absolute path.
package readlink

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the readlink applet.
type Command struct{}

// New returns a readlink command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "readlink" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Print resolved symbolic links or canonical file names" }

// Run executes readlink.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... FILE...", stdio.Err).WithHelp(command.Help{
		Description: "Print the target of each symbolic link FILE. With -f, print the canonicalized " +
			"absolute path, following every symlink in the chain.",
		Examples: []command.Example{
			{Command: "readlink /usr/bin/vi", Explain: "Print the immediate target of the symlink."},
			{Command: "readlink -f /usr/bin/vi", Explain: "Print the fully resolved canonical path."},
			{Command: "readlink -n link", Explain: "Print the target without a trailing newline."},
		},
		ExitStatus: "0  every operand was resolved.\n1  an operand was not a symlink or could not be resolved.",
	})
	canon := fs.BoolP("canonicalize", "f", false, "canonicalize by following every symlink")
	canonExisting := fs.BoolP("canonicalize-existing", "e", false, "canonicalize, but all components must exist")
	canonMissing := fs.BoolP("canonicalize-missing", "m", false, "canonicalize without requiring existence")
	quiet := fs.BoolP("no-newline", "n", false, "do not output the trailing newline")
	zero := fs.BoolP("zero", "z", false, "end each output line with NUL, not newline")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	files := fs.Args()
	if len(files) == 0 {
		return command.Failuref("missing operand")
	}

	mode := canonMode{
		canonicalize: *canon,
		existing:     *canonExisting,
		missing:      *canonMissing,
	}

	var firstErr error
	for _, name := range files {
		target, err := resolve(name, mode)
		if err != nil {
			if firstErr == nil {
				firstErr = command.SilentFailure()
			}
			continue
		}
		end := "\n"
		switch {
		case *zero:
			end = "\x00"
		case *quiet:
			end = ""
		}
		if _, werr := io.WriteString(stdio.Out, target+end); werr != nil {
			return command.Failure(werr)
		}
	}
	return firstErr
}

// canonMode selects how resolve canonicalizes a path: -f follows the chain and
// requires the directory portion to exist, -e requires every component
// (including the last) to exist, and -m canonicalizes without requiring that
// any component exists.
type canonMode struct {
	canonicalize bool // -f
	existing     bool // -e
	missing      bool // -m
}

func (m canonMode) any() bool { return m.canonicalize || m.existing || m.missing }

// resolve returns the link target for name. With a canonicalize mode it returns
// a canonicalized absolute path; otherwise it reads the link directly, which
// fails when name is not a symlink (matching GNU readlink).
//
//   - -f: follow every symlink; the directory portion must exist.
//   - -e: like -f, but the final component must also exist.
//   - -m: canonicalize lexically; no component is required to exist.
func resolve(name string, mode canonMode) (string, error) {
	if !mode.any() {
		target, err := os.Readlink(name)
		if err != nil {
			return "", fmt.Errorf("readlink: %w", err)
		}
		return target, nil
	}

	abs, err := filepath.Abs(name)
	if err != nil {
		return "", err
	}

	if mode.missing {
		// -m never touches the filesystem; resolve as far as it exists and
		// canonicalize the remainder lexically.
		return canonicalizeMissing(abs)
	}

	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		if mode.existing {
			// -e requires every component, including the last, to exist.
			return "", err
		}
		// -f tolerates a missing final component: resolve the parent and
		// re-append the base name.
		dir := filepath.Dir(abs)
		base := filepath.Base(abs)
		rdir, derr := filepath.EvalSymlinks(dir)
		if derr != nil {
			return "", derr
		}
		return filepath.Join(rdir, base), nil
	}
	return resolved, nil
}

// canonicalizeMissing resolves the longest existing prefix of abs with
// EvalSymlinks and re-appends the remaining (non-existent) components
// lexically, so the whole path is canonicalized without requiring it to exist.
func canonicalizeMissing(abs string) (string, error) {
	abs = filepath.Clean(abs)
	rest := ""
	cur := abs
	for {
		if resolved, err := filepath.EvalSymlinks(cur); err == nil {
			if rest == "" {
				return resolved, nil
			}
			return filepath.Join(resolved, rest), nil
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			// Reached the root without finding an existing prefix.
			return abs, nil
		}
		rest = filepath.Join(filepath.Base(cur), rest)
		cur = parent
	}
}
