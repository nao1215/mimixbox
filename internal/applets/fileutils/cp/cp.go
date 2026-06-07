// Package cp implements the cp applet: copy files and directories, with the
// common GNU options (-r/-R, -f, -v, -i, -p).
package cp

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the cp applet.
type Command struct{}

// New returns a cp command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "cp" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Copy file(s) otr Directory(s)" }

type options struct {
	recursive   bool
	force       bool
	verbose     bool
	interactive bool
	preserve    bool
}

// Run executes cp.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... SOURCE... DEST", stdio.Err)
	recursive := fs.BoolP("recursive", "r", false, "copy directories recursively")
	recursiveR := fs.BoolP("Recursive", "R", false, "copy directories recursively")
	force := fs.BoolP("force", "f", false, "if an existing destination file cannot be opened, remove it and try again")
	verbose := fs.BoolP("verbose", "v", false, "explain what is being done")
	interactive := fs.BoolP("interactive", "i", false, "prompt before overwrite")
	preserve := fs.BoolP("preserve", "p", false, "preserve mode and timestamps")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	opts := options{
		recursive:   *recursive || *recursiveR,
		force:       *force,
		verbose:     *verbose,
		interactive: *interactive,
		preserve:    *preserve,
	}

	operands := fs.Args()
	if len(operands) == 0 {
		_, _ = fmt.Fprintf(stdio.Err, "cp: missing file operand\n")
		return command.SilentFailure()
	}
	if len(operands) == 1 {
		_, _ = fmt.Fprintf(stdio.Err, "cp: missing destination file operand after '%s'\n", operands[0])
		return command.SilentFailure()
	}

	return cp(stdio, operands, opts)
}

// cp copies every source operand to the final destination operand. With more
// than one source, the destination must be a directory.
func cp(stdio command.IO, operands []string, opts options) error {
	dest := os.ExpandEnv(operands[len(operands)-1])
	sources := operands[:len(operands)-1]

	// With more than one source, GNU cp requires the destination to be an
	// existing directory; otherwise each source would overwrite the last.
	if len(sources) > 1 {
		if di, err := os.Stat(dest); err != nil || !di.IsDir() {
			_, _ = fmt.Fprintf(stdio.Err, "cp: target '%s' is not a directory\n", dest)
			return command.SilentFailure()
		}
	}

	for _, raw := range sources {
		src := os.ExpandEnv(raw)

		info, err := os.Stat(src)
		if err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "cp: %s\n", command.FileError(src, err))
			return command.SilentFailure()
		}

		if info.IsDir() && !opts.recursive {
			_, _ = fmt.Fprintf(stdio.Err, "cp: --recursive is not specified: omitting directory: %s\n", src)
			return command.SilentFailure()
		}

		if isSamePath(src, dest) {
			_, _ = fmt.Fprintf(stdio.Err, "cp: %s and %s is same.\n", src, dest)
			return command.SilentFailure()
		}

		if info.IsDir() {
			if err := cpDir(stdio, src, dest, opts); err != nil {
				_, _ = fmt.Fprintf(stdio.Err, "cp: %s\n", err)
				return command.SilentFailure()
			}
		} else {
			if err := cpFile(stdio, src, dest, info, opts); err != nil {
				_, _ = fmt.Fprintf(stdio.Err, "cp: %s\n", err)
				return command.SilentFailure()
			}
		}
	}
	return nil
}

// cpFile copies a single regular file to dest. When dest is an existing
// directory, the file keeps its base name inside it.
func cpFile(stdio command.IO, src, dest string, info os.FileInfo, opts options) error {
	target := dest
	if di, err := os.Stat(dest); err == nil && di.IsDir() {
		target = filepath.Join(dest, filepath.Base(src))
	}

	// The early src-vs-dest check cannot see this: when dest is a directory the
	// effective target becomes dest/<base(src)>, which may equal src. Opening
	// that target for writing would truncate the source, so reject it here.
	if isSamePath(src, target) {
		return fmt.Errorf("'%s' and '%s' are the same file", src, target)
	}

	if opts.interactive {
		if _, err := os.Stat(target); err == nil {
			if !question(stdio, fmt.Sprintf("cp: overwrite '%s'? ", target)) {
				return nil // skip this file
			}
		}
	}

	if err := copyFileContents(src, target, info, opts); err != nil {
		return err
	}

	if opts.verbose {
		_, _ = fmt.Fprintf(stdio.Out, "'%s' -> '%s'\n", src, target)
	}
	return nil
}

// cpDir copies the directory tree rooted at src under dest. As GNU cp does,
// when dest already exists the tree is placed at dest/<base(src)>.
func cpDir(stdio command.IO, src, dest string, opts options) error {
	root := dest
	if di, err := os.Stat(dest); err == nil && di.IsDir() {
		root = filepath.Join(dest, filepath.Base(src))
	}

	// Refuse to copy a directory into itself or one of its own descendants;
	// filepath.Walk would otherwise recurse into the growing destination.
	if isSubpath(root, src) {
		return fmt.Errorf("cannot copy a directory, '%s', into itself, '%s'", src, root)
	}

	return filepath.Walk(src, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, p)
		if err != nil {
			return err
		}
		target := filepath.Join(root, rel)

		if info.IsDir() {
			mode := os.FileMode(0755)
			if opts.preserve {
				mode = info.Mode().Perm()
			}
			if err := os.MkdirAll(target, mode); err != nil {
				return err
			}
			return nil
		}
		if err := copyFileContents(p, target, info, opts); err != nil {
			return err
		}
		if opts.verbose {
			_, _ = fmt.Fprintf(stdio.Out, "'%s' -> '%s'\n", p, target)
		}
		return nil
	})
}

// isSubpath reports whether path is base itself or a descendant of base, after
// resolving both to absolute, cleaned paths.
func isSubpath(path, base string) bool {
	pAbs, err1 := filepath.Abs(path)
	bAbs, err2 := filepath.Abs(base)
	if err1 != nil || err2 != nil {
		return false
	}
	if pAbs == bAbs {
		return true
	}
	return strings.HasPrefix(pAbs, bAbs+string(os.PathSeparator))
}

// copyFileContents writes src's contents to dst, honoring -p (mode and
// timestamps).
func copyFileContents(src, dst string, info os.FileInfo, opts options) error {
	in, err := os.Open(src) //nolint:gosec // operating on a user-named file is the whole point
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()

	mode := os.FileMode(0644)
	if opts.preserve {
		mode = info.Mode().Perm()
	}

	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode) //nolint:gosec // user-named destination
	if err != nil {
		return err
	}

	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}

	if opts.preserve {
		_ = os.Chmod(dst, info.Mode().Perm())
		_ = os.Chtimes(dst, info.ModTime(), info.ModTime())
	}
	return nil
}

// isSamePath reports whether src and dest resolve to the same absolute path.
func isSamePath(src, dest string) bool {
	s, err := filepath.Abs(src)
	if err != nil {
		return false
	}
	d, err := filepath.Abs(dest)
	if err != nil {
		return false
	}
	return s == d
}

// question asks the user a yes/no prompt on stdio.Out and reads the answer from
// stdio.In, returning true for an affirmative reply.
func question(stdio command.IO, prompt string) bool {
	_, _ = fmt.Fprint(stdio.Out, prompt)
	scanner := bufio.NewScanner(stdio.In)
	if !scanner.Scan() {
		return false
	}
	answer := strings.ToLower(strings.TrimSpace(scanner.Text()))
	return answer == "y" || answer == "yes"
}
