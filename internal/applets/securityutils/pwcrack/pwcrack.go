// Package pwcrack implements the pwcrack applet: audit crypt(3) password hashes
// (as found in /etc/shadow) against a wordlist, the way a sysadmin checks for
// weak passwords.
//
// This is a clean-room implementation written from the documented crypt hash
// formats; it copies no John the Ripper (GPL) source. The actual hashing uses
// github.com/GehirnInc/crypt, a permissively (BSD-2-Clause) licensed library,
// as the issue recommends.
package pwcrack

import (
	"bufio"
	"context"
	"fmt"
	"strings"

	"github.com/GehirnInc/crypt"
	_ "github.com/GehirnInc/crypt/apr1_crypt"   // register $apr1$
	_ "github.com/GehirnInc/crypt/md5_crypt"    // register $1$
	_ "github.com/GehirnInc/crypt/sha256_crypt" // register $5$
	_ "github.com/GehirnInc/crypt/sha512_crypt" // register $6$

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the pwcrack applet.
type Command struct{}

// New returns a pwcrack command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "pwcrack" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Audit crypt(3) password hashes against a wordlist" }

// target is one hash to crack, optionally labelled with a user name.
type target struct {
	user string
	hash string
}

// Run executes pwcrack.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "-w WORDLIST [HASH | --shadow FILE]", stdio.Err).WithHelp(command.Help{
		Description: "Audit crypt(3) password hashes against a wordlist. The hash is given as an operand, " +
			"read from a passwd/shadow-format file with --shadow, or read from standard input. " +
			"Use only on systems you are authorized to test.",
		Examples: []command.Example{
			{Command: "pwcrack -w words.txt '$6$salt$hash'", Explain: "Crack a single hash using words.txt."},
			{Command: "pwcrack -w words.txt --shadow /etc/shadow", Explain: "Crack every entry of a shadow file."},
			{Command: "echo '$1$salt$hash' | pwcrack -w words.txt", Explain: "Read the hash to crack from stdin."},
		},
		ExitStatus: "0  at least one hash was cracked.\n1  no hash was cracked, or an error occurred.",
	})
	wordlist := fs.StringP("wordlist", "w", "", "file of candidate passwords, one per line")
	shadow := fs.StringP("shadow", "s", "", "crack every entry of a passwd/shadow-format file")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	if *wordlist == "" {
		return command.Failuref("a wordlist is required (-w)")
	}

	targets, err := c.targets(stdio, *shadow, fs.Args())
	if err != nil {
		return err
	}
	if len(targets) == 0 {
		return command.Failuref("no hash to crack (give a HASH or --shadow FILE)")
	}

	words, err := c.loadWords(stdio, *wordlist)
	if err != nil {
		return err
	}

	cracked := 0
	for _, t := range targets {
		if word, ok := crack(t.hash, words); ok {
			cracked++
			if t.user != "" {
				_, _ = fmt.Fprintf(stdio.Out, "%s: %s\n", t.user, word)
			} else {
				_, _ = fmt.Fprintf(stdio.Out, "%s: %s\n", t.hash, word)
			}
		} else if t.user != "" {
			_, _ = fmt.Fprintf(stdio.Out, "%s: (not found)\n", t.user)
		} else {
			_, _ = fmt.Fprintf(stdio.Out, "%s: (not found)\n", t.hash)
		}
	}
	if cracked == 0 {
		return &command.ExitError{Code: command.ExitFailure}
	}
	return nil
}

// targets resolves the hashes to crack: every crackable entry of the shadow
// file, or the single HASH operand, or the hash read from standard input.
func (c *Command) targets(stdio command.IO, shadow string, operands []string) ([]target, error) {
	if shadow != "" {
		r, err := command.Open(stdio, shadow)
		if err != nil {
			return nil, command.Failuref("%s", command.FileError(shadow, err))
		}
		defer func() { _ = r.Close() }()
		var ts []target
		sc := bufio.NewScanner(r)
		for sc.Scan() {
			fields := strings.Split(sc.Text(), ":")
			if len(fields) >= 2 && crypt.IsHashSupported(fields[1]) {
				ts = append(ts, target{user: fields[0], hash: fields[1]})
			}
		}
		return ts, sc.Err()
	}

	if len(operands) > 0 {
		return []target{{hash: operands[0]}}, nil
	}

	sc := bufio.NewScanner(stdio.In)
	if sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line != "" {
			return []target{{hash: line}}, nil
		}
	}
	return nil, sc.Err()
}

// loadWords reads the wordlist into a slice.
func (c *Command) loadWords(stdio command.IO, path string) ([]string, error) {
	r, err := command.Open(stdio, path)
	if err != nil {
		return nil, command.Failuref("%s", command.FileError(path, err))
	}
	defer func() { _ = r.Close() }()

	var words []string
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		words = append(words, sc.Text())
	}
	return words, sc.Err()
}

// crack returns the first word whose crypt hash matches hash.
func crack(hash string, words []string) (string, bool) {
	if !crypt.IsHashSupported(hash) {
		return "", false
	}
	crypter := crypt.NewFromHash(hash)
	for _, w := range words {
		if crypter.Verify(hash, []byte(w)) == nil {
			return w, true
		}
	}
	return "", false
}
