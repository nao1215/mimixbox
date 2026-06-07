// Package md5sum implements the md5sum applet: print or check MD5 message
// digests, in the GNU coreutils output format.
package md5sum

import (
	"context"
	"crypto/md5" //nolint:gosec // md5sum is by definition an MD5 utility

	"github.com/nao1215/mimixbox/internal/command"
	"github.com/nao1215/mimixbox/internal/hashsum"
)

// synopsis is the one-line description shown in the applet list. It is kept
// byte-for-byte identical to the legacy table so the listing does not change.
const synopsis = "Calculate or Check md5sum message digest"

// Command is the md5sum applet.
type Command struct{ inner *hashsum.Command }

// New returns an md5sum command.
func New() *Command {
	return &Command{inner: hashsum.New("md5sum", synopsis, md5.New)}
}

// Name returns the command name.
func (c *Command) Name() string { return c.inner.Name() }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return c.inner.Synopsis() }

// Run executes md5sum.
func (c *Command) Run(ctx context.Context, stdio command.IO, args []string) error {
	return c.inner.Run(ctx, stdio, args)
}
