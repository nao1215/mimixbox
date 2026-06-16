// Package stat implements the stat applet: display file status, either in a
// human-readable default layout or with a user-supplied -c format string.
package stat

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/user"
	"strconv"
	"strings"
	"syscall"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the stat applet.
type Command struct{}

// New returns a stat command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "stat" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Display file or file system status" }

// Run executes stat.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... FILE...", stdio.Err).WithHelp(command.Help{
		Description: "Display the status of each FILE: its size, permissions, ownership, and timestamps. " +
			"With -c, print only the fields named by the FORMAT string.",
		Examples: []command.Example{
			{Command: "stat file.txt", Explain: "Show the full status of file.txt."},
			{Command: "stat -c %s file.txt", Explain: "Print only the size of file.txt in bytes."},
			{Command: "stat --printf '%n %s\\n' file.txt", Explain: "Print name and size, interpreting backslash escapes with no trailing newline."},
			{Command: "stat --terse file.txt", Explain: "Print the status as a single space-separated line of raw fields."},
		},
		ExitStatus: "0  success.\n1  an error occurred (e.g. a file could not be stat'd).",
	})
	format := fs.StringP("format", "c", "", "use the specified FORMAT instead of the default; output a trailing newline")
	printf := fs.String("printf", "", "like --format, but interpret backslash escapes and print no trailing newline")
	terse := fs.Bool("terse", false, "print the information in terse form")
	deref := fs.BoolP("dereference", "L", false, "follow links")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	files := fs.Args()
	if len(files) == 0 {
		return command.Failuref("missing operand")
	}

	// --printf takes precedence and is reported by GNU when both are given;
	// here --printf overrides --format. The format/printf strings differ only in
	// whether backslash escapes are honored and whether a newline is appended.
	usePrintf := fs.Changed("printf")

	var firstErr error
	for _, name := range files {
		info, err := lstatOrStat(name, *deref)
		if err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "stat: cannot stat %q: %v\n", name, unwrap(err))
			if firstErr == nil {
				firstErr = command.SilentFailure()
			}
			continue
		}
		var line string
		switch {
		case usePrintf:
			line = applyFormat(unescape(*printf), name, info)
		case *terse:
			line = terseLine(name, info) + "\n"
		case *format != "":
			line = applyFormat(unescape(*format), name, info) + "\n"
		default:
			line = defaultLayout(name, info)
		}
		if _, werr := stdio.Out.Write([]byte(line)); werr != nil {
			return command.Failure(werr)
		}
	}
	return firstErr
}

// lstatOrStat stats name, following a final symlink only when deref is set.
func lstatOrStat(name string, deref bool) (os.FileInfo, error) {
	if deref {
		return os.Stat(name)
	}
	return os.Lstat(name)
}

func unwrap(err error) error {
	var pe *os.PathError
	if errors.As(err, &pe) {
		return pe.Err
	}
	return err
}

// applyFormat renders the GNU -c/--printf format string. It supports the common
// specifiers; an unknown specifier is emitted verbatim. Backslash-escape
// expansion is the caller's responsibility (via unescape).
func applyFormat(format, name string, info os.FileInfo) string {
	sys, _ := info.Sys().(*syscall.Stat_t)

	var b strings.Builder
	for i := 0; i < len(format); i++ {
		if format[i] != '%' || i+1 >= len(format) {
			b.WriteByte(format[i])
			continue
		}
		i++
		switch format[i] {
		case 'n':
			b.WriteString(name)
		case 's':
			fmt.Fprintf(&b, "%d", info.Size())
		case 'F':
			b.WriteString(fileType(info.Mode()))
		case 'a':
			fmt.Fprintf(&b, "%o", info.Mode().Perm())
		case 'A':
			b.WriteString(info.Mode().String())
		case 'f':
			fmt.Fprintf(&b, "%x", info.Mode())
		case 'i':
			if sys != nil {
				fmt.Fprintf(&b, "%d", sys.Ino)
			}
		case 'h':
			if sys != nil {
				fmt.Fprintf(&b, "%d", sys.Nlink)
			}
		case 'u':
			if sys != nil {
				fmt.Fprintf(&b, "%d", sys.Uid)
			}
		case 'U':
			b.WriteString(userName(sys))
		case 'g':
			if sys != nil {
				fmt.Fprintf(&b, "%d", sys.Gid)
			}
		case 'G':
			b.WriteString(groupName(sys))
		case 'X':
			fmt.Fprintf(&b, "%d", atime(sys, info))
		case 'Y':
			fmt.Fprintf(&b, "%d", info.ModTime().Unix())
		case 'Z':
			fmt.Fprintf(&b, "%d", ctime(sys, info))
		case 'b':
			if sys != nil {
				fmt.Fprintf(&b, "%d", sys.Blocks)
			}
		case 'B':
			if sys != nil {
				fmt.Fprintf(&b, "%d", sys.Blksize)
			}
		case '%':
			b.WriteByte('%')
		default:
			b.WriteByte('%')
			b.WriteByte(format[i])
		}
	}
	return b.String()
}

// terseLine renders the single space-separated field line GNU stat prints for
// --terse: name size blocks rawmode uid gid devnumber inode nlink major minor
// atime mtime ctime blocksize.
func terseLine(name string, info os.FileInfo) string {
	sys, _ := info.Sys().(*syscall.Stat_t)
	var (
		blocks, uid, gid, dev, ino, nlink, major, minor uint64
		rawmode                                         uint64
		blksize                                         int64
		at, mt, ct                                      int64
	)
	mt = info.ModTime().Unix()
	rawmode = uint64(info.Mode())
	if sys != nil {
		blocks = uint64(sys.Blocks)
		uid = uint64(sys.Uid)
		gid = uint64(sys.Gid)
		dev = uint64(sys.Dev)
		ino = sys.Ino
		nlink = uint64(sys.Nlink)
		major = uint64(sys.Rdev >> 8 & 0xff)
		minor = uint64(sys.Rdev & 0xff)
		blksize = int64(sys.Blksize)
		at = atime(sys, info)
		ct = ctime(sys, info)
	} else {
		at, ct = mt, mt
	}
	return fmt.Sprintf("%s %d %d %x %d %d %d %d %d %d %d %d %d %d %d",
		name, info.Size(), blocks, rawmode, uid, gid, dev, ino, nlink,
		major, minor, at, mt, ct, blksize)
}

// unescape expands the backslash escapes GNU stat honors inside a format.
func unescape(s string) string {
	r := strings.NewReplacer(`\n`, "\n", `\t`, "\t", `\\`, "\\")
	return r.Replace(s)
}

// userName resolves the owning user's name for %U, falling back to the numeric
// uid when the lookup fails.
func userName(sys *syscall.Stat_t) string {
	if sys == nil {
		return ""
	}
	if u, err := user.LookupId(strconv.FormatUint(uint64(sys.Uid), 10)); err == nil {
		return u.Username
	}
	return strconv.FormatUint(uint64(sys.Uid), 10)
}

// groupName resolves the owning group's name for %G, falling back to the numeric
// gid when the lookup fails.
func groupName(sys *syscall.Stat_t) string {
	if sys == nil {
		return ""
	}
	if g, err := user.LookupGroupId(strconv.FormatUint(uint64(sys.Gid), 10)); err == nil {
		return g.Name
	}
	return strconv.FormatUint(uint64(sys.Gid), 10)
}

// atime returns the access time in Unix epoch seconds, falling back to the
// modification time when the raw stat is unavailable.
func atime(sys *syscall.Stat_t, info os.FileInfo) int64 {
	if sys == nil {
		return info.ModTime().Unix()
	}
	return sys.Atim.Sec
}

// ctime returns the status-change time in Unix epoch seconds, falling back to
// the modification time when the raw stat is unavailable.
func ctime(sys *syscall.Stat_t, info os.FileInfo) int64 {
	if sys == nil {
		return info.ModTime().Unix()
	}
	return sys.Ctim.Sec
}

// defaultLayout renders the multi-line human-readable status block.
func defaultLayout(name string, info os.FileInfo) string {
	var b strings.Builder
	fmt.Fprintf(&b, "  File: %s\n", name)
	fmt.Fprintf(&b, "  Size: %-10d\tType: %s\n", info.Size(), fileType(info.Mode()))
	fmt.Fprintf(&b, "Access: (%04o/%s)\n", info.Mode().Perm(), info.Mode().String())
	if sys, ok := info.Sys().(*syscall.Stat_t); ok {
		fmt.Fprintf(&b, " Inode: %-10d\tLinks: %d\tUid: %d\tGid: %d\n",
			sys.Ino, sys.Nlink, sys.Uid, sys.Gid)
	}
	fmt.Fprintf(&b, "Modify: %s\n", info.ModTime().Format("2006-01-02 15:04:05"))
	return b.String()
}

// fileType maps a file mode to the description GNU stat prints for %F.
func fileType(mode fs.FileMode) string {
	switch {
	case mode.IsDir():
		return "directory"
	case mode&fs.ModeSymlink != 0:
		return "symbolic link"
	case mode&fs.ModeNamedPipe != 0:
		return "fifo"
	case mode&fs.ModeSocket != 0:
		return "socket"
	case mode&fs.ModeDevice != 0:
		if mode&fs.ModeCharDevice != 0 {
			return "character special file"
		}
		return "block special file"
	default:
		return "regular file"
	}
}
