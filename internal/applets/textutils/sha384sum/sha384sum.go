// Package sha384sum implements the sha384sum applet: print or check SHA-384
// message digests, in the GNU coreutils output format.
package sha384sum

import (
	"context"
	"crypto/sha512"

	"github.com/nao1215/mimixbox/internal/command"
	"github.com/nao1215/mimixbox/internal/hashsum"
)

const synopsis = "Calculate or Check secure hash 384 algorithm"

// Command is the sha384sum applet.
type Command struct{ inner *hashsum.Command }

// New returns a sha384sum command.
func New() *Command {
	return &Command{inner: hashsum.New("sha384sum", synopsis, sha512.New384)}
}

// Name returns the command name.
func (c *Command) Name() string { return c.inner.Name() }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return c.inner.Synopsis() }

// Run executes sha384sum.
func (c *Command) Run(ctx context.Context, stdio command.IO, args []string) error {
	return c.inner.Run(ctx, stdio, args)
}
