// Package sha3sum implements the sha3sum applet: print or check SHA-3 message
// digests in the GNU coreutils output format. The digest length is selected with
// -a (224, 256, 384, or 512; default 256).
package sha3sum

import (
	"context"
	"hash"

	"github.com/nao1215/mimixbox/internal/command"
	"github.com/nao1215/mimixbox/internal/hashsum"
	"golang.org/x/crypto/sha3"
)

const synopsis = "Calculate or Check SHA-3 message digest"

// Command is the sha3sum applet.
type Command struct{}

// New returns a sha3sum command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "sha3sum" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return synopsis }

// Run executes sha3sum: it selects the digest length from -a, then delegates the
// file handling and GNU-format output to the shared hashsum framework.
func (c *Command) Run(ctx context.Context, stdio command.IO, args []string) error {
	bits, rest, err := extractAlgo(args)
	if err != nil {
		return command.Failuref("%v", err)
	}
	var newHash func() hash.Hash
	switch bits {
	case 224:
		newHash = sha3.New224
	case 256:
		newHash = sha3.New256
	case 384:
		newHash = sha3.New384
	case 512:
		newHash = sha3.New512
	default:
		return command.Failuref("unsupported digest length %d (use 224, 256, 384, or 512)", bits)
	}
	return hashsum.New("sha3sum", synopsis, newHash).Run(ctx, stdio, rest)
}

// extractAlgo pulls the -a/--algorithm option (default 256) out of args so the
// remaining arguments can be handled by the standard sum flag set.
func extractAlgo(args []string) (bits int, rest []string, err error) {
	bits = 256
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "-a" || a == "--algorithm":
			if i+1 >= len(args) {
				return 0, nil, command.Failuref("option %s requires an argument", a)
			}
			i++
			if bits, err = atoi(args[i]); err != nil {
				return 0, nil, err
			}
		case len(a) > 2 && a[:2] == "-a":
			if bits, err = atoi(a[2:]); err != nil {
				return 0, nil, err
			}
		case len(a) > len("--algorithm=") && a[:len("--algorithm=")] == "--algorithm=":
			if bits, err = atoi(a[len("--algorithm="):]); err != nil {
				return 0, nil, err
			}
		default:
			rest = append(rest, a)
		}
	}
	return bits, rest, nil
}

func atoi(s string) (int, error) {
	n := 0
	if s == "" {
		return 0, command.Failuref("invalid digest length %q", s)
	}
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return 0, command.Failuref("invalid digest length %q", s)
		}
		n = n*10 + int(ch-'0')
	}
	return n, nil
}
