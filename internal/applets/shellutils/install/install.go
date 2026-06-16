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
	"os/exec"
	"os/user"
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

// backupMode selects how --backup names the backup of an overwritten
// destination, mirroring GNU install/cp CONTROL semantics.
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

// options holds the parsed command-line switches.
type options struct {
	directory bool   // -d: create directories instead of copying
	createDir bool   // -D: create leading directory components of DEST
	noTarget  bool   // -T: treat DEST as a normal file, never a directory
	target    string // -t: directory into which every SOURCE is copied
	mode      os.FileMode
	preserve  bool   // -p: preserve modification/access times
	verbose   bool   // -v: print what is being done
	owner     string // -o: owner name or uid to set via chown
	group     string // -g: group name or gid to set via chown
	strip     bool   // -s: run the system strip on the installed file
	backup    backupMode
	suffix    string // simple-backup suffix (default "~")
}

// Run executes install.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... [-T] SOURCE DEST", stdio.Err).WithHelp(command.Help{
		Description: "Copy SOURCE to DEST, or several SOURCEs to an existing DIRECTORY, while " +
			"setting permission modes. With -d, create the named directories instead of " +
			"copying files.",
		Examples: []command.Example{
			{Command: "install -m 644 file /etc/file", Explain: "Copy file to /etc/file with mode 644."},
			{Command: "install -d /opt/app/bin", Explain: "Create the directory /opt/app/bin."},
			{Command: "install -t /usr/local/bin prog", Explain: "Copy prog into the /usr/local/bin directory."},
		},
		ExitStatus: "0  all files or directories were installed.\n1  an operand was invalid or an install failed.",
		Notes: []string{
			"--owner/-o and --group/-g set ownership via chown after install; as a non-root user this fails and install exits nonzero, matching GNU.",
			"--strip/-s runs the system strip program on the installed file; if strip is not found, install reports an error and fails, matching GNU.",
			"--backup CONTROL is one of none/off, numbered/t, existing/nil, simple/never; the simple suffix defaults to '~' or $SIMPLE_BACKUP_SUFFIX and can be overridden with --suffix/-S.",
		},
	})
	directory := fs.BoolP("directory", "d", false, "treat all arguments as directory names; create them")
	createDir := fs.BoolP("create-leading", "D", false, "create all leading components of DEST, then copy SOURCE")
	noTarget := fs.BoolP("no-target-directory", "T", false, "treat DEST as a normal file")
	target := fs.StringP("target-directory", "t", "", "copy all SOURCE arguments into DIRECTORY")
	modeStr := fs.StringP("mode", "m", "755", "set permission mode (as in chmod), instead of rwxr-xr-x")
	preserve := fs.BoolP("preserve-timestamps", "p", false, "apply access/modification times of SOURCE files to DEST")
	verbose := fs.BoolP("verbose", "v", false, "print the name of each created file or directory")
	owner := fs.StringP("owner", "o", "", "set ownership (super-user only)")
	group := fs.StringP("group", "g", "", "set group ownership, instead of process' current group")
	strip := fs.BoolP("strip", "s", false, "strip symbol tables from installed binaries")
	// --backup takes an optional CONTROL word; with no value (bare --backup)
	// the default is "existing". pflag's NoOptDefVal makes "--backup" alone valid.
	backup := fs.String("backup", "", "make a backup of each existing destination file (CONTROL: none, numbered, existing, simple)")
	fs.Lookup("backup").NoOptDefVal = "existing"
	suffix := fs.StringP("suffix", "S", "", "override the usual backup suffix")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	mode, err := parseMode(*modeStr)
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "install: invalid mode '%s'\n", *modeStr)
		return command.SilentFailure()
	}

	backupMode, err := resolveBackup(*backup)
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "install: %v\n", err)
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
		owner:     *owner,
		group:     *group,
		strip:     *strip,
		backup:    backupMode,
		suffix:    resolveSuffix(*suffix),
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
	if len(sources) > 1 && !destIsDir {
		_, _ = fmt.Fprintf(stdio.Err, "install: target '%s' is not a directory\n", dest)
		return command.SilentFailure()
	}

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

	// Move any existing destination aside before overwriting it.
	if err := makeBackup(dest, opts); err != nil {
		return fmt.Errorf("cannot backup %s: %w", dest, err)
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
	// -o/-g: set ownership via chown. As a non-root user this typically fails
	// with EPERM; like GNU install, report the failure and propagate it.
	if opts.owner != "" || opts.group != "" {
		if err := setOwnership(dest, opts); err != nil {
			return err
		}
	}
	// -s: run the system strip on the installed file. GNU install errors out
	// when strip is unavailable; we mirror that.
	if opts.strip {
		if err := stripFile(dest); err != nil {
			return err
		}
	}
	if opts.verbose {
		_, _ = fmt.Fprintf(stdio.Out, "'%s' -> '%s'\n", src, dest)
	}
	return nil
}

// setOwnership resolves the -o owner and -g group operands to a uid/gid and
// applies them with chown. A value left empty maps to -1 (leave unchanged).
func setOwnership(dest string, opts options) error {
	uid := -1
	gid := -1
	if opts.owner != "" {
		v, ok := lookupUID(opts.owner)
		if !ok {
			return fmt.Errorf("invalid user '%s'", opts.owner)
		}
		uid = v
	}
	if opts.group != "" {
		v, ok := lookupGID(opts.group)
		if !ok {
			return fmt.Errorf("invalid group '%s'", opts.group)
		}
		gid = v
	}
	if err := os.Chown(dest, uid, gid); err != nil {
		return fmt.Errorf("cannot change ownership of %s: %w", dest, err)
	}
	return nil
}

// lookupUID resolves a user name or numeric uid to its integer uid.
func lookupUID(owner string) (int, bool) {
	if u, err := user.Lookup(owner); err == nil {
		if uid, err := strconv.Atoi(u.Uid); err == nil {
			return uid, true
		}
	}
	if uid, err := strconv.Atoi(owner); err == nil {
		return uid, true
	}
	return 0, false
}

// lookupGID resolves a group name or numeric gid to its integer gid.
func lookupGID(group string) (int, bool) {
	if g, err := user.LookupGroup(group); err == nil {
		if gid, err := strconv.Atoi(g.Gid); err == nil {
			return gid, true
		}
	}
	if gid, err := strconv.Atoi(group); err == nil {
		return gid, true
	}
	return 0, false
}

// stripFile runs the system "strip" on dest. Like GNU install, an unavailable
// strip program is an error rather than a silent skip.
func stripFile(dest string) error {
	prog, err := exec.LookPath("strip")
	if err != nil {
		return fmt.Errorf("strip program not found")
	}
	cmd := exec.Command(prog, dest) //nolint:gosec // operating on a user-named file
	if out, err := cmd.CombinedOutput(); err != nil {
		if len(out) > 0 {
			return fmt.Errorf("strip %s: %s", dest, string(out))
		}
		return fmt.Errorf("strip %s: %w", dest, err)
	}
	return nil
}

// resolveBackup maps a --backup CONTROL word to a backupMode. An empty string
// means --backup was not given (no backup).
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
