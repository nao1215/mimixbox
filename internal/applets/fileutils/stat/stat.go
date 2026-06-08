// Package stat implements the stat applet: display file status, either in a
// human-readable default layout or with a user-supplied -c format string.
package stat

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
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
	fs := command.NewFlagSet(c.Name(), "[OPTION]... FILE...", stdio.Err)
	format := fs.StringP("format", "c", "", "use the specified FORMAT instead of the default")
	deref := fs.BoolP("dereference", "L", false, "follow links")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	files := fs.Args()
	if len(files) == 0 {
		return command.Failuref("missing operand")
	}

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
		if *format != "" {
			line = applyFormat(*format, name, info) + "\n"
		} else {
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

// applyFormat renders the GNU -c format string. It supports the common
// specifiers; an unknown specifier is emitted verbatim.
func applyFormat(format, name string, info os.FileInfo) string {
	format = unescape(format)
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
		case 'g':
			if sys != nil {
				fmt.Fprintf(&b, "%d", sys.Gid)
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

// unescape expands the backslash escapes GNU stat honours inside a format.
func unescape(s string) string {
	r := strings.NewReplacer(`\n`, "\n", `\t`, "\t", `\\`, "\\")
	return r.Replace(s)
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
