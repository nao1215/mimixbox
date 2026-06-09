// Package ls implements the ls applet: list directory contents. It covers the
// everyday desktop subset - plain, -1, -a, -A, -d, -l, -F, -h, -R - with stable
// name sorting and deterministic error reporting.
package ls

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the ls applet.
type Command struct{}

// New returns an ls command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "ls" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "List directory contents" }

type options struct {
	all       bool // -a: include . and ..
	almostAll bool // -A: include dotfiles but not . and ..
	long      bool // -l
	dirSelf   bool // -d
	classify  bool // -F
	human     bool // -h
	recursive bool // -R
}

// Run executes ls.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... [FILE]...", stdio.Err).WithHelp(command.Help{
		Description: "List information about FILEs (the current directory by default), sorted by name.",
		Examples: []command.Example{
			{Command: "ls -la", Explain: "Long format, including dotfiles."},
			{Command: "ls -R dir", Explain: "List dir and its subdirectories recursively."},
		},
		ExitStatus: "0  success.\n2  a FILE could not be accessed.",
	})
	all := fs.BoolP("all", "a", false, "do not ignore entries starting with .")
	almost := fs.BoolP("almost-all", "A", false, "like -a but omit . and ..")
	long := fs.BoolP("long", "l", false, "use a long listing format")
	dirSelf := fs.BoolP("directory", "d", false, "list directories themselves, not their contents")
	classify := fs.BoolP("classify", "F", false, "append an indicator (one of */=@|) to entries")
	human := fs.BoolP("human-readable", "h", false, "with -l, print sizes like 1K 234M")
	recursive := fs.BoolP("recursive", "R", false, "list subdirectories recursively")
	_ = fs.BoolP("one-per-line", "1", false, "list one file per line (the default for non-terminals)")
	_ = fs.String("color", "never", "colorize the output; only 'never' is supported")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	opts := options{all: *all, almostAll: *almost, long: *long, dirSelf: *dirSelf, classify: *classify, human: *human, recursive: *recursive}

	operands := fs.Args()
	if len(operands) == 0 {
		operands = []string{"."}
	}

	var files []string
	var dirs []string
	exitErr := false
	for _, name := range operands {
		info, err := os.Lstat(name)
		if err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "ls: cannot access '%s': %s\n", name, errMessage(err))
			exitErr = true
			continue
		}
		if info.IsDir() && !opts.dirSelf {
			dirs = append(dirs, name)
		} else {
			files = append(files, name)
		}
	}

	if len(files) > 0 {
		c.listNames(stdio.Out, "", files, opts)
		if len(dirs) > 0 {
			_, _ = fmt.Fprintln(stdio.Out)
		}
	}

	header := len(operands) > 1 || opts.recursive
	for i, dir := range dirs {
		if header {
			if i > 0 || len(files) > 0 {
				_, _ = fmt.Fprintln(stdio.Out)
			}
			_, _ = fmt.Fprintf(stdio.Out, "%s:\n", dir)
		}
		if err := c.listDir(stdio.Out, stdio.Err, dir, opts); err != nil {
			exitErr = true
		}
	}

	if exitErr {
		return &command.ExitError{Code: 2}
	}
	return nil
}

// listDir lists one directory's entries, recursing when -R is set.
func (c *Command) listDir(out, errw io.Writer, dir string, opts options) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		_, _ = fmt.Fprintf(errw, "ls: cannot open directory '%s': %s\n", dir, errMessage(err))
		return err
	}

	names := make([]string, 0, len(entries)+2)
	if opts.all {
		names = append(names, ".", "..")
	}
	for _, e := range entries {
		if !opts.all && !opts.almostAll && strings.HasPrefix(e.Name(), ".") {
			continue
		}
		names = append(names, e.Name())
	}
	sort.Strings(names)

	c.listNames(out, dir, names, opts)

	if opts.recursive {
		var subdirs []string
		for _, n := range names {
			if n == "." || n == ".." {
				continue
			}
			full := filepath.Join(dir, n)
			if info, err := os.Lstat(full); err == nil && info.IsDir() {
				subdirs = append(subdirs, full)
			}
		}
		for _, sd := range subdirs {
			_, _ = fmt.Fprintln(out)
			_, _ = fmt.Fprintf(out, "%s:\n", sd)
			if err := c.listDir(out, errw, sd, opts); err != nil {
				return err
			}
		}
	}
	return nil
}

// listNames prints the given names (which live in dir) in the selected format.
func (c *Command) listNames(out io.Writer, dir string, names []string, opts options) {
	if !opts.long {
		for _, n := range names {
			_, _ = fmt.Fprintln(out, n+c.indicator(dir, n, opts))
		}
		return
	}
	for _, n := range names {
		_, _ = fmt.Fprintln(out, c.longLine(dir, n, opts))
	}
}

// indicator returns the -F suffix for a name, or "" when -F is off.
func (c *Command) indicator(dir, name string, opts options) string {
	if !opts.classify {
		return ""
	}
	info, err := os.Lstat(pathOf(dir, name))
	if err != nil {
		return ""
	}
	switch {
	case info.IsDir():
		return "/"
	case info.Mode()&os.ModeSymlink != 0:
		return "@"
	case info.Mode()&0o111 != 0 && info.Mode().IsRegular():
		return "*"
	default:
		return ""
	}
}

// longLine formats one entry for -l.
func (c *Command) longLine(dir, name string, opts options) string {
	info, err := os.Lstat(pathOf(dir, name))
	if err != nil {
		return name
	}
	mode := modeString(info)
	nlink := uint64(1)
	owner, group := "?", "?"
	if st, ok := info.Sys().(*syscall.Stat_t); ok {
		nlink = uint64(st.Nlink)
		owner = lookupUser(st.Uid)
		group = lookupGroup(st.Gid)
	}
	size := sizeString(info.Size(), opts.human)
	when := timeString(info.ModTime())
	display := name + c.indicator(dir, name, opts)
	if info.Mode()&os.ModeSymlink != 0 {
		if target, err := os.Readlink(pathOf(dir, name)); err == nil {
			display = name + " -> " + target
		}
	}
	return fmt.Sprintf("%s %d %s %s %s %s %s", mode, nlink, owner, group, size, when, display)
}

func pathOf(dir, name string) string {
	if dir == "" || name == "." || name == ".." {
		if dir == "" {
			return name
		}
	}
	return filepath.Join(dir, name)
}

// modeString renders the rwx-style permission string for ls -l.
func modeString(info os.FileInfo) string {
	m := info.Mode()
	var b strings.Builder
	switch {
	case m&os.ModeDir != 0:
		b.WriteByte('d')
	case m&os.ModeSymlink != 0:
		b.WriteByte('l')
	case m&os.ModeDevice != 0 && m&os.ModeCharDevice != 0:
		b.WriteByte('c')
	case m&os.ModeDevice != 0:
		b.WriteByte('b')
	case m&os.ModeNamedPipe != 0:
		b.WriteByte('p')
	case m&os.ModeSocket != 0:
		b.WriteByte('s')
	default:
		b.WriteByte('-')
	}
	const rwx = "rwxrwxrwx"
	perm := m.Perm()
	for i := 0; i < 9; i++ {
		if perm&(1<<uint(8-i)) != 0 {
			b.WriteByte(rwx[i])
		} else {
			b.WriteByte('-')
		}
	}
	return b.String()
}

func sizeString(size int64, human bool) string {
	if !human {
		return strconv.FormatInt(size, 10)
	}
	const unit = 1024
	if size < unit {
		return strconv.FormatInt(size, 10)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	value := float64(size) / float64(div)
	suffix := "KMGTPE"[exp]
	if value < 10 {
		return fmt.Sprintf("%.1f%c", value, suffix)
	}
	return fmt.Sprintf("%.0f%c", value, suffix)
}

// timeString formats a modification time the way ls does (recent vs old).
func timeString(t time.Time) string {
	now := time.Now()
	sixMonths := time.Hour * 24 * 182
	if now.Sub(t) > sixMonths || t.Sub(now) > sixMonths {
		return t.Format("Jan _2  2006")
	}
	return t.Format("Jan _2 15:04")
}

var userCache = map[uint32]string{}
var groupCache = map[uint32]string{}

func lookupUser(uid uint32) string {
	if name, ok := userCache[uid]; ok {
		return name
	}
	name := strconv.FormatUint(uint64(uid), 10)
	if u, err := user.LookupId(name); err == nil {
		name = u.Username
	}
	userCache[uid] = name
	return name
}

func lookupGroup(gid uint32) string {
	if name, ok := groupCache[gid]; ok {
		return name
	}
	name := strconv.FormatUint(uint64(gid), 10)
	if g, err := user.LookupGroupId(name); err == nil {
		name = g.Name
	}
	groupCache[gid] = name
	return name
}

// errMessage returns the human-friendly tail of a path error.
func errMessage(err error) string {
	if os.IsNotExist(err) {
		return "No such file or directory"
	}
	if os.IsPermission(err) {
		return "Permission denied"
	}
	return err.Error()
}
