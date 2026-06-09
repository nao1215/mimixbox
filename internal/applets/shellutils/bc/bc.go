// Package bc implements the bc applet: an arbitrary-precision infix calculator.
// It supports the everyday subset - the four operators plus % and ^, parentheses
// and precedence, unary minus, variables, and the scale precision control - which
// covers scripted arithmetic without the full bc programming language.
package bc

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"math/big"
	"os"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the bc applet.
type Command struct{}

// New returns a bc command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "bc" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "An arbitrary-precision calculator language" }

// num is a value with a display scale (number of fractional digits).
type num struct {
	v     *big.Rat
	scale int
}

// machine holds the variables and the current precision.
type machine struct {
	vars  map[string]num
	scale int
	out   io.Writer
	errw  io.Writer
}

// Run executes bc.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[FILE]...", stdio.Err).WithHelp(command.Help{
		Description: "Evaluate infix arithmetic. Each non-assignment statement prints its value. " +
			"Statements are separated by newlines or ';'. Read FILEs in order, or standard input " +
			"when none are given.",
		Examples: []command.Example{
			{Command: "echo '2 + 3 * 4' | bc", Explain: "Print 14."},
			{Command: "echo 'scale=2; 7/3' | bc", Explain: "Print 2.33."},
		},
		Notes: []string{
			"Supported: + - * / % ^, parentheses, unary minus, variables, and the scale precision variable.",
		},
	})
	_ = fs.BoolP("mathlib", "l", false, "accepted for compatibility; the math library is not loaded")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	m := &machine{vars: map[string]num{}, out: stdio.Out, errw: stdio.Err}

	files := fs.Args()
	if len(files) == 0 {
		sc := bufio.NewScanner(stdio.In)
		sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		for sc.Scan() {
			m.run(sc.Text())
		}
		return nil
	}
	for _, f := range files {
		data, rerr := os.ReadFile(f) //nolint:gosec // user-named file
		if rerr != nil {
			_, _ = fmt.Fprintf(stdio.Err, "bc: %s\n", command.FileError(f, rerr))
			return command.SilentFailure()
		}
		m.run(string(data))
	}
	return nil
}

// run evaluates every statement in s.
func (m *machine) run(s string) {
	toks := lex(s)
	p := &parser{toks: toks, m: m}
	for !p.atEnd() {
		p.skipSeparators()
		if p.atEnd() {
			break
		}
		printed, val, ok := p.statement()
		if !ok {
			return // a parse error was already reported
		}
		if printed {
			_, _ = fmt.Fprintln(m.out, format(val))
		}
	}
}

type tokKind int

const (
	tNum tokKind = iota
	tName
	tOp
	tSep
	tEOF
)

type token struct {
	kind tokKind
	val  string
}

// lex splits s into tokens.
func lex(s string) []token {
	var toks []token
	i := 0
	for i < len(s) {
		ch := s[i]
		switch {
		case ch == ' ' || ch == '\t' || ch == '\r':
			i++
		case ch == '\n' || ch == ';':
			toks = append(toks, token{tSep, string(ch)})
			i++
		case (ch >= '0' && ch <= '9') || ch == '.':
			j := i
			for j < len(s) && ((s[j] >= '0' && s[j] <= '9') || s[j] == '.') {
				j++
			}
			toks = append(toks, token{tNum, s[i:j]})
			i = j
		case (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_':
			j := i
			for j < len(s) && ((s[j] >= 'a' && s[j] <= 'z') || (s[j] >= 'A' && s[j] <= 'Z') || (s[j] >= '0' && s[j] <= '9') || s[j] == '_') {
				j++
			}
			toks = append(toks, token{tName, s[i:j]})
			i = j
		case strings.IndexByte("+-*/%^()=", ch) >= 0:
			toks = append(toks, token{tOp, string(ch)})
			i++
		default:
			i++ // ignore anything unrecognized
		}
	}
	return append(toks, token{tEOF, ""})
}

type parser struct {
	toks []token
	pos  int
	m    *machine
	err  bool
}

func (p *parser) peek() token  { return p.toks[p.pos] }
func (p *parser) next() token  { t := p.toks[p.pos]; p.pos++; return t }
func (p *parser) atEnd() bool  { return p.peek().kind == tEOF }
func (p *parser) fail(msg string) {
	if !p.err {
		_, _ = fmt.Fprintf(p.m.errw, "bc: %s\n", msg)
		p.err = true
	}
}

func (p *parser) skipSeparators() {
	for p.peek().kind == tSep {
		p.pos++
	}
}

// statement parses one assignment or expression. printed reports whether the
// value should be printed (true for a bare expression).
func (p *parser) statement() (printed bool, val num, ok bool) {
	// Assignment: NAME = expr.
	if p.peek().kind == tName && p.toks[p.pos+1].kind == tOp && p.toks[p.pos+1].val == "=" {
		name := p.next().val
		p.next() // '='
		v := p.expr()
		if p.err {
			return false, num{}, false
		}
		p.assign(name, v)
		return false, num{}, true
	}
	v := p.expr()
	if p.err {
		return false, num{}, false
	}
	return true, v, true
}

func (p *parser) assign(name string, v num) {
	if name == "scale" {
		p.m.scale = int(ratToInt(v.v))
		return
	}
	p.m.vars[name] = v
}

// expr -> addSub.
func (p *parser) expr() num { return p.addSub() }

func (p *parser) addSub() num {
	left := p.mulDiv()
	for p.peek().kind == tOp && (p.peek().val == "+" || p.peek().val == "-") {
		op := p.next().val
		right := p.mulDiv()
		left = apply(p.m, op, left, right)
	}
	return left
}

func (p *parser) mulDiv() num {
	left := p.unary()
	for p.peek().kind == tOp && (p.peek().val == "*" || p.peek().val == "/" || p.peek().val == "%") {
		op := p.next().val
		right := p.unary()
		left = apply(p.m, op, left, right)
	}
	return left
}

func (p *parser) unary() num {
	if p.peek().kind == tOp && p.peek().val == "-" {
		p.next()
		v := p.unary()
		return num{v: new(big.Rat).Neg(v.v), scale: v.scale}
	}
	return p.power()
}

func (p *parser) power() num {
	base := p.primary()
	if p.peek().kind == tOp && p.peek().val == "^" {
		p.next()
		exp := p.unary() // right-associative
		return apply(p.m, "^", base, exp)
	}
	return base
}

func (p *parser) primary() num {
	t := p.peek()
	switch {
	case t.kind == tOp && t.val == "(":
		p.next()
		v := p.expr()
		if p.peek().kind == tOp && p.peek().val == ")" {
			p.next()
		} else {
			p.fail("missing )")
		}
		return v
	case t.kind == tNum:
		p.next()
		n, ok := parseNum(t.val)
		if !ok {
			p.fail("invalid number " + t.val)
		}
		return n
	case t.kind == tName:
		p.next()
		if t.val == "scale" {
			return num{v: new(big.Rat).SetInt64(int64(p.m.scale)), scale: 0}
		}
		if v, ok := p.m.vars[t.val]; ok {
			return v
		}
		return num{v: new(big.Rat), scale: 0} // unset variables are zero
	default:
		p.fail("unexpected token")
		p.next()
		return num{v: new(big.Rat), scale: 0}
	}
}

// parseNum converts a decimal literal to a num.
func parseNum(tok string) (num, bool) {
	scale := 0
	if dot := strings.IndexByte(tok, '.'); dot >= 0 {
		scale = len(tok) - dot - 1
	}
	r := new(big.Rat)
	if _, ok := r.SetString(tok); !ok {
		return num{}, false
	}
	return num{v: r, scale: scale}, true
}

// apply computes op over two operands, following bc's scale rules.
func apply(m *machine, op string, a, b num) num {
	res := new(big.Rat)
	scale := 0
	switch op {
	case "+":
		res.Add(a.v, b.v)
		scale = maxInt(a.scale, b.scale)
	case "-":
		res.Sub(a.v, b.v)
		scale = maxInt(a.scale, b.scale)
	case "*":
		res.Mul(a.v, b.v)
		// bc caps the product scale at min(s1+s2, max(scale, s1, s2)) and truncates.
		scale = minInt(a.scale+b.scale, maxInt(m.scale, maxInt(a.scale, b.scale)))
		res = truncate(res, scale)
	case "/":
		if b.v.Sign() == 0 {
			_, _ = fmt.Fprintln(m.errw, "bc: divide by zero")
			return num{v: new(big.Rat), scale: 0}
		}
		res = truncate(new(big.Rat).Quo(a.v, b.v), m.scale)
		scale = m.scale
	case "%":
		if b.v.Sign() == 0 {
			_, _ = fmt.Fprintln(m.errw, "bc: divide by zero")
			return num{v: new(big.Rat), scale: 0}
		}
		q := truncate(new(big.Rat).Quo(a.v, b.v), m.scale)
		res.Sub(a.v, new(big.Rat).Mul(q, b.v))
		scale = maxInt(a.scale, b.scale+m.scale)
	case "^":
		res = power(a.v, b.v)
		scale = a.scale * int(ratToInt(b.v))
		if scale < 0 {
			scale = m.scale
		}
	}
	return num{v: res, scale: scale}
}

func power(base, exp *big.Rat) *big.Rat {
	e := ratToInt(exp)
	neg := e < 0
	if neg {
		e = -e
	}
	result := new(big.Rat).SetInt64(1)
	for ; e > 0; e-- {
		result.Mul(result, base)
	}
	if neg && result.Sign() != 0 {
		result.Inv(result)
	}
	return result
}

// truncate rounds r toward zero to k fractional digits.
func truncate(r *big.Rat, k int) *big.Rat {
	scale := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(k)), nil)
	scaled := new(big.Int).Mul(r.Num(), scale)
	scaled.Quo(scaled, r.Denom())
	return new(big.Rat).SetFrac(scaled, scale)
}

func ratToInt(r *big.Rat) int64 {
	return new(big.Int).Quo(r.Num(), r.Denom()).Int64()
}

// format renders a num with its scale's worth of fractional digits, using bc's
// convention of omitting the leading zero (".5", "-.5").
func format(n num) string {
	if n.scale <= 0 {
		return new(big.Int).Quo(n.v.Num(), n.v.Denom()).String()
	}
	s := n.v.FloatString(n.scale)
	switch {
	case strings.HasPrefix(s, "0."):
		return s[1:]
	case strings.HasPrefix(s, "-0."):
		return "-" + s[2:]
	default:
		return s
	}
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
