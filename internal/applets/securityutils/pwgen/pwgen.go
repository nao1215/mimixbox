// Package pwgen implements the pwgen applet: generate random candidate
// passwords for authorized security testing and password-strength auditing. It
// is a clean-room port of morrigan's pwlist.
package pwgen

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the pwgen applet.
type Command struct{}

// New returns a pwgen command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "pwgen" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Generate random passwords for authorized testing" }

// Character classes the generator can draw from.
const (
	lowers  = "abcdefghijklmnopqrstuvwxyz"
	uppers  = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digits  = "0123456789"
	symbols = "!@#$%^&*()-_=+[]{}"
)

// Run executes pwgen.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]...", stdio.Err).WithHelp(command.Help{
		Description: "Generate cryptographically random passwords for authorized testing and " +
			"password-strength auditing, printing one password per line.",
		Examples: []command.Example{
			{Command: "pwgen", Explain: "Generate one 16-character password."},
			{Command: "pwgen -l 24 -n 5", Explain: "Generate five 24-character passwords."},
			{Command: "pwgen -s -o out.txt", Explain: "Include symbols and write the result to out.txt."},
		},
		ExitStatus: "0  the passwords were generated.\n1  an invalid option value was given or output failed.",
	})
	length := fs.IntP("length", "l", 16, "length of each password")
	count := fs.IntP("number", "n", 1, "how many passwords to generate")
	withSymbols := fs.BoolP("symbols", "s", false, "include symbol characters")
	noDigits := fs.BoolP("no-numerals", "0", false, "exclude digits")
	outFile := fs.StringP("output", "o", "", "write to FILE instead of stdout")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	if *length < 1 {
		return command.Failuref("length must be at least 1")
	}
	if *count < 1 {
		return command.Failuref("number must be at least 1")
	}

	charset := buildCharset(*withSymbols, !*noDigits)

	var b strings.Builder
	for i := 0; i < *count; i++ {
		pw, err := generate(*length, charset)
		if err != nil {
			return command.Failuref("%v", err)
		}
		b.WriteString(pw)
		b.WriteByte('\n')
	}

	if *outFile != "" {
		if err := os.WriteFile(*outFile, []byte(b.String()), 0o600); err != nil {
			return command.Failuref("cannot write %q: %v", *outFile, err)
		}
		return nil
	}
	if _, err := fmt.Fprint(stdio.Out, b.String()); err != nil {
		return command.Failure(err)
	}
	return nil
}

// buildCharset assembles the alphabet from the enabled character classes.
func buildCharset(withSymbols, withDigits bool) string {
	cs := lowers + uppers
	if withDigits {
		cs += digits
	}
	if withSymbols {
		cs += symbols
	}
	return cs
}

// generate returns a cryptographically random password of the given length
// drawn uniformly from charset.
func generate(length int, charset string) (string, error) {
	if charset == "" {
		return "", fmt.Errorf("empty character set")
	}
	out := make([]byte, length)
	max := big.NewInt(int64(len(charset)))
	for i := range out {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		out[i] = charset[n.Int64()]
	}
	return string(out), nil
}
