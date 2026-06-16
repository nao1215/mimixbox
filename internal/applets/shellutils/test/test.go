// Package testcmd implements the test applet: evaluate a conditional
// expression. Like the POSIX/shell builtin, test produces no output; its only
// result is the exit status (0 = true, 1 = false, >1 = error). Its operands form
// an expression (e.g. "test -f foo", "test 1 -eq 2"), so it deliberately does
// not use getopt-style flag parsing.
package testcmd

import (
	"context"
	"errors"
	"os"
	"strconv"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the test applet.
type Command struct{}

// New returns a test command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "test" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Evaluate a conditional expression" }

// Run evaluates the expression in args. It writes nothing to stdout; the result
// is the exit status only: nil for true (0), ExitError{Code: 1} for false, and
// ExitError{Code: 2} for a malformed expression (whose message the runner prints
// as "test: <message>").
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	// Like GNU test, honor --help/--version only when it is the sole argument so
	// that `test --help = x` still evaluates a string comparison.
	if len(args) == 1 && command.HandleHelpVersionWith(stdio, c.Name(), "EXPRESSION", command.Help{
		Description: "Evaluate a conditional EXPRESSION and exit with its truth value. Supports the " +
			"usual file tests (-e, -f, -d, ...), string tests (-z, -n, =, !=), integer comparisons " +
			"(-eq, -ne, -lt, ...), and the !/-a/-o operators.",
		Examples: []command.Example{
			{Command: "test -f /etc/hosts", Explain: "Succeed when /etc/hosts exists and is a regular file."},
			{Command: `test "$x" = y`, Explain: "Succeed when $x equals y."},
		},
		ExitStatus: "0  the expression is true.\n1  the expression is false.\n2  the expression is malformed.",
	}, args) {
		return nil
	}
	ok, err := eval(args)
	if err != nil {
		return &command.ExitError{Code: 2, Err: err}
	}
	if !ok {
		return &command.ExitError{Code: 1}
	}
	return nil
}

// errSyntax is the generic message for a malformed expression.
var errSyntax = errors.New("syntax error")

// eval evaluates the test expression in args and reports its truth value. It is
// a pure function (its only side effects are file-system stat calls), so it can
// be unit-tested without the command framework. A malformed expression returns
// a non-nil error.
func eval(args []string) (bool, error) {
	p := &parser{args: args}
	ok, err := p.parseExpr()
	if err != nil {
		return false, err
	}
	if !p.atEnd() {
		return false, errSyntax
	}
	return ok, nil
}

// parser is a small recursive-descent evaluator over the operand list. The
// grammar mirrors POSIX test: -o binds loosest, then -a, then a leading !, then
// the primaries (file/string/integer tests and parenthesized groups).
//
//	expr   := term { -o term }
//	term   := factor { -a factor }
//	factor := ! factor | primary
type parser struct {
	args []string
	pos  int
}

func (p *parser) atEnd() bool { return p.pos >= len(p.args) }

func (p *parser) peek() (string, bool) {
	if p.atEnd() {
		return "", false
	}
	return p.args[p.pos], true
}

func (p *parser) next() (string, bool) {
	tok, ok := p.peek()
	if ok {
		p.pos++
	}
	return tok, ok
}

func (p *parser) parseExpr() (bool, error) {
	left, err := p.parseTerm()
	if err != nil {
		return false, err
	}
	for {
		tok, ok := p.peek()
		if !ok || tok != "-o" {
			break
		}
		p.pos++ // consume -o
		right, err := p.parseTerm()
		if err != nil {
			return false, err
		}
		left = left || right
	}
	return left, nil
}

func (p *parser) parseTerm() (bool, error) {
	left, err := p.parseFactor()
	if err != nil {
		return false, err
	}
	for {
		tok, ok := p.peek()
		if !ok || tok != "-a" {
			break
		}
		p.pos++ // consume -a
		right, err := p.parseFactor()
		if err != nil {
			return false, err
		}
		left = left && right
	}
	return left, nil
}

func (p *parser) parseFactor() (bool, error) {
	tok, ok := p.peek()
	if ok && tok == "!" {
		p.pos++ // consume !
		v, err := p.parseFactor()
		if err != nil {
			return false, err
		}
		return !v, nil
	}
	return p.parsePrimary()
}

// parsePrimary evaluates a single primary: a parenthesized expression, a unary
// file/string test, or a binary string/integer comparison. It uses limited
// look-ahead on the surrounding tokens to disambiguate, the way POSIX test does
// for its fixed-arg-count cases.
func (p *parser) parsePrimary() (bool, error) {
	tok, ok := p.peek()
	if !ok {
		// An expression position with nothing left (e.g. "! " alone) is empty,
		// which test treats as false rather than a syntax error.
		return false, nil
	}

	// Parenthesized group: ( EXPR )
	if tok == "(" {
		p.pos++ // consume (
		v, err := p.parseExpr()
		if err != nil {
			return false, err
		}
		close, ok := p.next()
		if !ok || close != ")" {
			return false, errSyntax
		}
		return v, nil
	}

	// Binary operator: STR1 <op> STR2. Look at the token after the next one.
	if p.pos+1 < len(p.args) && isBinaryOp(p.args[p.pos+1]) {
		if p.pos+2 >= len(p.args) {
			return false, errSyntax
		}
		left, op, right := p.args[p.pos], p.args[p.pos+1], p.args[p.pos+2]
		p.pos += 3
		return evalBinary(left, op, right)
	}

	// Unary operator: -X STR, when the token is a known unary test and an
	// operand follows.
	if isUnaryOp(tok) {
		if p.pos+1 >= len(p.args) {
			return false, errSyntax
		}
		operand := p.args[p.pos+1]
		p.pos += 2
		return evalUnary(tok, operand)
	}

	// A bare string is true when it is non-empty.
	p.pos++
	return tok != "", nil
}

// isBinaryOp reports whether op is a binary string or integer operator.
func isBinaryOp(op string) bool {
	switch op {
	case "=", "!=",
		"-eq", "-ne", "-gt", "-ge", "-lt", "-le":
		return true
	}
	return false
}

// isUnaryOp reports whether op is a unary string or file-test operator.
func isUnaryOp(op string) bool {
	switch op {
	case "-z", "-n",
		"-e", "-f", "-d", "-r", "-w", "-x", "-s",
		"-L", "-h", "-b", "-c", "-p", "-S":
		return true
	}
	return false
}

// evalBinary evaluates "left op right".
func evalBinary(left, op, right string) (bool, error) {
	switch op {
	case "=":
		return left == right, nil
	case "!=":
		return left != right, nil
	}

	// Remaining operators are integer comparisons.
	l, err := strconv.Atoi(left)
	if err != nil {
		return false, errors.New("integer expression expected: " + left)
	}
	r, err := strconv.Atoi(right)
	if err != nil {
		return false, errors.New("integer expression expected: " + right)
	}
	switch op {
	case "-eq":
		return l == r, nil
	case "-ne":
		return l != r, nil
	case "-gt":
		return l > r, nil
	case "-ge":
		return l >= r, nil
	case "-lt":
		return l < r, nil
	case "-le":
		return l <= r, nil
	}
	return false, errSyntax
}

// evalUnary evaluates "op operand" for string and file tests.
func evalUnary(op, operand string) (bool, error) {
	switch op {
	case "-z":
		return operand == "", nil
	case "-n":
		return operand != "", nil
	}
	return evalFileTest(op, operand)
}

// evalFileTest evaluates a file-test primary. A test against a nonexistent path
// (or one whose mode does not match) is simply false, never an error.
func evalFileTest(op, path string) (bool, error) {
	switch op {
	case "-e", "-f", "-d", "-s", "-b", "-c", "-p", "-S":
		info, err := os.Stat(path)
		if err != nil {
			return false, nil
		}
		switch op {
		case "-e":
			return true, nil
		case "-f":
			return info.Mode().IsRegular(), nil
		case "-d":
			return info.IsDir(), nil
		case "-s":
			return info.Size() > 0, nil
		case "-b":
			return info.Mode()&os.ModeDevice != 0 && info.Mode()&os.ModeCharDevice == 0, nil
		case "-c":
			return info.Mode()&os.ModeCharDevice != 0, nil
		case "-p":
			return info.Mode()&os.ModeNamedPipe != 0, nil
		case "-S":
			return info.Mode()&os.ModeSocket != 0, nil
		}
	case "-L", "-h":
		info, err := os.Lstat(path)
		if err != nil {
			return false, nil
		}
		return info.Mode()&os.ModeSymlink != 0, nil
	case "-r", "-w", "-x":
		info, err := os.Stat(path)
		if err != nil {
			return false, nil
		}
		return hasPerm(info.Mode(), op), nil
	}
	return false, errSyntax
}

// hasPerm reports whether mode grants the permission named by op (-r/-w/-x) to
// any of owner, group or other. It is an approximation of access(2) that does
// not consult the caller's uid/gid, which keeps the function pure and portable.
func hasPerm(mode os.FileMode, op string) bool {
	perm := mode.Perm()
	switch op {
	case "-r":
		return perm&0o444 != 0
	case "-w":
		return perm&0o222 != 0
	case "-x":
		return perm&0o111 != 0
	}
	return false
}
