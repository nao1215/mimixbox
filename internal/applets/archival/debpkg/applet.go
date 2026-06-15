package debpkg

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the shared implementation behind the dpkg-deb and dpkg applets.
// Both are read-only front ends over a local .deb file (no package database);
// the name field selects which CLI surface to present. dpkg-deb exposes the
// full inspect/extract surface (-c, -x, -X, -e, -I, -f), while dpkg exposes the
// rescue subset (-x, -X, -c) and rejects database operations with a clear
// error.
type Command struct {
	name string
}

// NewDpkgDeb returns the dpkg-deb applet.
func NewDpkgDeb() *Command { return &Command{name: "dpkg-deb"} }

// NewDpkg returns the dpkg applet.
func NewDpkg() *Command { return &Command{name: "dpkg"} }

// Name returns the command name.
func (c *Command) Name() string { return c.name }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string {
	if c.name == "dpkg" {
		return "Inspect and unpack local Debian .deb files"
	}
	return "Inspect and extract Debian .deb archives"
}

// Run executes the applet.
func (c *Command) Run(ctx context.Context, stdio command.IO, args []string) error {
	if c.name == "dpkg" {
		return c.runDpkg(stdio, args)
	}
	return c.runDpkgDeb(stdio, args)
}

// runDpkgDeb implements the dpkg-deb CLI.
func (c *Command) runDpkgDeb(stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "{-c|-x|-X|-e|-I|-f} ARCHIVE [DIRECTORY|FIELD...]", stdio.Err).WithHelp(c.helpDpkgDeb())
	contents := fs.BoolP("contents", "c", false, "list contents of the data tarball")
	extract := fs.BoolP("extract", "x", false, "extract the data tarball into DIRECTORY")
	vextract := fs.BoolP("vextract", "X", false, "extract and list files as they are extracted")
	control := fs.BoolP("control", "e", false, "extract control files into DIRECTORY (default DEBIAN)")
	info := fs.BoolP("info", "I", false, "print control file information")
	field := fs.BoolP("field", "f", false, "show the named control field(s): ARCHIVE [FIELD...]")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	rest := fs.Args()
	if len(rest) == 0 {
		_, _ = fmt.Fprintf(stdio.Err, "%s: need an action option and an archive; try '%s --help'\n", c.Name(), c.Name())
		return command.SilentFailure()
	}
	pkg, err := Open(rest[0])
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "%s: %v\n", c.Name(), err)
		return command.SilentFailure()
	}

	var runErr error
	switch {
	case *contents:
		runErr = listContents(stdio, pkg)
	case *extract || *vextract:
		runErr = doExtract(stdio, pkg, rest, *vextract)
	case *control:
		runErr = extractControl(pkg, rest)
	case *info:
		runErr = printInfo(stdio, pkg)
	case *field:
		runErr = printField(stdio, pkg, rest[1:])
	default:
		_, _ = fmt.Fprintf(stdio.Err, "%s: need an action option (-c, -x, -X, -e, -I or -f); try '%s --help'\n", c.Name(), c.Name())
		return command.SilentFailure()
	}
	if runErr != nil {
		_, _ = fmt.Fprintf(stdio.Err, "%s: %v\n", c.Name(), runErr)
		return command.SilentFailure()
	}
	return nil
}

// runDpkg implements the dpkg rescue subset.
func (c *Command) runDpkg(stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "{-x|-X|-c} ARCHIVE [DIRECTORY]", stdio.Err).WithHelp(c.helpDpkg())
	extract := fs.BoolP("extract", "x", false, "extract the data tarball of ARCHIVE into DIRECTORY")
	vextract := fs.BoolP("vextract", "X", false, "extract and list files as they are extracted")
	contents := fs.BoolP("contents", "c", false, "list the contents of ARCHIVE's data tarball")
	// Database operations: accepted so we can reject them with a clear message
	// instead of pflag's generic "unknown flag" error.
	install := fs.BoolP("install", "i", false, "(unsupported) install a package")
	remove := fs.BoolP("remove", "r", false, "(unsupported) remove a package")
	list := fs.BoolP("list", "l", false, "(unsupported) list installed packages")
	configure := fs.Bool("configure", false, "(unsupported) configure an unpacked package")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	if *install || *remove || *list || *configure {
		_, _ = fmt.Fprintf(stdio.Err, "%s: package-database operations are not supported in this build; "+
			"use -x/-X to unpack or -c to list a local .deb file\n", c.Name())
		return command.SilentFailure()
	}
	if !*extract && !*vextract && !*contents {
		_, _ = fmt.Fprintf(stdio.Err, "%s: need an action option (-x, -X or -c); try '%s --help'\n", c.Name(), c.Name())
		return command.SilentFailure()
	}

	rest := fs.Args()
	if len(rest) == 0 {
		_, _ = fmt.Fprintf(stdio.Err, "%s: missing .deb archive operand\n", c.Name())
		return command.SilentFailure()
	}
	pkg, err := Open(rest[0])
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "%s: %v\n", c.Name(), err)
		return command.SilentFailure()
	}

	var runErr error
	switch {
	case *contents:
		runErr = listContents(stdio, pkg)
	case *extract || *vextract:
		runErr = doExtract(stdio, pkg, rest, *vextract)
	}
	if runErr != nil {
		_, _ = fmt.Fprintf(stdio.Err, "%s: %v\n", c.Name(), runErr)
		return command.SilentFailure()
	}
	return nil
}

// listContents prints the data tarball entries in an ls -l-like format.
func listContents(stdio command.IO, pkg *Package) error {
	entries, err := pkg.DataEntries()
	if err != nil {
		return err
	}
	for _, e := range entries {
		_, _ = fmt.Fprintf(stdio.Out, "%s %d/%d %10d %s\n",
			modeString(e.Type, e.Mode), e.UID, e.GID, e.Size, e.Name)
	}
	return nil
}

// doExtract extracts the data tarball into the requested directory (default ".").
func doExtract(stdio command.IO, pkg *Package, rest []string, verbose bool) error {
	dest := "."
	if len(rest) > 1 {
		dest = rest[1]
	}
	if verbose {
		entries, err := pkg.DataEntries()
		if err != nil {
			return err
		}
		for _, e := range entries {
			_, _ = fmt.Fprintf(stdio.Out, "%s\n", strings.TrimPrefix(e.Name, "./"))
		}
	}
	return pkg.Extract(dest)
}

// extractControl writes the control files into DIRECTORY (default "DEBIAN").
func extractControl(pkg *Package, rest []string) error {
	dest := "DEBIAN"
	if len(rest) > 1 {
		dest = rest[1]
	}
	names, err := pkg.ControlNames()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dest, 0o755); err != nil {
		return err
	}
	absDest, err := filepath.Abs(dest)
	if err != nil {
		return err
	}
	for _, name := range names {
		data, err := pkg.ControlFile(name)
		if err != nil {
			return err
		}
		target, err := safeJoin(absDest, name)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(target, data, 0o644); err != nil { //nolint:gosec // control files are world-readable by convention
			return err
		}
	}
	return nil
}

// printInfo prints the version line and the control file, like dpkg-deb -I.
func printInfo(stdio command.IO, pkg *Package) error {
	_, _ = fmt.Fprintf(stdio.Out, " new Debian package, version %s\n", strings.TrimSpace(pkg.Version))
	control, err := pkg.ControlFile("control")
	if err == nil {
		_, _ = fmt.Fprintf(stdio.Out, "\n%s", string(control))
	}
	return nil
}

// printField prints the requested control fields (one per FIELD operand). With
// no FIELD operands it prints the whole control file.
func printField(stdio command.IO, pkg *Package, fields []string) error {
	control, err := pkg.ControlFile("control")
	if err != nil {
		return err
	}
	if len(fields) == 0 {
		_, _ = fmt.Fprint(stdio.Out, string(control))
		return nil
	}
	values := parseControl(string(control))
	for _, f := range fields {
		f = strings.TrimSpace(f)
		if f == "" {
			continue
		}
		if v, ok := values[strings.ToLower(f)]; ok {
			_, _ = fmt.Fprintln(stdio.Out, v)
		}
	}
	return nil
}

// parseControl parses a Debian control file's top-level fields into a map keyed
// by the lower-cased field name.
func parseControl(s string) map[string]string {
	out := map[string]string{}
	for _, line := range strings.Split(s, "\n") {
		if line == "" || line[0] == ' ' || line[0] == '\t' {
			continue
		}
		idx := strings.IndexByte(line, ':')
		if idx < 0 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(line[:idx]))
		out[key] = strings.TrimSpace(line[idx+1:])
	}
	return out
}

// modeString renders a tar type+mode as a 10-character ls-style string.
func modeString(typ byte, mode int64) string {
	var b [10]byte
	switch typ {
	case '5':
		b[0] = 'd'
	case '2':
		b[0] = 'l'
	default:
		b[0] = '-'
	}
	rwx := []byte("rwxrwxrwx")
	for i := 0; i < 9; i++ {
		if mode&(1<<uint(8-i)) != 0 {
			b[i+1] = rwx[i]
		} else {
			b[i+1] = '-'
		}
	}
	return string(b[:])
}

func (c *Command) helpDpkgDeb() command.Help {
	return command.Help{
		Description: "Inspect or unpack a local Debian binary package (.deb) without touching the system package database. Exactly one action option selects the operation. This first slice is read-only: it cannot build packages and has no install/database support.",
		Examples: []command.Example{
			{Command: "dpkg-deb -c pkg.deb", Explain: "List the files the package would install."},
			{Command: "dpkg-deb -x pkg.deb out/", Explain: "Extract the data tarball under 'out/'."},
			{Command: "dpkg-deb -e pkg.deb DEBIAN", Explain: "Extract the control files into 'DEBIAN'."},
			{Command: "dpkg-deb -I pkg.deb", Explain: "Print the package's control information."},
			{Command: "dpkg-deb -f pkg.deb Package Version", Explain: "Print selected control fields."},
		},
		ExitStatus: "0  success.\n1  the archive could not be read or an action failed.",
		Notes: []string{
			"Supported tarball compression: plain tar, gzip, xz, lzma and bzip2. zstd is not yet supported and fails with a clear error.",
			"Extraction is path-safe: entries that would escape the destination directory are rejected.",
			"Package build (-b) and the dpkg database are intentionally out of scope for this first slice.",
		},
	}
}

func (c *Command) helpDpkg() command.Help {
	return command.Help{
		Description: "Operate on local Debian binary packages (.deb). This first slice supports only the file-level rescue/recovery workflows: -x and -X unpack the data tarball into DIRECTORY, and -c lists its contents. Package-database actions (-i/--install, -r/--remove, -l/--list, --configure) are intentionally unsupported and fail with a clear error.",
		Examples: []command.Example{
			{Command: "dpkg -x pkg.deb out/", Explain: "Unpack the package's files under 'out/'."},
			{Command: "dpkg -X pkg.deb out/", Explain: "Unpack and list each extracted file."},
			{Command: "dpkg -c pkg.deb", Explain: "List the files inside the package."},
		},
		ExitStatus: "0  success.\n1  the archive could not be read, or an unsupported operation was requested.",
		Notes: []string{
			"Extraction is path-safe: archive entries that would escape DIRECTORY are rejected.",
			"No package database is read or written; install/remove/list/configure are out of scope.",
		},
	}
}
