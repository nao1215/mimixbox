// Package grepalias implements the egrep and fgrep compatibility applets as thin
// wrappers over the grep applet: egrep forces -E (extended regular expressions)
// and fgrep forces -F (fixed strings), so neither can drift from grep's
// behavior.
package grepalias

import (
	"context"

	"github.com/nao1215/mimixbox/internal/applets/findutils/grep"
	"github.com/nao1215/mimixbox/internal/command"
)

// alias is the shared metadata that distinguishes egrep from fgrep: the grep
// mode flag each one forces, plus the words used in their --help and synopsis.
// Holding both aliases in one table keeps the egrep/fgrep behavior from drifting
// apart and lets the constructors and Run share a single description.
type alias struct {
	flag     string // grep mode flag forced for this alias ("-E" or "-F")
	mode     string // human-readable mode phrase used in --help
	synopsis string // one-line applet-list description
}

// aliases maps each alias name to its grep mode metadata.
var aliases = map[string]alias{
	"egrep": {flag: "-E", mode: "extended (-E)", synopsis: "Search with extended regular expressions (grep -E)"},
	"fgrep": {flag: "-F", mode: "fixed-string (-F)", synopsis: "Search for fixed strings (grep -F)"},
}

// Command is the egrep or fgrep alias.
type Command struct {
	name string
}

// NewEgrep returns the egrep alias (grep -E).
func NewEgrep() *Command { return &Command{name: "egrep"} }

// NewFgrep returns the fgrep alias (grep -F).
func NewFgrep() *Command { return &Command{name: "fgrep"} }

// Name returns the command name.
func (c *Command) Name() string { return c.name }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return aliases[c.name].synopsis }

// Run delegates to grep with the alias's mode flag prepended. A leading --help
// renders alias-named structured help so the usage and example lines match the
// invoked command rather than grep's.
func (c *Command) Run(ctx context.Context, stdio command.IO, args []string) error {
	a := aliases[c.name]
	if command.HandleHelpVersionWith(stdio, c.Name(), "[OPTION]... PATTERN [FILE]...", command.Help{
		Description: "Search each FILE (or standard input) for lines matching PATTERN. " + c.Name() +
			" is grep in " + a.mode + " mode.",
		Examples: []command.Example{
			{Command: c.Name() + " 'foo|bar' file.txt", Explain: "Print lines of file.txt that match the pattern."},
			{Command: "ls | " + c.Name() + " -i readme", Explain: "Filter the listing case-insensitively."},
		},
		ExitStatus: "0  a line matched.\n1  no lines matched.\n2  an error occurred.",
	}, args) {
		return nil
	}
	return grep.New().Run(ctx, stdio, append([]string{a.flag}, args...))
}
