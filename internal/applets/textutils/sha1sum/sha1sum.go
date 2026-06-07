// Package sha1sum implements the sha1sum applet: print or check SHA-1 message
// digests, in the GNU coreutils output format.
package sha1sum

import (
	"context"
	"crypto/sha1" //nolint:gosec // sha1sum is by definition a SHA-1 utility

	"github.com/nao1215/mimixbox/internal/command"
	"github.com/nao1215/mimixbox/internal/hashsum"
)

// synopsis is the one-line description shown in the applet list. It is kept
// byte-for-byte identical to the legacy table (including its typo) so the
// listing does not change.
const synopsis = "Calculate or Check secure hash 1 algorithm"

// Command is the sha1sum applet.
type Command struct{ inner *hashsum.Command }

// New returns a sha1sum command.
func New() *Command {
	return &Command{inner: hashsum.New("sha1sum", synopsis, sha1.New)}
}

// Name returns the command name.
func (c *Command) Name() string { return c.inner.Name() }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return c.inner.Synopsis() }

// Run executes sha1sum.
func (c *Command) Run(ctx context.Context, stdio command.IO, args []string) error {
	return c.inner.Run(ctx, stdio, args)
}
