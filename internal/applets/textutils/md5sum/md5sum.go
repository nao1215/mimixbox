// Package md5sum implements the md5sum applet: print or check MD5 message
// digests, in the GNU coreutils output format.
package md5sum

import (
	"crypto/md5" //nolint:gosec // md5sum is by definition an MD5 utility

	"github.com/nao1215/mimixbox/internal/hashsum"
)

// synopsis is the one-line description shown in the applet list. It is kept
// byte-for-byte identical to the legacy table so the listing does not change.
const synopsis = "Calculate or Check md5sum message digest"

// Command is the md5sum applet. It embeds the shared hashsum backend, which
// supplies the Name, Synopsis and Run methods; only the hash differs.
type Command struct{ *hashsum.Command }

// New returns an md5sum command.
func New() *Command {
	return &Command{Command: hashsum.New("md5sum", synopsis, md5.New)}
}
