// Package cp implements the cp applet: copy files and directories, with the
// common GNU options (-r/-R, -f, -v, -i, -p) and the symlink-dereference
// controls (-P, -L, -H, -d).
package cp

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
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
func (c *Command) Synopsis() string { return "Copy file(s) to Directory(s)" }

// derefMode selects how cp treats symbolic links in SOURCE.
type derefMode int

const (
	// derefDefault follows command-line symlinks and symlinks within a copied
	// tree (the historical behavior and GNU cp's default without -d/-P/-H).
	derefDefault derefMode = iota
	// derefNever (-P, -d) copies every symlink as a link.
	derefNever
	// derefAlways (-L) follows every symlink, copying what it points at.
	derefAlways
	// derefCmdline (-H) follows symlinks named on the command line but copies
	// symlinks found within a tree as links.
	derefCmdline
)

// backupMode selects how --backup names the backup of an overwritten
// destination, mirroring GNU cp's CONTROL values.
type backupMode int

const (
	// backupNone makes no backup (CONTROL none/off, or --backup absent).
	backupNone backupMode = iota
	// backupSimple appends the simple suffix (CONTROL simple/never), e.g. file~.
	backupSimple
	// backupNumbered makes numbered backups (CONTROL numbered/t), e.g. file.~1~.
	backupNumbered
	// backupExisting makes numbered backups if any already exist, else simple
	// (CONTROL existing/nil).
	backupExisting
)

type options struct {
	recursive   bool
	force       bool
	verbose     bool
	interactive bool
	preserve    bool
	noClobber   bool
	update      bool
	deref       derefMode

	// noTargetDir is -T: never treat the destination as a directory.
	noTargetDir bool
	// parents is --parents: recreate each source's full path prefix under the
	// destination directory.
	parents bool

	// backup selects the backup naming scheme for an overwritten destination.
	backup backupMode
	// suffix is the simple-backup suffix (default "~").
	suffix string
}

// Run executes cp.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... SOURCE... DEST", stdio.Err).WithHelp(command.Help{
		Description: "Copy SOURCE to DEST, or one or more SOURCEs into a DEST directory. " +
			"With -r/-R, directories are copied recursively. By default an existing " +
			"destination is overwritten; -i prompts first and -n never overwrites.",
		Examples: []command.Example{
			{Command: "cp a.txt b.txt", Explain: "Copy a file."},
			{Command: "cp -r src/ dst/", Explain: "Copy a directory tree."},
			{Command: "cp -a src/ dst/", Explain: "Copy recursively, preserving mode and timestamps (= -rp)."},
			{Command: "cp -i a.txt dir/", Explain: "Prompt before overwriting dir/a.txt."},
			{Command: "cp -t dir/ a.txt b.txt", Explain: "Copy each source into dir/ (destination-first)."},
			{Command: "cp --parents src/a/b.txt dst/", Explain: "Recreate the prefix as dst/src/a/b.txt."},
			{Command: "cp -u a.txt dir/", Explain: "Copy only if a.txt is newer than dir/a.txt."},
			{Command: "cp --backup a.txt dir/", Explain: "Back up an existing dir/a.txt before overwriting."},
		},
		ExitStatus: "0  all files were copied successfully.\n1  one or more files could not be copied.",
		Notes: []string{
			"Symlinks: by default they are followed. -P copies them as links, -L always follows, " +
				"-H follows only those named on the command line, and -d is shorthand for -P. -a implies -d.",
		},
	})
	recursive := fs.BoolP("recursive", "r", false, "copy directories recursively (-R is an alias)")
	// -R is the other GNU spelling of -r; pflag cannot give one flag two
	// shorthands, so it is a hidden alias whose value is OR'd into recursive.
	recursiveR := fs.BoolP("recursive-R", "R", false, "copy directories recursively")
	_ = fs.MarkHidden("recursive-R")
	archive := fs.BoolP("archive", "a", false, "same as -rp (recursive and preserve)")
	force := fs.BoolP("force", "f", false, "if an existing destination file cannot be opened, remove it and try again")
	verbose := fs.BoolP("verbose", "v", false, "explain what is being done")
	interactive := fs.BoolP("interactive", "i", false, "prompt before overwrite")
	noClobber := fs.BoolP("no-clobber", "n", false, "do not overwrite an existing file")
	preserve := fs.BoolP("preserve", "p", false, "preserve mode and timestamps")
	noDeref := fs.BoolP("no-dereference", "P", false, "never follow symbolic links in SOURCE")
	deref := fs.BoolP("dereference", "L", false, "always follow symbolic links in SOURCE")
	// -H and -d are short-only in GNU cp; pflag needs a long name, so register
	// hidden long aliases (the -R alias above uses the same trick).
	followCmdline := fs.BoolP("dereference-cmdline", "H", false, "follow command-line symbolic links in SOURCE")
	_ = fs.MarkHidden("dereference-cmdline")
	noDerefPreserve := fs.BoolP("no-deref-preserve-links", "d", false, "same as --no-dereference")
	_ = fs.MarkHidden("no-deref-preserve-links")
	update := fs.BoolP("update", "u", false, "copy only when the SOURCE file is newer than the destination file or when the destination file is missing")
	targetDir := fs.StringP("target-directory", "t", "", "copy all SOURCE arguments into DIRECTORY")
	noTargetDir := fs.BoolP("no-target-directory", "T", false, "treat DEST as a normal file")
	parents := fs.Bool("parents", false, "use full source file name under DIRECTORY")
	// --backup takes an optional CONTROL word; with no value (bare --backup) the
	// default is "existing". pflag's NoOptDefVal makes "--backup" alone valid.
	backup := fs.String("backup", "", "make a backup of each existing destination file (CONTROL: none, numbered, existing, simple)")
	fs.Lookup("backup").NoOptDefVal = "existing"
	suffix := fs.StringP("suffix", "S", "", "override the usual backup suffix")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	backupMode, err := resolveBackup(*backup)
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "cp: %s\n", err)
		return command.SilentFailure()
	}

	opts := options{
		recursive:   *recursive || *recursiveR || *archive,
		force:       *force,
		verbose:     *verbose,
		interactive: *interactive,
		preserve:    *preserve || *archive,
		noClobber:   *noClobber,
		update:      *update,
		deref:       resolveDeref(*deref, *noDeref || *noDerefPreserve, *followCmdline, *archive),
		noTargetDir: *noTargetDir,
		parents:     *parents,
		backup:      backupMode,
		suffix:      resolveSuffix(*suffix),
	}

	if *noTargetDir && *targetDir != "" {
		_, _ = fmt.Fprintf(stdio.Err, "cp: cannot combine --target-directory (-t) and --no-target-directory (-T)\n")
		return command.SilentFailure()
	}

	operands := fs.Args()

	// -t DIR: every operand is a SOURCE copied into DIR. Rearrange the operands
	// into the internal "SOURCE... DEST" shape by appending DIR as the final
	// destination, then proceed through the normal path.
	if *targetDir != "" {
		if len(operands) == 0 {
			_, _ = fmt.Fprintf(stdio.Err, "cp: missing file operand\n")
			return command.SilentFailure()
		}
		operands = append(operands, *targetDir)
		return cp(stdio, operands, opts)
	}

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

	// -T (--no-target-directory): the destination is always a normal file, never
	// a directory, so it accepts exactly one source and must not already be a
	// directory (GNU refuses to overwrite a directory with a non-directory).
	if opts.noTargetDir {
		if len(sources) > 1 {
			_, _ = fmt.Fprintf(stdio.Err, "cp: extra operand '%s'\n", sources[1])
			return command.SilentFailure()
		}
		if di, err := os.Stat(dest); err == nil && di.IsDir() {
			_, _ = fmt.Fprintf(stdio.Err, "cp: cannot overwrite directory '%s' with non-directory\n", dest)
			return command.SilentFailure()
		}
	}

	// With more than one source, GNU cp requires the destination to be an
	// existing directory; otherwise each source would overwrite the last.
	// --parents always copies into the destination directory, so it has the same
	// requirement even with a single source.
	if len(sources) > 1 || opts.parents {
		if di, err := os.Stat(dest); err != nil || !di.IsDir() {
			_, _ = fmt.Fprintf(stdio.Err, "cp: target '%s' is not a directory\n", dest)
			return command.SilentFailure()
		}
	}

	for _, raw := range sources {
		src := os.ExpandEnv(raw)

		// A command-line symlink is copied as a link only with -P/-d; -L, -H
		// and the default all follow it (handled by os.Stat below).
		if li, lerr := os.Lstat(src); lerr == nil && li.Mode()&os.ModeSymlink != 0 && opts.deref == derefNever {
			target := symlinkTarget(src, dest)
			if isSamePath(src, target) {
				_, _ = fmt.Fprintf(stdio.Err, "cp: %s and %s is same.\n", src, dest)
				return command.SilentFailure()
			}
			if err := copySymlink(stdio, src, target, opts); err != nil {
				_, _ = fmt.Fprintf(stdio.Err, "cp: %s\n", err)
				return command.SilentFailure()
			}
			continue
		}

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

		// --parents: recreate src's full directory prefix under the destination
		// directory, so "cp --parents a/b/c.txt dst" creates dst/a/b/c.txt. The
		// effective destination becomes dst/<dir(src)>, created up front.
		effectiveDest := dest
		if opts.parents {
			prefix := filepath.Dir(src)
			if prefix != "." && prefix != string(os.PathSeparator) {
				effectiveDest = filepath.Join(dest, prefix)
				if err := os.MkdirAll(effectiveDest, 0o755); err != nil {
					_, _ = fmt.Fprintf(stdio.Err, "cp: %s\n", err)
					return command.SilentFailure()
				}
			}
		}

		if info.IsDir() {
			if err := cpDir(stdio, src, effectiveDest, opts); err != nil {
				_, _ = fmt.Fprintf(stdio.Err, "cp: %s\n", err)
				return command.SilentFailure()
			}
		} else {
			if err := cpFile(stdio, src, effectiveDest, info, opts); err != nil {
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
	// -T forces the destination to be a normal file; otherwise an existing
	// directory means the file keeps its base name inside it.
	if !opts.noTargetDir {
		if di, err := os.Stat(dest); err == nil && di.IsDir() {
			target = filepath.Join(dest, filepath.Base(src))
		}
	}

	// The early src-vs-dest check cannot see this: when dest is a directory the
	// effective target becomes dest/<base(src)>, which may equal src. Opening
	// that target for writing would truncate the source, so reject it here.
	if isSamePath(src, target) {
		return fmt.Errorf("'%s' and '%s' are the same file", src, target)
	}

	return copyRegularFile(stdio, src, target, info, opts)
}

// copyRegularFile is the single per-file execution path for copying a regular
// file's contents to target. It centralizes the overwrite policy — -n
// (no-clobber), -u (update), -i (interactive prompt), and --backup — so that
// direct copies (cpFile) and files discovered during a recursive walk (cpDir)
// share exactly one decision path. Traversal stays in the callers; per-file
// policy lives here. The caller resolves target (including same-file guards)
// before calling.
func copyRegularFile(stdio command.IO, src, target string, info os.FileInfo, opts options) error {
	// -n: never overwrite an existing destination.
	if opts.noClobber {
		if _, err := os.Stat(target); err == nil {
			return nil // skip this file
		}
	}

	// -u: copy only when src is newer than an existing destination.
	if opts.update {
		if di, err := os.Stat(target); err == nil && !info.ModTime().After(di.ModTime()) {
			return nil // destination is at least as new; skip
		}
	}

	// -i: prompt before overwriting an existing destination.
	if opts.interactive {
		if _, err := os.Stat(target); err == nil {
			if !question(stdio, fmt.Sprintf("cp: overwrite '%s'? ", target)) {
				return nil // skip this file
			}
		}
	}

	// --backup: before overwriting an existing destination, move it aside.
	if err := makeBackup(target, opts); err != nil {
		return err
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

		// Walk reports symlinks via Lstat. Copy them as links for -P/-d and -H
		// (within a tree); otherwise follow them and copy what they point at.
		if info.Mode()&os.ModeSymlink != 0 {
			if opts.deref == derefNever || opts.deref == derefCmdline {
				return copySymlink(stdio, p, target, opts)
			}
			ti, terr := os.Stat(p)
			if terr != nil {
				return terr
			}
			if ti.IsDir() {
				// Following a symlink to a directory within a tree is not
				// recursed into; copying its contents is out of scope.
				return nil
			}
			// A followed symlink yields a regular file; route it through the
			// shared overwrite policy just like any other walked file.
			return copyRegularFile(stdio, p, target, ti, opts)
		}

		if info.IsDir() {
			// Use the source directory's mode (GNU cp does this even without
			// -p); a hardcoded 0755 would widen a private tree such as 0700.
			if err := os.MkdirAll(target, info.Mode().Perm()); err != nil {
				return err
			}
			// MkdirAll is a no-op when target already exists, and umask may have
			// masked the mode on creation; with -p, set the exact source mode.
			if opts.preserve {
				_ = os.Chmod(target, info.Mode().Perm())
			}
			return nil
		}
		// Regular files share the same overwrite policy as direct copies.
		return copyRegularFile(stdio, p, target, info, opts)
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

	// Create the destination with the source file's mode (GNU cp does this even
	// without -p); a hardcoded 0644 would strip the execute bit from scripts and
	// binaries. The mode only takes effect when the file is created, so an
	// existing destination keeps its own permissions.
	mode := info.Mode().Perm()

	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode) //nolint:gosec // user-named destination
	if err != nil {
		// cp -f: if an existing destination cannot be opened (for example it is
		// read-only), remove it and try once more.
		if opts.force && os.IsPermission(err) {
			if rmErr := os.Remove(dst); rmErr == nil {
				out, err = os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode) //nolint:gosec // user-named destination
			}
		}
		if err != nil {
			return err
		}
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

// resolveDeref maps the symlink flags to a derefMode. An explicit -L or -P/-d
// wins over -a's implied -d, and -H applies when nothing stronger is set.
func resolveDeref(deref, noDeref, followCmdline, archive bool) derefMode {
	switch {
	case deref:
		return derefAlways
	case noDeref:
		return derefNever
	case followCmdline:
		return derefCmdline
	case archive:
		return derefNever // -a implies -d
	default:
		return derefDefault
	}
}

// resolveBackup maps a --backup CONTROL word to a backupMode. An empty string
// means --backup was not given (no backup). GNU also honors VERSION_CONTROL but
// the explicit flag value takes precedence here.
func resolveBackup(control string) (backupMode, error) {
	switch control {
	case "":
		return backupNone, nil
	case "none", "off":
		return backupNone, nil
	case "numbered", "t":
		return backupNumbered, nil
	case "existing", "nil":
		return backupExisting, nil
	case "simple", "never":
		return backupSimple, nil
	default:
		return backupNone, fmt.Errorf("invalid argument '%s' for '--backup'", control)
	}
}

// resolveSuffix returns the simple-backup suffix: the explicit --suffix value,
// then $SIMPLE_BACKUP_SUFFIX, then the default "~".
func resolveSuffix(suffix string) string {
	if suffix != "" {
		return suffix
	}
	if env := os.Getenv("SIMPLE_BACKUP_SUFFIX"); env != "" {
		return env
	}
	return "~"
}

// makeBackup moves an existing target aside before it is overwritten, following
// the selected backup scheme. It is a no-op when --backup was not requested or
// the target does not exist.
func makeBackup(target string, opts options) error {
	if opts.backup == backupNone {
		return nil
	}
	if _, err := os.Lstat(target); err != nil {
		return nil // nothing to back up
	}

	mode := opts.backup
	if mode == backupExisting {
		// Numbered if any numbered backup already exists, else simple.
		if numberedBackupsExist(target) {
			mode = backupNumbered
		} else {
			mode = backupSimple
		}
	}

	var backupPath string
	switch mode {
	case backupNumbered:
		backupPath = nextNumberedBackup(target)
	default: // backupSimple
		backupPath = target + opts.suffix
	}
	return os.Rename(target, backupPath)
}

// numberedBackupsExist reports whether any file named "<target>.~N~" exists.
func numberedBackupsExist(target string) bool {
	for n := 1; ; n++ {
		p := fmt.Sprintf("%s.~%d~", target, n)
		if _, err := os.Lstat(p); err != nil {
			return n > 1
		}
	}
}

// nextNumberedBackup returns the first unused "<target>.~N~" name.
func nextNumberedBackup(target string) string {
	for n := 1; ; n++ {
		p := target + ".~" + strconv.Itoa(n) + "~"
		if _, err := os.Lstat(p); err != nil {
			return p
		}
	}
}

// symlinkTarget returns where a command-line symlink src should be written:
// dest itself, or dest/<base(src)> when dest is an existing directory.
func symlinkTarget(src, dest string) string {
	if di, err := os.Stat(dest); err == nil && di.IsDir() {
		return filepath.Join(dest, filepath.Base(src))
	}
	return dest
}

// copySymlink copies the symbolic link src to dst as a link (rather than what
// it points to), honoring -n (skip existing) and -f (replace existing).
func copySymlink(stdio command.IO, src, dst string, opts options) error {
	linkDest, err := os.Readlink(src)
	if err != nil {
		return err
	}
	if _, err := os.Lstat(dst); err == nil {
		if opts.noClobber {
			return nil
		}
		if rmErr := os.Remove(dst); rmErr != nil {
			if !opts.force {
				return rmErr
			}
			_ = os.Remove(dst)
		}
	}
	if err := os.Symlink(linkDest, dst); err != nil {
		return err
	}
	if opts.verbose {
		_, _ = fmt.Fprintf(stdio.Out, "'%s' -> '%s'\n", src, dst)
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
