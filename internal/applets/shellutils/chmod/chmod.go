// Package chmod implements the chmod applet: change the file mode bits of each
// FILE, with the common GNU options. MODE is either an octal number (e.g. 755)
// or a comma-separated list of symbolic clauses (e.g. u+x,go-w).
package chmod

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the chmod applet.
type Command struct{}

// New returns a chmod command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "chmod" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Change file mode bits" }

type options struct {
	recursive bool
	verbose   bool
	changes   bool
	silent    bool
}

// Run executes chmod.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... MODE FILE...", stdio.Err)
	recursive := fs.BoolP("recursive", "R", false, "change files and directories recursively")
	verbose := fs.BoolP("verbose", "v", false, "output a diagnostic for every file processed")
	changes := fs.BoolP("changes", "c", false, "like verbose but report only when a change is made")
	silent := fs.BoolP("silent", "f", false, "suppress most error messages")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	rest := fs.Args()
	if len(rest) < 2 {
		_, _ = fmt.Fprintf(stdio.Err, "%s: missing operand\n", c.Name())
		return command.SilentFailure()
	}

	opts := options{
		recursive: *recursive,
		verbose:   *verbose,
		changes:   *changes,
		silent:    *silent,
	}

	mode := rest[0]
	return c.chmod(stdio, mode, rest[1:], opts)
}

// chmod changes the mode of every file, continuing past any failures. A failed
// file is reported GNU-style on stderr; the returned error only sets the exit
// code, because its message was already printed.
func (c *Command) chmod(stdio command.IO, mode string, files []string, opts options) error {
	var failErr error
	for _, path := range files {
		path = os.ExpandEnv(path)
		var err error
		if opts.recursive {
			err = c.changeModeRecursive(stdio, path, mode, opts)
		} else {
			err = c.changeMode(stdio, path, mode, opts)
		}
		if err != nil {
			failErr = command.SilentFailure()
		}
	}
	return failErr
}

func (c *Command) changeModeRecursive(stdio command.IO, path, mode string, opts options) error {
	// Collect every path first, then apply modes children-before-parent. Doing
	// the walk up front means changing a directory's bits (which may drop the
	// execute bit needed to traverse it) cannot break the descent into its own
	// children.
	var paths []string
	var walkErr error
	err := filepath.WalkDir(path, func(p string, _ fs.DirEntry, err error) error {
		if err != nil {
			c.reportAccess(stdio, p, opts)
			walkErr = err
			return nil
		}
		paths = append(paths, p)
		return nil
	})
	if err != nil {
		return err
	}
	// Deepest paths first so a directory is modified after its contents.
	for i := len(paths) - 1; i >= 0; i-- {
		if cerr := c.changeMode(stdio, paths[i], mode, opts); cerr != nil {
			walkErr = cerr
		}
	}
	return walkErr
}

func (c *Command) changeMode(stdio command.IO, path, mode string, opts options) error {
	info, err := os.Stat(path)
	if err != nil {
		c.reportAccess(stdio, path, opts)
		return err
	}

	cur := info.Mode()
	newMode, err := applyMode(cur, mode, info.IsDir())
	if err != nil {
		if !opts.silent {
			_, _ = fmt.Fprintf(stdio.Err, "%s: invalid mode: '%s'\n", c.Name(), mode)
		}
		return err
	}

	if err := os.Chmod(path, newMode); err != nil {
		if !opts.silent {
			_, _ = fmt.Fprintf(stdio.Err, "%s: changing permissions of '%s': %v\n", c.Name(), path, unwrap(err))
		}
		return err
	}

	changed := cur.Perm() != newMode.Perm()
	if opts.verbose || (opts.changes && changed) {
		if changed {
			_, _ = fmt.Fprintf(stdio.Out, "mode of '%s' changed from %s (%s) to %s (%s)\n",
				path, octal(cur), cur.Perm().String(), octal(newMode), newMode.Perm().String())
		} else {
			_, _ = fmt.Fprintf(stdio.Out, "mode of '%s' retained as %s (%s)\n",
				path, octal(newMode), newMode.Perm().String())
		}
	}
	return nil
}

// reportAccess writes the GNU "cannot access" diagnostic for a missing file.
func (c *Command) reportAccess(stdio command.IO, path string, opts options) {
	if opts.silent {
		return
	}
	_, _ = fmt.Fprintf(stdio.Err, "%s: cannot access '%s': No such file or directory\n", c.Name(), path)
}

func unwrap(err error) error {
	var pe *os.PathError
	if errors.As(err, &pe) {
		return pe.Err
	}
	return err
}

// octal renders the permission bits (including setuid/setgid/sticky) as a
// 4-digit octal string, e.g. "0755".
func octal(m os.FileMode) string {
	return fmt.Sprintf("%04o", permBits(m))
}

// permBits extracts the 12 permission bits (rwx for ugo plus setuid, setgid and
// sticky) from a FileMode as a plain integer in the conventional octal layout.
func permBits(m os.FileMode) uint32 {
	bits := uint32(m.Perm())
	if m&os.ModeSetuid != 0 {
		bits |= 0o4000
	}
	if m&os.ModeSetgid != 0 {
		bits |= 0o2000
	}
	if m&os.ModeSticky != 0 {
		bits |= 0o1000
	}
	return bits
}

// modeFromBits converts the 12 conventional permission bits back into the
// os.FileMode flags, preserving the non-permission bits of base (file type
// etc.).
func modeFromBits(base os.FileMode, bits uint32) os.FileMode {
	m := base &^ (os.ModePerm | os.ModeSetuid | os.ModeSetgid | os.ModeSticky)
	m |= os.FileMode(bits & 0o777)
	if bits&0o4000 != 0 {
		m |= os.ModeSetuid
	}
	if bits&0o2000 != 0 {
		m |= os.ModeSetgid
	}
	if bits&0o1000 != 0 {
		m |= os.ModeSticky
	}
	return m
}

// applyMode computes the new FileMode for a file whose current mode is cur, given
// a chmod MODE string. mode is either an octal number (e.g. "755", "0644") or a
// comma-separated list of symbolic clauses "[ugoa]*[-+=][rwxXst]*". isDir selects
// the behavior of the "X" perm (execute only for directories or files that are
// already executable). The non-permission bits of cur are preserved.
func applyMode(cur os.FileMode, mode string, isDir bool) (os.FileMode, error) {
	if mode == "" {
		return cur, errors.New("empty mode")
	}

	if isOctal(mode) {
		v, err := strconv.ParseUint(mode, 8, 32)
		if err != nil {
			return cur, fmt.Errorf("invalid octal mode: %q", mode)
		}
		return modeFromBits(cur, uint32(v)), nil
	}

	bits := permBits(cur)
	for _, clause := range strings.Split(mode, ",") {
		var err error
		bits, err = applyClause(bits, clause, isDir)
		if err != nil {
			return cur, err
		}
	}
	return modeFromBits(cur, bits), nil
}

// isOctal reports whether s is a pure octal number (chmod's numeric form).
func isOctal(s string) bool {
	for _, r := range s {
		if r < '0' || r > '7' {
			return false
		}
	}
	return true
}

// applyClause applies one symbolic clause "[ugoa]*[-+=][rwxXst]*" to the current
// permission bits and returns the result. bits holds the 12 conventional
// permission bits (see permBits).
func applyClause(bits uint32, clause string, isDir bool) (uint32, error) {
	// Parse the "who" part: which of user/group/other is affected.
	i := 0
	var hasU, hasG, hasO bool
whoLoop:
	for i < len(clause) {
		switch clause[i] {
		case 'u':
			hasU = true
		case 'g':
			hasG = true
		case 'o':
			hasO = true
		case 'a':
			hasU, hasG, hasO = true, true, true
		default:
			break whoLoop
		}
		i++
	}

	if i >= len(clause) {
		return bits, fmt.Errorf("invalid mode: %q", clause)
	}
	op := clause[i]
	if op != '+' && op != '-' && op != '=' {
		return bits, fmt.Errorf("invalid mode: %q", clause)
	}
	i++

	// When no "who" is given, the clause applies to all three.
	allWho := !hasU && !hasG && !hasO
	if allWho {
		hasU, hasG, hasO = true, true, true
	}

	// rwxMask is the 9-bit window of ordinary permission bits selected by who.
	var rwxMask uint32
	if hasU {
		rwxMask |= 0o700
	}
	if hasG {
		rwxMask |= 0o070
	}
	if hasO {
		rwxMask |= 0o007
	}
	// specMask is the special bits (setuid/setgid/sticky) that this who can set.
	var specMask uint32
	if hasU {
		specMask |= 0o4000
	}
	if hasG {
		specMask |= 0o2000
	}
	// Sticky (t) is conventionally tied to "other"/all rather than a who letter.

	perm, err := permMask(clause[i:], bits, isDir, hasU, hasG, allWho)
	if err != nil {
		return bits, err
	}

	set := perm & (rwxMask | specMask | 0o1000)
	switch op {
	case '+':
		bits |= set
	case '-':
		bits &^= set
	case '=':
		// Clear the affected who's ordinary and special bits, then apply.
		bits &^= rwxMask | specMask | 0o1000
		bits |= set
	}
	return bits, nil
}

// permMask builds a permission-bit mask from the perms part of a clause (the
// "[rwxXst]*" portion), expressed in the full ugo layout. cur and isDir drive
// the "X" perm; the who flags decide which setuid/setgid bits "s" sets.
func permMask(perms string, cur uint32, isDir, hasU, hasG, allWho bool) (uint32, error) {
	var mask uint32
	for _, p := range perms {
		switch p {
		case 'r':
			mask |= 0o0444
		case 'w':
			mask |= 0o0222
		case 'x':
			mask |= 0o0111
		case 'X':
			if isDir || cur&0o0111 != 0 {
				mask |= 0o0111
			}
		case 's':
			if hasU || allWho {
				mask |= 0o4000
			}
			if hasG || allWho {
				mask |= 0o2000
			}
		case 't':
			mask |= 0o1000
		default:
			return 0, fmt.Errorf("invalid mode perm: %q", string(p))
		}
	}
	return mask, nil
}
