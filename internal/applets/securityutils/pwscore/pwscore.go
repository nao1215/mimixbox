// Package pwscore implements the pwscore applet: estimate the strength of a
// password and explain the reasoning. It is a clean-room implementation written
// from first principles (length, character-class diversity, common-password
// check); it copies no GPL or libpwquality source.
package pwscore

import (
	"bufio"
	"context"
	"fmt"
	"strings"
	"unicode"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the pwscore applet.
type Command struct{}

// New returns a pwscore command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "pwscore" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Estimate the strength of a password" }

// Run executes pwscore.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[PASSWORD]", stdio.Err)

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	var password string
	if rest := fs.Args(); len(rest) > 0 {
		password = rest[0]
	} else {
		sc := bufio.NewScanner(stdio.In)
		if sc.Scan() {
			password = sc.Text()
		}
		if err := sc.Err(); err != nil {
			return command.Failure(err)
		}
	}
	if password == "" {
		return command.Failuref("no password provided")
	}

	score, reasons := Score(password)
	_, _ = fmt.Fprintf(stdio.Out, "Score: %d/100 (%s)\n", score, rating(score))
	for _, r := range reasons {
		_, _ = fmt.Fprintf(stdio.Out, "  - %s\n", r)
	}
	return nil
}

// Score rates password from 0 to 100 and returns the reasoning behind the
// score. The model rewards length and character-class diversity and heavily
// penalizes passwords on the common-password list.
func Score(password string) (int, []string) {
	var reasons []string

	if commonPasswords[strings.ToLower(password)] {
		return 0, []string{"this is a very common password"}
	}

	score := 0

	switch {
	case len(password) >= 16:
		score += 40
		reasons = append(reasons, "good length (16+)")
	case len(password) >= 12:
		score += 30
		reasons = append(reasons, "decent length (12+)")
	case len(password) >= 8:
		score += 20
		reasons = append(reasons, "minimum length (8+)")
	default:
		reasons = append(reasons, "too short (under 8 characters)")
	}

	classes, classNames := classify(password)
	score += classes * 15
	reasons = append(reasons, fmt.Sprintf("uses %d character classes (%s)", classes, strings.Join(classNames, ", ")))

	if classes >= 3 && len(password) >= 12 {
		score += 5
		reasons = append(reasons, "strong mix of length and variety")
	}

	if score > 100 {
		score = 100
	}
	return score, reasons
}

// classify reports how many of the four character classes (lowercase,
// uppercase, digit, symbol) appear, and their names.
func classify(password string) (int, []string) {
	var lower, upper, digit, symbol bool
	for _, r := range password {
		switch {
		case unicode.IsLower(r):
			lower = true
		case unicode.IsUpper(r):
			upper = true
		case unicode.IsDigit(r):
			digit = true
		default:
			symbol = true
		}
	}
	var names []string
	for _, c := range []struct {
		on   bool
		name string
	}{{lower, "lowercase"}, {upper, "uppercase"}, {digit, "digits"}, {symbol, "symbols"}} {
		if c.on {
			names = append(names, c.name)
		}
	}
	return len(names), names
}

// rating turns a numeric score into a one-word verdict.
func rating(score int) string {
	switch {
	case score >= 80:
		return "strong"
	case score >= 50:
		return "fair"
	case score >= 25:
		return "weak"
	default:
		return "very weak"
	}
}

// commonPasswords is a small, freely-authored list of passwords that should
// always score zero.
var commonPasswords = map[string]bool{
	"password": true, "123456": true, "123456789": true, "qwerty": true,
	"abc123": true, "password1": true, "111111": true, "letmein": true,
	"admin": true, "welcome": true, "monkey": true, "iloveyou": true,
	"12345678": true, "1234567890": true, "dragon": true, "sunshine": true,
}
