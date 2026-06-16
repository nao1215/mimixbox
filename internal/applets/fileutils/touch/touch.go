// Package touch implements the touch applet: update the access and
// modification times of each file to the current time, creating the file if it
// does not yet exist.
package touch

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
	"golang.org/x/sys/unix"
)

// Command is the touch applet.
type Command struct{}

// New returns a touch command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "touch" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string {
	return "Update the access and modification times of each FILE to the current time"
}

type options struct {
	noCreate      bool // -c, --no-create
	accessOnly    bool // -a
	modifyOnly    bool // -m
	noDereference bool // -h, --no-dereference: affect the symlink itself

	// When useTimes is set, atime/mtime hold an explicit timestamp pair derived
	// from --reference or --date instead of "now".
	useTimes bool
	atime    time.Time
	mtime    time.Time
}

// Run executes touch.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... FILE...", stdio.Err).WithHelp(command.Help{
		Description: "Update the access and modification times of each FILE to the current time. " +
			"A FILE that does not exist is created empty, unless -c is given.",
		Examples: []command.Example{
			{Command: "touch report.txt", Explain: "Create report.txt if missing, else update its timestamps."},
			{Command: "touch -c existing.log", Explain: "Update timestamps but never create the file."},
			{Command: "touch -m notes.md", Explain: "Change only the modification time."},
		},
		ExitStatus: "0  all files were touched successfully.\n1  a file could not be created or its times could not be changed.",
	})
	noCreate := fs.BoolP("no-create", "c", false, "do not create any files")
	accessOnly := fs.BoolP("access", "a", false, "change only the access time")
	modifyOnly := fs.BoolP("modify", "m", false, "change only the modification time")
	noDereference := fs.BoolP("no-dereference", "h", false, "affect each symbolic link instead of any referenced file")
	reference := fs.StringP("reference", "r", "", "use this file's times instead of the current time")
	date := fs.StringP("date", "d", "", "parse STRING and use it instead of the current time")
	timeWord := fs.String("time", "", "change the specified time: WORD is access, atime, use, modify or mtime")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	files := fs.Args()
	if len(files) == 0 {
		_, _ = fmt.Fprintln(stdio.Err, "touch: missing file operand")
		return command.SilentFailure()
	}

	accessSel := *accessOnly
	modifySel := *modifyOnly
	// --time=WORD is an alternate way to spell -a / -m.
	switch *timeWord {
	case "":
	case "access", "atime", "use":
		accessSel = true
	case "modify", "mtime":
		modifySel = true
	default:
		_, _ = fmt.Fprintf(stdio.Err, "touch: invalid argument %q for '--time'\n", *timeWord)
		return command.SilentFailure()
	}

	opts := options{
		noCreate:      *noCreate,
		accessOnly:    accessSel,
		modifyOnly:    modifySel,
		noDereference: *noDereference,
	}

	// --reference and --date provide an explicit timestamp. When both are given
	// GNU touch lets --date win for whichever component it sets; here --date,
	// when present, overrides the reference time entirely (the common case).
	if *reference != "" {
		ref, err := referenceTimes(*reference)
		if err != nil {
			_, _ = fmt.Fprintln(stdio.Err, "touch: "+err.Error())
			return command.SilentFailure()
		}
		opts.useTimes = true
		opts.atime = ref.atime
		opts.mtime = ref.mtime
	}
	if *date != "" {
		t, err := parseDate(*date)
		if err != nil {
			_, _ = fmt.Fprintln(stdio.Err, "touch: "+err.Error())
			return command.SilentFailure()
		}
		opts.useTimes = true
		opts.atime = t
		opts.mtime = t
	}

	var firstErr error
	for _, file := range files {
		if err := touch(file, opts); err != nil {
			_, _ = fmt.Fprintln(stdio.Err, "touch: "+err.Error())
			if firstErr == nil {
				firstErr = command.SilentFailure()
			}
		}
	}
	return firstErr
}

// touch updates the access and modification times of file. When the file does
// not exist it is created, unless -c/--no-create was given. The -a and -m
// options restrict which of the two timestamps is changed; --reference/--date
// supply an explicit timestamp instead of "now"; -h/--no-dereference acts on a
// symbolic link itself rather than its target.
func touch(file string, opts options) error {
	path := os.ExpandEnv(file)

	info, statErr := os.Lstat(path)
	if errors.Is(statErr, os.ErrNotExist) {
		if opts.noCreate {
			return nil
		}
		f, err := os.Create(path) //nolint:gosec // operating on a user-named file is the whole point
		if err != nil {
			return err
		}
		if err := f.Close(); err != nil {
			return err
		}
		info, statErr = os.Lstat(path)
		if statErr != nil {
			return statErr
		}
	} else if statErr != nil {
		return statErr
	}

	base := time.Now().Local()
	if opts.useTimes {
		// Both default to the explicit timestamp; -a / -m below narrow it.
		base = opts.atime
	}
	atime, mtime := base, base
	if opts.useTimes {
		atime, mtime = opts.atime, opts.mtime
	}

	// The current access time is not available portably, so when only one of
	// -a / -m is given we keep the other timestamp at the file's existing value.
	if opts.accessOnly && !opts.modifyOnly {
		mtime = info.ModTime()
	}
	if opts.modifyOnly && !opts.accessOnly {
		atime = info.ModTime()
	}

	if opts.noDereference {
		return lutimes(path, atime, mtime)
	}
	return os.Chtimes(path, atime, mtime)
}

// lutimes sets the access and modification times of path without following a
// symbolic link, so a symlink's own timestamps are changed rather than its
// target's.
func lutimes(path string, atime, mtime time.Time) error {
	tv := []unix.Timeval{
		unix.NsecToTimeval(atime.UnixNano()),
		unix.NsecToTimeval(mtime.UnixNano()),
	}
	return unix.Lutimes(path, tv)
}

// refTimes carries the access and modification times read from a --reference
// file.
type refTimes struct {
	atime time.Time
	mtime time.Time
}

// referenceTimes returns the access and modification times of the --reference
// file. The link itself is not dereferenced is left to the caller; here we
// follow GNU touch and read the referenced file's times.
func referenceTimes(ref string) (refTimes, error) {
	info, err := os.Stat(os.ExpandEnv(ref))
	if err != nil {
		return refTimes{}, fmt.Errorf("failed to get attributes of %q: %w", ref, err)
	}
	mtime := info.ModTime()
	atime := mtime
	if st, ok := info.Sys().(*syscall.Stat_t); ok {
		atime = time.Unix(st.Atim.Sec, st.Atim.Nsec)
	}
	return refTimes{atime: atime, mtime: mtime}, nil
}

// parseDate parses a --date STRING. It accepts an @UNIX seconds form and a few
// common absolute formats (RFC3339 and "2006-01-02 15:04:05" with or without a
// time-of-day component).
func parseDate(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "@") {
		secs, err := strconv.ParseInt(strings.TrimPrefix(s, "@"), 10, 64)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid date format %q", s)
		}
		return time.Unix(secs, 0), nil
	}
	formats := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04",
		"2006-01-02",
	}
	for _, f := range formats {
		if t, err := time.ParseInLocation(f, s, time.Local); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid date format %q", s)
}
