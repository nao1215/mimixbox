// Package cryptpw implements the cryptpw applet: print the crypt(3) hash of a
// password, defaulting to reading the password from standard input.
package cryptpw

import (
	"bufio"
	"context"
	"fmt"
	"strings"

	"github.com/GehirnInc/crypt"
	_ "github.com/GehirnInc/crypt/md5_crypt"    // register $1$
	_ "github.com/GehirnInc/crypt/sha256_crypt" // register $5$
	_ "github.com/GehirnInc/crypt/sha512_crypt" // register $6$

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the cryptpw applet.
type Command struct{}

// New returns a cryptpw command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "cryptpw" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Crypt-hash a password from stdin" }

// methods maps the supported algorithm names to their crypt type and magic.
var methods = map[string]struct {
	crypt crypt.Crypt
	magic string
}{
	"sha-512": {crypt.SHA512, "$6$"}, "sha512": {crypt.SHA512, "$6$"},
	"sha-256": {crypt.SHA256, "$5$"}, "sha256": {crypt.SHA256, "$5$"},
	"md5": {crypt.MD5, "$1$"},
}

// Run executes cryptpw.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-m METHOD] [-S SALT] [PASSWORD]", stdio.Err).WithHelp(command.Help{
		Description: "Print the crypt(3) hash of a password. The password is taken from PASSWORD if " +
			"given, otherwise from standard input. -m selects the method (sha-512 by default, also " +
			"sha-256 or md5); -S sets the salt, otherwise a random one is used. This is the same as " +
			"mkpasswd, with the password read from stdin by default.",
		Examples: []command.Example{
			{Command: "echo secret | cryptpw", Explain: "Hash a password read from stdin."},
			{Command: "cryptpw -m md5 -S abcdefgh secret", Explain: "MD5-hash with a fixed salt."},
		},
		ExitStatus: "0  the hash was printed.\n1  an unknown method or a hashing error.",
	})
	methodName := fs.StringP("method", "m", "sha-512", "hashing method: sha-512, sha-256, or md5")
	salt := fs.StringP("salt", "S", "", "salt to use (random if unset)")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	m, ok := methods[strings.ToLower(*methodName)]
	if !ok {
		return command.Failuref("unknown method: %q (use sha-512, sha-256, or md5)", *methodName)
	}

	password, err := readPassword(stdio, fs.Args())
	if err != nil {
		return command.Failuref("%v", err)
	}

	var saltArg []byte
	if *salt != "" {
		saltArg = []byte(m.magic + *salt)
	}
	hash, err := crypt.New(m.crypt).Generate([]byte(password), saltArg)
	if err != nil {
		return command.Failuref("cannot hash the password: %v", err)
	}
	_, _ = fmt.Fprintln(stdio.Out, hash)
	return nil
}

// readPassword returns the password from the first operand, or the first line of
// standard input.
func readPassword(stdio command.IO, operands []string) (string, error) {
	if len(operands) > 0 {
		return operands[0], nil
	}
	sc := bufio.NewScanner(stdio.In)
	if !sc.Scan() {
		if err := sc.Err(); err != nil {
			return "", err
		}
		return "", fmt.Errorf("no password given")
	}
	return sc.Text(), nil
}
