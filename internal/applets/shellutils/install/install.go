// Package install implements the install applet: copy files and set their
// permission bits, or create directories, the way the GNU coreutils "install"
// command does. It is the command Makefiles reach for to place a built binary
// in its final location with a chosen mode in one step.
package install

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the install applet.
type Command struct{}

// New returns an install command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "install" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Copy files and set attributes" }

// options holds the parsed command-line switches.
type options struct {
	directory bool   // -d: create directories instead of copying
	createDir bool   // -D: create leading directory components of DEST
	noTarget  bool   // -T: treat DEST as a normal file, never a directory
	target    string // -t: directory into which every SOURCE is copied
	mode      os.FileMode
	preserve  bool // -p: preserve modification/access times
	verbose   bool // -v: print what is being done
}

// Run executes install.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... [-T] SOURCE DEST", stdio.Err)
	directory := fs.BoolP("directory", "d", false, "treat all arguments as directory names; create them")
	createDir := fs.BoolP("create-leading", "D", false, "create all leading components of DEST, then copy SOURCE")
	noTarget := fs.BoolP("no-target-directory", "T", false, "treat DEST as a normal file")
	target := fs.StringP("target-directory", "t", "", "copy all SOURCE arguments into DIRECTORY")
	modeStr := fs.StringP("mode", "m", "755", "set permission mode (as in chmod), instead of rwxr-xr-x")
	preserve := fs.BoolP("preserve-timestamps", "p", false, "apply access/modification times of SOURCE files to DEST")
	verbose := fs.BoolP("verbose", "v", false, "print the name of each created file or directory")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	mode, err := parseMode(*modeStr)
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "install: invalid mode '%s'\n", *modeStr)
		return command.SilentFailure()
	}

	opts := options{
		directory: *directory,
		createDir: *createDir,
		noTarget:  *noTarget,
		target:    *target,
		mode:      mode,
		preserve:  *preserve,
		verbose:   *verbose,
	}

	operands := fs.Args()
	if opts.directory {
		return c.makeDirectories(stdio, operands, opts)
	}
	return c.copyFiles(stdio, operands, opts)
}

// parseMode interprets an octal permission string such as "755" or "0644".
func parseMode(s string) (os.FileMode, error) {
	v, err := strconv.ParseUint(s, 8, 32)
	if err != nil {
		return 0, err
	}
	return os.FileMode(v), nil
}

// makeDirectories implements "install -d": create each operand as a directory,
// including any missing parents, applying the requested mode.
func (c *Command) makeDirectories(stdio command.IO, dirs []string, opts options) error {
	if len(dirs) == 0 {
		_, _ = fmt.Fprintln(stdio.Err, "install: missing operand")
		return command.SilentFailure()
	}

	var failed bool
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, opts.mode); err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "install: cannot create directory %s: %v\n", command.FileError(dir, err), err)
			failed = true
			continue
		}
		// MkdirAll honors umask, so set the exact mode explicitly.
		if err := os.Chmod(dir, opts.mode); err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "install: cannot set mode of %s: %v\n", dir, err)
			failed = true
			continue
		}
		if opts.verbose {
			_, _ = fmt.Fprintf(stdio.Out, "install: creating directory '%s'\n", dir)
		}
	}
	if failed {
		return command.SilentFailure()
	}
	return nil
}

// copyFiles implements the file-copying forms of install. It resolves the
// destination (an explicit -t directory, an existing directory, or a single
// file) and copies every source into place with the requested mode.
func (c *Command) copyFiles(stdio command.IO, operands []string, opts options) error {
	sources, dest, err := c.resolveDest(operands, opts)
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "install: %v\n", err)
		return command.SilentFailure()
	}

	destIsDir := opts.target != "" || (!opts.noTarget && isDir(dest))

	var failed bool
	for _, src := range sources {
		out := dest
		if destIsDir {
			out = filepath.Join(dest, filepath.Base(src))
		}
		if err := c.installOne(stdio, src, out, opts); err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "install: %v\n", err)
			failed = true
		}
	}
	if failed {
		return command.SilentFailure()
	}
	return nil
}

// resolveDest splits operands into the source list and destination according to
// the -t and -T flags and the usual "last operand is the destination" rule.
func (c *Command) resolveDest(operands []string, opts options) (sources []string, dest string, err error) {
	if opts.target != "" {
		if len(operands) == 0 {
			return nil, "", fmt.Errorf("missing file operand")
		}
		return operands, opts.target, nil
	}
	if len(operands) < 2 {
		return nil, "", fmt.Errorf("missing destination file operand after '%s'", lastOrEmpty(operands))
	}
	return operands[:len(operands)-1], operands[len(operands)-1], nil
}

// installOne copies a single source file to dest, creating leading directories
// when -D is set, then applies the mode and (with -p) the source timestamps.
func (c *Command) installOne(stdio command.IO, src, dest string, opts options) error {
	if opts.createDir {
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return fmt.Errorf("cannot create directory for %s: %w", dest, err)
		}
	}

	info, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("cannot stat %s", command.FileError(src, err))
	}
	if info.IsDir() {
		return fmt.Errorf("omitting directory '%s'", src)
	}

	if err := copyFile(src, dest); err != nil {
		return fmt.Errorf("cannot install %s: %w", src, err)
	}
	if err := os.Chmod(dest, opts.mode); err != nil {
		return fmt.Errorf("cannot set mode of %s: %w", dest, err)
	}
	if opts.preserve {
		if err := os.Chtimes(dest, info.ModTime(), info.ModTime()); err != nil {
			return fmt.Errorf("cannot set times of %s: %w", dest, err)
		}
	}
	if opts.verbose {
		_, _ = fmt.Fprintf(stdio.Out, "'%s' -> '%s'\n", src, dest)
	}
	return nil
}

// copyFile copies the contents of src to dest, truncating dest if it exists.
func copyFile(src, dest string) error {
	in, err := os.Open(src) //nolint:gosec // operating on a user-named file
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()

	out, err := os.Create(dest) //nolint:gosec // operating on a user-named file
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return err
	}
	return out.Close()
}

// isDir reports whether path exists and is a directory.
func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// lastOrEmpty returns the last element of s, or "" when s is empty.
func lastOrEmpty(s []string) string {
	if len(s) == 0 {
		return ""
	}
	return s[len(s)-1]
}
