// Package sha512sum implements the sha512sum applet: print or check SHA-512
// message digests, in the GNU coreutils output format.
package sha512sum

import (
	"context"
	"crypto/sha512"

	"github.com/nao1215/mimixbox/internal/command"
	"github.com/nao1215/mimixbox/internal/hashsum"
)

// synopsis is the one-line description shown in the applet list. It is kept
// byte-for-byte identical to the legacy table (including its typo) so the
// listing does not change.
const synopsis = "Calculate or Check secure hash 512 algorithm"

// Command is the sha512sum applet.
type Command struct{ inner *hashsum.Command }

// New returns a sha512sum command.
func New() *Command {
	return &Command{inner: hashsum.New("sha512sum", synopsis, sha512.New)}
}

// Name returns the command name.
func (c *Command) Name() string { return c.inner.Name() }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return c.inner.Synopsis() }

// Run executes sha512sum.
func (c *Command) Run(ctx context.Context, stdio command.IO, args []string) error {
	return c.inner.Run(ctx, stdio, args)
}
