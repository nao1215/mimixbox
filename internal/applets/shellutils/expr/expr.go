// Package expr implements the expr applet: evaluate an expression and write the
// result to standard output. Like GNU expr, the operands are not parsed with
// getopt; each argument is one token of the expression. The exit status encodes
// the result: 0 if the result is neither null nor 0, 1 if it is, 2 for an
// invalid expression, and 3 for any other error.
package expr

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the expr applet.
type Command struct{}

// New returns an expr command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "expr" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Evaluate expressions" }

// syntaxError marks an invalid expression, which maps to exit status 2.
type syntaxError struct{ msg string }

func (e *syntaxError) Error() string { return e.msg }

// Run executes expr. The arguments form the expression; eval reduces them to a
// single string. The result is printed to stdout, and the exit status reflects
// whether the result is null/zero (status 1) or not (status 0). A malformed
// expression prints "expr: <message>" to stderr and exits with status 2.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	if command.HandleHelpVersionWith(stdio, c.Name(), "EXPRESSION", command.Help{
		Description: "Evaluate EXPRESSION and write the result to standard output. Supports arithmetic " +
			"(+, -, *, /, %), comparisons (=, !=, <, <=, >, >=), the string operators : (match), " +
			"length, substr, index, and grouping with parentheses.",
		Examples: []command.Example{
			{Command: "expr 6 + 7", Explain: "Print 13."},
			{Command: `expr length "hello"`, Explain: "Print 5."},
		},
		ExitStatus: "0  the result is neither null nor zero.\n1  the result is null or zero.\n" +
			"2  the expression is malformed.\n3  an internal or write error occurred.",
	}, args) {
		return nil
	}
	result, err := eval(args)
	if err != nil {
		var se *syntaxError
		if errors.As(err, &se) {
			return &command.ExitError{Code: 2, Err: err}
		}
		return &command.ExitError{Code: 3, Err: err}
	}

	if _, err := fmt.Fprintln(stdio.Out, result); err != nil {
		return &command.ExitError{Code: 3, Err: err}
	}

	// GNU expr exits 1 when the printed result is null (empty) or zero.
	if isFalse(result) {
		return &command.ExitError{Code: command.ExitFailure}
	}
	return nil
}

// isFalse reports whether s is the empty string or the numeric value zero, the
// two values expr treats as "false" for its exit status.
func isFalse(s string) bool {
	if s == "" {
		return true
	}
	if n, err := strconv.Atoi(s); err == nil {
		return n == 0
	}
	return false
}

// Eval reduces the token slice to a single result string. It is the exported,
// testable entry point of the recursive-descent evaluator.
func Eval(tokens []string) (string, error) { return eval(tokens) }

// eval reduces the token slice to a single result string.
func eval(tokens []string) (string, error) {
	p := &parser{tokens: tokens}
	v, err := p.parseOr()
	if err != nil {
		return "", err
	}
	if !p.atEnd() {
		return "", &syntaxError{msg: fmt.Sprintf("syntax error: unexpected argument %q", p.peek())}
	}
	return v, nil
}

// parser walks the token slice with a single-token lookahead. The grammar, from
// lowest to highest precedence:
//
//	or:      and        ( '|' and )*
//	and:     compare    ( '&' compare )*
//	compare: addsub     ( ('<'|'<='|'='|'!='|'>='|'>') addsub )*
//	addsub:  muldiv     ( ('+'|'-') muldiv )*
//	muldiv:  match      ( ('*'|'/'|'%') match )*
//	match:   primary    ( ':' primary )*
//	primary: '(' or ')' | keyword-op | literal
type parser struct {
	tokens []string
	pos    int
}

func (p *parser) atEnd() bool { return p.pos >= len(p.tokens) }

func (p *parser) peek() string {
	if p.atEnd() {
		return ""
	}
	return p.tokens[p.pos]
}

func (p *parser) next() string {
	t := p.peek()
	p.pos++
	return t
}

func (p *parser) parseOr() (string, error) {
	left, err := p.parseAnd()
	if err != nil {
		return "", err
	}
	for p.peek() == "|" {
		p.next()
		right, err := p.parseAnd()
		if err != nil {
			return "", err
		}
		// '|' yields its first argument if it is neither null nor zero,
		// otherwise its second argument.
		if !isFalse(left) {
			continue
		}
		left = right
	}
	return left, nil
}

func (p *parser) parseAnd() (string, error) {
	left, err := p.parseCompare()
	if err != nil {
		return "", err
	}
	for p.peek() == "&" {
		p.next()
		right, err := p.parseCompare()
		if err != nil {
			return "", err
		}
		// '&' yields its first argument if neither argument is null or
		// zero, otherwise 0.
		if isFalse(left) || isFalse(right) {
			left = "0"
		}
	}
	return left, nil
}

func (p *parser) parseCompare() (string, error) {
	left, err := p.parseAddSub()
	if err != nil {
		return "", err
	}
	for isCompareOp(p.peek()) {
		op := p.next()
		right, err := p.parseAddSub()
		if err != nil {
			return "", err
		}
		left = boolToStr(compare(left, right, op))
	}
	return left, nil
}

func (p *parser) parseAddSub() (string, error) {
	left, err := p.parseMulDiv()
	if err != nil {
		return "", err
	}
	for p.peek() == "+" || p.peek() == "-" {
		op := p.next()
		right, err := p.parseMulDiv()
		if err != nil {
			return "", err
		}
		l, r, err := toInts(left, right)
		if err != nil {
			return "", err
		}
		if op == "+" {
			left = strconv.Itoa(l + r)
		} else {
			left = strconv.Itoa(l - r)
		}
	}
	return left, nil
}

func (p *parser) parseMulDiv() (string, error) {
	left, err := p.parseMatch()
	if err != nil {
		return "", err
	}
	for p.peek() == "*" || p.peek() == "/" || p.peek() == "%" {
		op := p.next()
		right, err := p.parseMatch()
		if err != nil {
			return "", err
		}
		l, r, err := toInts(left, right)
		if err != nil {
			return "", err
		}
		switch op {
		case "*":
			left = strconv.Itoa(l * r)
		case "/":
			if r == 0 {
				return "", &syntaxError{msg: "division by zero"}
			}
			left = strconv.Itoa(l / r)
		case "%":
			if r == 0 {
				return "", &syntaxError{msg: "division by zero"}
			}
			left = strconv.Itoa(l % r)
		}
	}
	return left, nil
}

func (p *parser) parseMatch() (string, error) {
	left, err := p.parsePrimary()
	if err != nil {
		return "", err
	}
	for p.peek() == ":" {
		p.next()
		right, err := p.parsePrimary()
		if err != nil {
			return "", err
		}
		res, err := matchOp(left, right)
		if err != nil {
			return "", err
		}
		left = res
	}
	return left, nil
}

func (p *parser) parsePrimary() (string, error) {
	if p.atEnd() {
		return "", &syntaxError{msg: "syntax error: missing argument"}
	}

	switch p.peek() {
	case "(":
		p.next()
		v, err := p.parseOr()
		if err != nil {
			return "", err
		}
		if p.peek() != ")" {
			return "", &syntaxError{msg: "syntax error: expecting ')'"}
		}
		p.next()
		return v, nil
	case "length":
		p.next()
		s, err := p.argFor("length")
		if err != nil {
			return "", err
		}
		return strconv.Itoa(len([]rune(s))), nil
	case "substr":
		return p.parseSubstr()
	case "index":
		return p.parseIndex()
	case "match":
		p.next()
		s, err := p.argFor("match")
		if err != nil {
			return "", err
		}
		re, err := p.argFor("match")
		if err != nil {
			return "", err
		}
		return matchOp(s, re)
	}

	return p.next(), nil
}

// argFor consumes and returns the next token as a literal operand of a keyword
// operator, reporting a syntax error when none remains.
func (p *parser) argFor(op string) (string, error) {
	if p.atEnd() {
		return "", &syntaxError{msg: fmt.Sprintf("syntax error: missing argument after %q", op)}
	}
	return p.next(), nil
}

func (p *parser) parseSubstr() (string, error) {
	p.next()
	s, err := p.argFor("substr")
	if err != nil {
		return "", err
	}
	posStr, err := p.argFor("substr")
	if err != nil {
		return "", err
	}
	lenStr, err := p.argFor("substr")
	if err != nil {
		return "", err
	}
	pos, err1 := strconv.Atoi(posStr)
	length, err2 := strconv.Atoi(lenStr)
	if err1 != nil || err2 != nil {
		// GNU expr returns the empty string for non-numeric pos/length.
		return "", nil
	}
	runes := []rune(s)
	// expr positions are 1-based.
	if pos < 1 || length < 0 || pos > len(runes) {
		return "", nil
	}
	start := pos - 1
	end := start + length
	if end > len(runes) {
		end = len(runes)
	}
	return string(runes[start:end]), nil
}

func (p *parser) parseIndex() (string, error) {
	p.next()
	s, err := p.argFor("index")
	if err != nil {
		return "", err
	}
	chars, err := p.argFor("index")
	if err != nil {
		return "", err
	}
	runes := []rune(s)
	set := make(map[rune]bool, len(chars))
	for _, ch := range chars {
		set[ch] = true
	}
	for i, ch := range runes {
		if set[ch] {
			return strconv.Itoa(i + 1), nil
		}
	}
	return "0", nil
}

// isCompareOp reports whether tok is one of the six comparison operators.
func isCompareOp(tok string) bool {
	switch tok {
	case "<", "<=", "=", "!=", ">=", ">":
		return true
	}
	return false
}

// compare evaluates a comparison. When both operands look like integers the
// comparison is numeric, otherwise it is lexicographic.
func compare(left, right, op string) bool {
	l, lerr := strconv.Atoi(left)
	r, rerr := strconv.Atoi(right)
	if lerr == nil && rerr == nil {
		switch op {
		case "<":
			return l < r
		case "<=":
			return l <= r
		case "=":
			return l == r
		case "!=":
			return l != r
		case ">=":
			return l >= r
		case ">":
			return l > r
		}
	}
	switch op {
	case "<":
		return left < right
	case "<=":
		return left <= right
	case "=":
		return left == right
	case "!=":
		return left != right
	case ">=":
		return left >= right
	case ">":
		return left > right
	}
	return false
}

func boolToStr(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

// toInts converts both operands to integers for arithmetic, reporting a syntax
// error (exit status 2) when either is not a valid integer.
func toInts(left, right string) (int, int, error) {
	l, err := strconv.Atoi(left)
	if err != nil {
		return 0, 0, &syntaxError{msg: fmt.Sprintf("non-integer argument: %q", left)}
	}
	r, err := strconv.Atoi(right)
	if err != nil {
		return 0, 0, &syntaxError{msg: fmt.Sprintf("non-integer argument: %q", right)}
	}
	return l, r, nil
}

// matchOp implements the ':' / match operator: an anchored match of pattern
// against s. When the pattern has a \( \) capture group, the matched substring
// is returned; otherwise the number of matched characters is returned.
func matchOp(s, pattern string) (string, error) {
	re, err := regexp.Compile("^" + basicToGoRegexp(pattern))
	if err != nil {
		return "", &syntaxError{msg: fmt.Sprintf("invalid regular expression: %v", err)}
	}
	m := re.FindStringSubmatch(s)
	if m == nil {
		if strings.Contains(pattern, `\(`) {
			return "", nil
		}
		return "0", nil
	}
	if len(m) > 1 {
		return m[1], nil
	}
	return strconv.Itoa(len([]rune(m[0]))), nil
}

// basicToGoRegexp translates the POSIX basic regular expression syntax expr
// uses into Go's (RE2) syntax. In a BRE the grouping and interval metacharacters
// are written escaped (\( \) \{ \}) and the unescaped forms are literal, which
// is the reverse of RE2.
func basicToGoRegexp(p string) string {
	var b strings.Builder
	for i := 0; i < len(p); i++ {
		switch {
		case p[i] == '\\' && i+1 < len(p):
			switch p[i+1] {
			case '(', ')', '{', '}':
				// Escaped in a BRE: these are the metacharacter forms.
				b.WriteByte(p[i+1])
			default:
				b.WriteByte('\\')
				b.WriteByte(p[i+1])
			}
			i++
		case p[i] == '(' || p[i] == ')' || p[i] == '{' || p[i] == '}' || p[i] == '+' || p[i] == '?' || p[i] == '|':
			// Unescaped: literal in a BRE.
			b.WriteByte('\\')
			b.WriteByte(p[i])
		default:
			b.WriteByte(p[i])
		}
	}
	return b.String()
}
