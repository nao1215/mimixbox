// Package dc implements the dc applet: a reverse-Polish (stack) desk calculator.
// It supports integer and fixed-point decimal arithmetic, registers, and the
// precision register, which is enough for scripted calculations.
package dc

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

// Command is the dc applet.
type Command struct{}

// New returns a dc command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "dc" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Reverse-Polish (stack) desk calculator" }

// num is a value with a display scale (number of fractional digits).
type num struct {
	v     *big.Rat
	scale int
}

// machine is the dc evaluator state.
type machine struct {
	stack []num
	regs  map[byte]num
	scale int // the precision register (k)
	out   io.Writer
	errw  io.Writer
}

// Run executes dc.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-e EXPR]... [FILE]...", stdio.Err).WithHelp(command.Help{
		Description: "Evaluate reverse-Polish expressions: numbers push onto a stack and operators " +
			"pop their operands. Read -e EXPR strings and FILEs in order, or standard input when " +
			"neither is given.",
		Examples: []command.Example{
			{Command: "dc -e '6 3 / p'", Explain: "Print 2."},
			{Command: "dc -e '2k 7 3 / p'", Explain: "Print 2.33 (two-digit precision)."},
		},
		Notes: []string{
			"Commands: + - * / % ^ (arithmetic); p n f (print); c d r (stack); sX lX (registers); k K (precision); q (quit).",
		},
	})
	var exprs []string
	fs.StringArrayVarP(&exprs, "expression", "e", nil, "evaluate EXPR")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	m := &machine{regs: map[byte]num{}, out: stdio.Out, errw: stdio.Err}

	for _, e := range exprs {
		m.eval(e)
	}
	files := fs.Args()
	for _, f := range files {
		data, rerr := os.ReadFile(f) //nolint:gosec // user-named file
		if rerr != nil {
			_, _ = fmt.Fprintf(stdio.Err, "dc: %s\n", command.FileError(f, rerr))
			return command.SilentFailure()
		}
		m.eval(string(data))
	}
	if len(exprs) == 0 && len(files) == 0 {
		sc := bufio.NewScanner(stdio.In)
		sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		for sc.Scan() {
			m.eval(sc.Text())
		}
	}
	return nil
}

// eval runs the dc command string.
func (m *machine) eval(s string) {
	i := 0
	for i < len(s) {
		ch := s[i]
		switch {
		case ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r':
			i++
		case isNumberStart(ch):
			tok, next := readNumber(s, i)
			i = next
			if n, ok := parseNum(tok); ok {
				m.push(n)
			} else {
				_, _ = fmt.Fprintln(m.errw, "dc: invalid number")
			}
		case ch == 's' || ch == 'l':
			if i+1 < len(s) {
				m.register(ch, s[i+1])
				i += 2
			} else {
				i++
			}
		default:
			m.command(ch)
			i++
		}
	}
}

func isNumberStart(ch byte) bool {
	return (ch >= '0' && ch <= '9') || ch == '.' || ch == '_'
}

// readNumber consumes a number token starting at i.
func readNumber(s string, i int) (string, int) {
	j := i
	if s[j] == '_' {
		j++
	}
	for j < len(s) && ((s[j] >= '0' && s[j] <= '9') || s[j] == '.') {
		j++
	}
	return s[i:j], j
}

// parseNum converts a dc number token (with '_' for negative) to a num.
func parseNum(tok string) (num, bool) {
	neg := false
	if strings.HasPrefix(tok, "_") {
		neg = true
		tok = tok[1:]
	}
	scale := 0
	if dot := strings.IndexByte(tok, '.'); dot >= 0 {
		scale = len(tok) - dot - 1
	}
	r := new(big.Rat)
	if _, ok := r.SetString(tok); !ok {
		return num{}, false
	}
	if neg {
		r.Neg(r)
	}
	return num{v: r, scale: scale}, true
}

func (m *machine) push(n num) { m.stack = append(m.stack, n) }

func (m *machine) pop() (num, bool) {
	if len(m.stack) == 0 {
		_, _ = fmt.Fprintln(m.errw, "dc: stack empty")
		return num{}, false
	}
	n := m.stack[len(m.stack)-1]
	m.stack = m.stack[:len(m.stack)-1]
	return n, true
}

// register handles sX (store) and lX (load).
func (m *machine) register(op, name byte) {
	if op == 's' {
		if n, ok := m.pop(); ok {
			m.regs[name] = n
		}
		return
	}
	if n, ok := m.regs[name]; ok {
		m.push(n)
	} else {
		m.push(num{v: new(big.Rat), scale: 0})
	}
}

// command runs a single-character command.
func (m *machine) command(ch byte) {
	switch ch {
	case '+', '-', '*', '/', '%', '^':
		m.binary(ch)
	case 'p':
		if len(m.stack) > 0 {
			_, _ = fmt.Fprintln(m.out, format(m.stack[len(m.stack)-1]))
		} else {
			_, _ = fmt.Fprintln(m.errw, "dc: stack empty")
		}
	case 'n':
		if n, ok := m.pop(); ok {
			_, _ = fmt.Fprint(m.out, format(n))
		}
	case 'f':
		for i := len(m.stack) - 1; i >= 0; i-- {
			_, _ = fmt.Fprintln(m.out, format(m.stack[i]))
		}
	case 'c':
		m.stack = nil
	case 'd':
		if len(m.stack) > 0 {
			m.push(m.stack[len(m.stack)-1])
		} else {
			_, _ = fmt.Fprintln(m.errw, "dc: stack empty")
		}
	case 'r':
		if len(m.stack) >= 2 {
			n := len(m.stack)
			m.stack[n-1], m.stack[n-2] = m.stack[n-2], m.stack[n-1]
		}
	case 'k':
		if n, ok := m.pop(); ok {
			m.scale = int(ratToInt(n.v))
		}
	case 'K':
		m.push(num{v: new(big.Rat).SetInt64(int64(m.scale)), scale: 0})
	case 'q', 'Q':
		// Quit: in this non-macro implementation, stop evaluating further input
		// by clearing the rest is not trivial; treat as a no-op terminator.
	default:
		_, _ = fmt.Fprintf(m.errw, "dc: %c (%d) unimplemented\n", ch, ch)
	}
}

// binary pops two operands and pushes op applied to them.
func (m *machine) binary(op byte) {
	b, ok := m.pop()
	if !ok {
		return
	}
	a, ok := m.pop()
	if !ok {
		m.push(b) // restore
		return
	}
	res := new(big.Rat)
	scale := 0
	switch op {
	case '+':
		res.Add(a.v, b.v)
		scale = max(a.scale, b.scale)
	case '-':
		res.Sub(a.v, b.v)
		scale = max(a.scale, b.scale)
	case '*':
		res.Mul(a.v, b.v)
		scale = a.scale + b.scale
	case '/':
		if b.v.Sign() == 0 {
			_, _ = fmt.Fprintln(m.errw, "dc: divide by zero")
			m.push(a)
			m.push(b)
			return
		}
		res.Quo(a.v, b.v)
		res = truncate(res, m.scale)
		scale = m.scale
	case '%':
		if b.v.Sign() == 0 {
			_, _ = fmt.Fprintln(m.errw, "dc: remainder by zero")
			m.push(a)
			m.push(b)
			return
		}
		q := truncate(new(big.Rat).Quo(a.v, b.v), m.scale)
		res.Sub(a.v, new(big.Rat).Mul(q, b.v))
		scale = max(a.scale, b.scale+m.scale)
	case '^':
		res = power(a.v, b.v)
		scale = a.scale * int(ratToInt(b.v))
		if scale < 0 {
			scale = m.scale
		}
	}
	m.push(num{v: res, scale: scale})
}

// power raises base to an integer exponent (the fractional part is ignored, as
// in dc for non-integer exponents in this slice).
func power(base, exp *big.Rat) *big.Rat {
	e := ratToInt(exp)
	result := new(big.Rat).SetInt64(1)
	b := new(big.Rat).Set(base)
	neg := e < 0
	if neg {
		e = -e
	}
	for ; e > 0; e-- {
		result.Mul(result, b)
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
	scaled.Quo(scaled, r.Denom()) // integer division truncates toward zero
	out := new(big.Rat).SetFrac(scaled, scale)
	return out
}

// ratToInt returns the integer part of r (truncated toward zero).
func ratToInt(r *big.Rat) int64 {
	q := new(big.Int).Quo(r.Num(), r.Denom())
	return q.Int64()
}

// format renders a num with its scale's worth of fractional digits.
func format(n num) string {
	if n.scale <= 0 {
		return new(big.Int).Quo(n.v.Num(), n.v.Denom()).String()
	}
	return n.v.FloatString(n.scale)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
