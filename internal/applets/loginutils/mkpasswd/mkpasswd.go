// Package mkpasswd implements the mkpasswd applet: compute the crypt(3) hash of
// a password, as stored in /etc/shadow.
package mkpasswd

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

// Command is the mkpasswd applet.
type Command struct{}

// New returns a mkpasswd command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "mkpasswd" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Compute the crypt hash of a password" }

// method describes a supported hashing method and its crypt magic prefix.
type method struct {
	crypt crypt.Crypt
	magic string
}

var methods = map[string]method{
	"sha-512": {crypt.SHA512, "$6$"},
	"sha512":  {crypt.SHA512, "$6$"},
	"sha-256": {crypt.SHA256, "$5$"},
	"sha256":  {crypt.SHA256, "$5$"},
	"md5":     {crypt.MD5, "$1$"},
}

// Run executes mkpasswd.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-m METHOD] [-S SALT] [PASSWORD]", stdio.Err).WithHelp(command.Help{
		Description: "Print the crypt(3) hash of PASSWORD, in the format stored in /etc/shadow. -m " +
			"selects the method: sha-512 (the default), sha-256, or md5. -S sets the salt; without it " +
			"a random salt is used. If PASSWORD is omitted it is read from standard input.",
		Examples: []command.Example{
			{Command: "mkpasswd -m sha-512 -S abcdefgh secret", Explain: "Hash 'secret' with a fixed salt."},
			{Command: "echo secret | mkpasswd", Explain: "Hash a password read from stdin."},
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

// readPassword returns the password from the first operand, or the first line
// of standard input.
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
