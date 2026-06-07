// Package sha256sum implements the sha256sum applet: print or check SHA-256
// message digests, in the GNU coreutils output format.
package sha256sum

import (
	"context"
	"crypto/sha256"

	"github.com/nao1215/mimixbox/internal/command"
	"github.com/nao1215/mimixbox/internal/hashsum"
)

// synopsis is the one-line description shown in the applet list. It is kept
// byte-for-byte identical to the legacy table (including its typo) so the
// listing does not change.
const synopsis = "alculate or Check sercure hash 256 algorithm"

// Command is the sha256sum applet.
type Command struct{ inner *hashsum.Command }

// New returns a sha256sum command.
func New() *Command {
	return &Command{inner: hashsum.New("sha256sum", synopsis, sha256.New)}
}

// Name returns the command name.
func (c *Command) Name() string { return c.inner.Name() }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return c.inner.Synopsis() }

// Run executes sha256sum.
func (c *Command) Run(ctx context.Context, stdio command.IO, args []string) error {
	return c.inner.Run(ctx, stdio, args)
}
