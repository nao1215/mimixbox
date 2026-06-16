// Package sha384sum implements the sha384sum applet: print or check SHA-384
// message digests, in the GNU coreutils output format.
package sha384sum

import (
	"crypto/sha512"

	"github.com/nao1215/mimixbox/internal/hashsum"
)

const synopsis = "Calculate or Check secure hash 384 algorithm"

// Command is the sha384sum applet. It embeds the shared hashsum backend, which
// supplies the Name, Synopsis and Run methods; only the hash differs.
type Command struct{ *hashsum.Command }

// New returns a sha384sum command.
func New() *Command {
	return &Command{Command: hashsum.New("sha384sum", synopsis, sha512.New384)}
}
