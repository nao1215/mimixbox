// Package sha256sum implements the sha256sum applet: print or check SHA-256
// message digests, in the GNU coreutils output format.
package sha256sum

import (
	"crypto/sha256"

	"github.com/nao1215/mimixbox/internal/hashsum"
)

// synopsis is the one-line description shown in the applet list. It is kept
// byte-for-byte identical to the legacy table (including its typo) so the
// listing does not change.
const synopsis = "Calculate or Check secure hash 256 algorithm"

// Command is the sha256sum applet. It embeds the shared hashsum backend, which
// supplies the Name, Synopsis and Run methods; only the hash differs.
type Command struct{ *hashsum.Command }

// New returns a sha256sum command.
func New() *Command {
	return &Command{Command: hashsum.New("sha256sum", synopsis, sha256.New)}
}
