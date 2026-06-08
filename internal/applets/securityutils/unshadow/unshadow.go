// Package unshadow implements the unshadow applet: combine the account fields
// of an /etc/passwd file with the hashes from an /etc/shadow file into a single
// passwd-format stream, the standard first step of an authorized local password
// audit. It is a clean-room implementation written from the documented file
// formats; it copies no John the Ripper (GPL) source.
package unshadow

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the unshadow applet.
type Command struct{}

// New returns an unshadow command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "unshadow" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Combine passwd and shadow files for password auditing" }

// Run executes unshadow.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "PASSWD SHADOW", stdio.Err)

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	rest := fs.Args()
	if len(rest) != 2 {
		return command.Failuref("two file operands are required: PASSWD SHADOW")
	}

	passwd, err := command.Open(stdio, rest[0])
	if err != nil {
		return command.Failuref("%s", command.FileError(rest[0], err))
	}
	defer func() { _ = passwd.Close() }()

	shadow, err := command.Open(stdio, rest[1])
	if err != nil {
		return command.Failuref("%s", command.FileError(rest[1], err))
	}
	defer func() { _ = shadow.Close() }()

	hashes, err := parseShadow(shadow)
	if err != nil {
		return command.Failure(err)
	}
	return c.merge(stdio, passwd, hashes)
}

// parseShadow reads a shadow file and returns a map from user name to password
// hash (fields 1 and 2 of each colon-separated line).
func parseShadow(r io.Reader) (map[string]string, error) {
	hashes := make(map[string]string)
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := sc.Text()
		if line == "" {
			continue
		}
		fields := strings.SplitN(line, ":", 3)
		if len(fields) >= 2 {
			hashes[fields[0]] = fields[1]
		}
	}
	return hashes, sc.Err()
}

// merge writes each passwd line with its password field (field 2) replaced by
// the matching shadow hash, like John the Ripper's unshadow.
func (c *Command) merge(stdio command.IO, passwd io.Reader, hashes map[string]string) error {
	sc := bufio.NewScanner(passwd)
	for sc.Scan() {
		line := sc.Text()
		if line == "" {
			continue
		}
		fields := strings.Split(line, ":")
		if len(fields) < 2 {
			continue
		}
		if hash, ok := hashes[fields[0]]; ok {
			fields[1] = hash
		}
		if _, err := fmt.Fprintln(stdio.Out, strings.Join(fields, ":")); err != nil {
			return command.Failure(err)
		}
	}
	return sc.Err()
}
