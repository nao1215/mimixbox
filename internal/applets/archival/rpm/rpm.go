// Package rpm implements a tiny read-only subset of the rpm applet: querying a
// package file with -qp. It can print the package identity (name-version-
// release.arch), an information summary (-i) and the file list (-l). It does
// not install, remove or query the system database.
package rpm

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/nao1215/mimixbox/internal/applets/archival/rpmfile"
	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the rpm applet.
type Command struct{}

// New returns an rpm command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "rpm" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Query an RPM package file" }

// Run executes rpm.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "-qp [-i] [-l] FILE.rpm", stdio.Err).WithHelp(command.Help{
		Description: "Query an RPM package file with -qp. By default the package identity is printed " +
			"as name-version-release.arch; -i prints an information summary and -l lists the files " +
			"contained in the package. Only package-file queries are supported: rpm does not " +
			"install, remove, or query the system RPM database.",
		Examples: []command.Example{
			{Command: "rpm -qp pkg.rpm", Explain: "Print the package's name-version-release.arch."},
			{Command: "rpm -qpi pkg.rpm", Explain: "Print a summary of the package information."},
			{Command: "rpm -qpl pkg.rpm", Explain: "List the files contained in the package."},
		},
		ExitStatus: "0  the package file was queried successfully.\n1  the file could not be read or the query mode was not -qp.",
	})
	query := fs.BoolP("query", "q", false, "query mode")
	pkgFile := fs.BoolP("package", "p", false, "query a package file")
	info := fs.BoolP("info", "i", false, "display package information")
	list := fs.BoolP("list", "l", false, "list files in the package")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	if !*query || !*pkgFile {
		_, _ = fmt.Fprintln(stdio.Err, "rpm: only package-file queries are supported (use -qp FILE.rpm)")
		return command.SilentFailure()
	}

	names := fs.Args()
	if len(names) == 0 {
		_, _ = fmt.Fprintln(stdio.Err, "rpm: no package file given")
		return command.SilentFailure()
	}

	var failed bool
	for _, name := range names {
		if err := c.queryFile(stdio, name, *info, *list); err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "rpm: %v\n", err)
			failed = true
		}
	}
	if failed {
		return command.SilentFailure()
	}
	return nil
}

// queryFile prints the requested view of one package file.
func (c *Command) queryFile(stdio command.IO, name string, info, list bool) error {
	f, err := os.Open(name) //nolint:gosec // user-named file
	if err != nil {
		return fmt.Errorf("%s", command.FileError(name, err))
	}
	defer func() { _ = f.Close() }()

	rpm, err := rpmfile.Open(f)
	if err != nil {
		return err
	}
	h := rpm.Header

	switch {
	case list:
		for _, p := range fileList(h) {
			_, _ = fmt.Fprintln(stdio.Out, p)
		}
	case info:
		_, _ = fmt.Fprintf(stdio.Out, "Name        : %s\n", h.String(rpmfile.TagName))
		_, _ = fmt.Fprintf(stdio.Out, "Version     : %s\n", h.String(rpmfile.TagVersion))
		_, _ = fmt.Fprintf(stdio.Out, "Release     : %s\n", h.String(rpmfile.TagRelease))
		_, _ = fmt.Fprintf(stdio.Out, "Architecture: %s\n", h.String(rpmfile.TagArch))
		_, _ = fmt.Fprintf(stdio.Out, "Summary     : %s\n", h.String(rpmfile.TagSummary))
	default:
		_, _ = fmt.Fprintln(stdio.Out, nevra(h))
	}
	return nil
}

// nevra builds the "name-version-release.arch" identity string.
func nevra(h *rpmfile.Header) string {
	s := fmt.Sprintf("%s-%s-%s", h.String(rpmfile.TagName), h.String(rpmfile.TagVersion), h.String(rpmfile.TagRelease))
	if arch := h.String(rpmfile.TagArch); arch != "" {
		s += "." + arch
	}
	return s
}

// fileList reconstructs the absolute file paths from the BASENAMES, DIRNAMES and
// DIRINDEXES tags.
func fileList(h *rpmfile.Header) []string {
	base := h.StringArray(rpmfile.TagBasenames)
	dirs := h.StringArray(rpmfile.TagDirnames)
	idx := h.Int32Array(rpmfile.TagDirindexes)
	var out []string
	for i, b := range base {
		dir := ""
		if i < len(idx) && idx[i] >= 0 && int(idx[i]) < len(dirs) {
			dir = dirs[idx[i]]
		}
		out = append(out, path.Clean(dir+b))
	}
	return out
}
