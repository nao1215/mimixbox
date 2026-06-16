// Package sha1sum implements the sha1sum applet: print or check SHA-1 message
// digests, in the GNU coreutils output format.
package sha1sum

import (
	"crypto/sha1" //nolint:gosec // sha1sum is by definition a SHA-1 utility

	"github.com/nao1215/mimixbox/internal/hashsum"
)

// synopsis is the one-line description shown in the applet list. It is kept
// byte-for-byte identical to the legacy table (including its typo) so the
// listing does not change.
const synopsis = "Calculate or Check secure hash 1 algorithm"

// Command is the sha1sum applet. It embeds the shared hashsum backend, which
// supplies the Name, Synopsis and Run methods; only the hash differs.
type Command struct{ *hashsum.Command }

// New returns a sha1sum command.
func New() *Command {
	return &Command{Command: hashsum.New("sha1sum", synopsis, sha1.New)}
}
