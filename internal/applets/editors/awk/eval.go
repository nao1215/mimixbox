package awk

import (
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

// ruleKind distinguishes BEGIN, END and ordinary (main) rules.
type ruleKind int

const (
	ruleMain ruleKind = iota
	ruleBegin
	ruleEnd
)

// rule is one "pattern { action }" entry from the program.
type rule struct {
	kind    ruleKind
	pattern string // raw pattern text ("" means always)
	action  string // raw action text ("" means default: print $0)
}

// parseProgram splits a program into rules. Rules are separated by newlines (or
// by the boundaries of brace blocks); each rule is "pattern { action }",
// "pattern" or "{ action }".
func parseProgram(text string) ([]rule, error) {
	var rules []rule
	i := 0
	for i < len(text) {
		// Skip separators.
		for i < len(text) && (text[i] == ' ' || text[i] == '\t' || text[i] == '\n' || text[i] == ';') {
			i++
		}
		if i >= len(text) {
			break
		}
		// Read the pattern up to '{' or a rule separator.
		start := i
		for i < len(text) && text[i] != '{' && text[i] != '\n' {
			i++
		}
		pattern := strings.TrimSpace(text[start:i])
		action := ""
		if i < len(text) && text[i] == '{' {
			depth := 0
			as := i
			for i < len(text) {
				if text[i] == '{' {
					depth++
				} else if text[i] == '}' {
					depth--
					if depth == 0 {
						i++
						break
					}
				}
				i++
			}
			if depth != 0 {
				return nil, fmt.Errorf("syntax error: unbalanced braces")
			}
			inner := text[as+1 : i-1]
			action = strings.TrimSpace(inner)
		}

		r := rule{kind: ruleMain, pattern: pattern, action: action}
		switch pattern {
		case "BEGIN":
			r.kind, r.pattern = ruleBegin, ""
		case "END":
			r.kind, r.pattern = ruleEnd, ""
		}
		rules = append(rules, r)
	}
	if len(rules) == 0 {
		return nil, fmt.Errorf("empty program")
	}
	return rules, nil
}

// state is the runtime context shared across rules.
type state struct {
	out    io.Writer
	fs     string
	ofs    string
	vars   map[string]string
	fields []string // fields[0] == $0
	nf     int
	nr     int
}

// setLine splits a new input record into fields.
func (st *state) setLine(line string) {
	st.fields = []string{line}
	var parts []string
	switch {
	case st.fs == "" || st.fs == " ":
		parts = strings.Fields(line)
	case len(st.fs) == 1:
		parts = strings.Split(line, st.fs)
	default:
		re, err := regexp.Compile(st.fs)
		if err != nil {
			parts = strings.Split(line, st.fs)
		} else {
			parts = re.Split(line, -1)
		}
	}
	st.fields = append(st.fields, parts...)
	st.nf = len(parts)
}

// field returns the i-th field ($0 is the whole record).
func (st *state) field(i int) string {
	if i == 0 {
		if len(st.fields) > 0 {
			return st.fields[0]
		}
		return ""
	}
	if i >= 1 && i < len(st.fields) {
		return st.fields[i]
	}
	return ""
}

// match evaluates a rule pattern against the current record.
func (st *state) match(pattern string) bool {
	if pattern == "" {
		return true
	}
	if strings.HasPrefix(pattern, "/") && strings.HasSuffix(pattern, "/") && len(pattern) >= 2 {
		re, err := regexp.Compile(pattern[1 : len(pattern)-1])
		if err != nil {
			return false
		}
		return re.MatchString(st.field(0))
	}
	return st.evalCondition(pattern)
}

// comparisonOps are the relational operators understood in patterns, longest
// first so "<=" is matched before "<".
var comparisonOps = []string{"==", "!=", "<=", ">=", "<", ">"}

// evalCondition evaluates a simple comparison pattern such as NR==2 or $1=="x".
// A bare value is true when it is a non-zero number or a non-empty string.
func (st *state) evalCondition(expr string) bool {
	expr = strings.TrimSpace(expr)
	for _, op := range comparisonOps {
		if idx := strings.Index(expr, op); idx >= 0 {
			lhs := st.evalPrimary(strings.TrimSpace(expr[:idx]))
			rhs := st.evalPrimary(strings.TrimSpace(expr[idx+len(op):]))
			return compare(lhs, rhs, op)
		}
	}
	v := st.evalPrimary(expr)
	if n, err := strconv.ParseFloat(v, 64); err == nil {
		return n != 0
	}
	return v != ""
}

// compare applies a relational operator, numerically when both sides parse as
// numbers and lexically otherwise.
func compare(lhs, rhs, op string) bool {
	ln, lerr := strconv.ParseFloat(lhs, 64)
	rn, rerr := strconv.ParseFloat(rhs, 64)
	if lerr == nil && rerr == nil {
		switch op {
		case "==":
			return ln == rn
		case "!=":
			return ln != rn
		case "<":
			return ln < rn
		case "<=":
			return ln <= rn
		case ">":
			return ln > rn
		case ">=":
			return ln >= rn
		}
	}
	switch op {
	case "==":
		return lhs == rhs
	case "!=":
		return lhs != rhs
	case "<":
		return lhs < rhs
	case "<=":
		return lhs <= rhs
	case ">":
		return lhs > rhs
	case ">=":
		return lhs >= rhs
	}
	return false
}

// evalPrimary resolves a single value: $n, NR, NF, a "string" literal, a named
// -v variable, or a numeric/!bare token.
func (st *state) evalPrimary(tok string) string {
	tok = strings.TrimSpace(tok)
	switch {
	case tok == "":
		return ""
	case tok == "NR":
		return strconv.Itoa(st.nr)
	case tok == "NF":
		return strconv.Itoa(st.nf)
	case tok == "$0":
		return st.field(0)
	case strings.HasPrefix(tok, "$"):
		idxTok := tok[1:]
		switch idxTok {
		case "NF":
			return st.field(st.nf)
		case "NR":
			return st.field(st.nr)
		}
		if n, err := strconv.Atoi(idxTok); err == nil {
			return st.field(n)
		}
		return ""
	case len(tok) >= 2 && tok[0] == '"' && tok[len(tok)-1] == '"':
		return tok[1 : len(tok)-1]
	}
	if v, ok := st.vars[tok]; ok {
		return v
	}
	return tok
}

// exec runs a rule's action (default: print $0).
func (st *state) exec(r rule) {
	action := r.action
	if action == "" {
		_, _ = fmt.Fprintln(st.out, st.field(0))
		return
	}
	for _, stmt := range splitStatements(action) {
		st.execStatement(strings.TrimSpace(stmt))
	}
}

// splitStatements splits an action body on top-level ';' and newlines.
func splitStatements(action string) []string {
	fields := strings.FieldsFunc(action, func(r rune) bool {
		return r == ';' || r == '\n'
	})
	return fields
}

// execStatement runs one statement; only print/printf are supported.
func (st *state) execStatement(stmt string) {
	switch {
	case stmt == "print" || stmt == "":
		_, _ = fmt.Fprintln(st.out, st.field(0))
	case strings.HasPrefix(stmt, "printf"):
		args := strings.TrimSpace(strings.TrimPrefix(stmt, "printf"))
		st.doPrintf(args)
	case strings.HasPrefix(stmt, "print"):
		args := strings.TrimSpace(strings.TrimPrefix(stmt, "print"))
		st.doPrint(args)
	}
}

// doPrint prints comma-separated values joined by OFS.
func (st *state) doPrint(args string) {
	if args == "" {
		_, _ = fmt.Fprintln(st.out, st.field(0))
		return
	}
	parts := splitTopComma(args)
	vals := make([]string, 0, len(parts))
	for _, p := range parts {
		vals = append(vals, st.evalPrimary(strings.TrimSpace(p)))
	}
	_, _ = fmt.Fprintln(st.out, strings.Join(vals, st.ofs))
}

// doPrintf implements a minimal printf: the first argument is a "format" string
// and %s/%d are filled from the remaining values.
func (st *state) doPrintf(args string) {
	parts := splitTopComma(args)
	if len(parts) == 0 {
		return
	}
	format := st.evalPrimary(strings.TrimSpace(parts[0]))
	var vals []any
	for _, p := range parts[1:] {
		vals = append(vals, st.evalPrimary(strings.TrimSpace(p)))
	}
	_, _ = fmt.Fprintf(st.out, unescape(format), vals...)
}

// splitTopComma splits on commas that are not inside a double-quoted string.
func splitTopComma(s string) []string {
	var parts []string
	var b strings.Builder
	inQuote := false
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if ch == '"' {
			inQuote = !inQuote
		}
		if ch == ',' && !inQuote {
			parts = append(parts, b.String())
			b.Reset()
			continue
		}
		b.WriteByte(ch)
	}
	parts = append(parts, b.String())
	return parts
}

// unescape resolves the common backslash escapes in a printf format string.
func unescape(s string) string {
	r := strings.NewReplacer(`\n`, "\n", `\t`, "\t", `\\`, "\\")
	return r.Replace(s)
}
