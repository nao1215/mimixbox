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

// Command is the egrep or fgrep alias.
type Command struct {
	name string
	flag string
}

// NewEgrep returns the egrep alias (grep -E).
func NewEgrep() *Command { return &Command{name: "egrep", flag: "-E"} }

// NewFgrep returns the fgrep alias (grep -F).
func NewFgrep() *Command { return &Command{name: "fgrep", flag: "-F"} }

// Name returns the command name.
func (c *Command) Name() string { return c.name }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string {
	if c.flag == "-F" {
		return "Search for fixed strings (grep -F)"
	}
	return "Search with extended regular expressions (grep -E)"
}

// Run delegates to grep with the alias's mode flag prepended.
func (c *Command) Run(ctx context.Context, stdio command.IO, args []string) error {
	return grep.New().Run(ctx, stdio, append([]string{c.flag}, args...))
}
