package expr_test

import (
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/shellutils/expr"
)

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := expr.New()
	if c.Name() != "expr" {
		t.Errorf("Name() = %q, want %q", c.Name(), "expr")
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}

// TestEvalLexicographicCompare drives compare()'s string (non-numeric) branch
// for every comparison operator, where at least one operand is not an integer.
func TestEvalLexicographicCompare(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"lt true", []string{"apple", "<", "banana"}, "1"},
		{"lt false", []string{"banana", "<", "apple"}, "0"},
		{"le equal", []string{"abc", "<=", "abc"}, "1"},
		{"eq false", []string{"foo", "=", "bar"}, "0"},
		{"ne true", []string{"foo", "!=", "bar"}, "1"},
		{"ge true", []string{"b", ">=", "a"}, "1"},
		{"gt false", []string{"a", ">", "b"}, "0"},
		// One numeric and one non-numeric operand still compares as strings.
		{"mixed lt", []string{"9", "<", "abc"}, "1"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := expr.Eval(tt.args)
			if err != nil {
				t.Fatalf("Eval(%v) error = %v", tt.args, err)
			}
			if got != tt.want {
				t.Errorf("Eval(%v) = %q, want %q", tt.args, got, tt.want)
			}
		})
	}
}

// TestEvalMatchOp covers matchOp's several outcomes: a match without a capture
// group yields the matched length, a match with a group yields the captured
// text, a non-match without a group yields "0", and a non-match with a group
// yields the empty string.
func TestEvalMatchOp(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"length of match", []string{"abc123", ":", "[a-z]*"}, "3"},
		{"capture group", []string{"hello", ":", `h\(.*\)o`}, "ell"},
		{"no match no group", []string{"xyz", ":", "abc"}, "0"},
		{"no match with group", []string{"xyz", ":", `a\(bc\)`}, ""},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := expr.Eval(tt.args)
			if err != nil {
				t.Fatalf("Eval(%v) error = %v", tt.args, err)
			}
			if got != tt.want {
				t.Errorf("Eval(%v) = %q, want %q", tt.args, got, tt.want)
			}
		})
	}
}

// TestEvalInvalidRegex covers matchOp's compile-error branch (an unmatched
// bracket is an invalid regular expression), which must be a syntax error.
func TestEvalInvalidRegex(t *testing.T) {
	t.Parallel()
	if _, err := expr.Eval([]string{"abc", ":", "[unterminated"}); err == nil {
		t.Fatal("expected a syntax error for an invalid regular expression")
	}
}

// TestEvalBasicToGoRegexp exercises basicToGoRegexp's translation of POSIX BRE
// metacharacters: unescaped +, ?, | are literals in a BRE, so they only match
// themselves rather than acting as regex operators.
func TestEvalBasicToGoRegexp(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		args []string
		want string
	}{
		// '+' is literal in a BRE: it matches a literal plus sign.
		{"literal plus", []string{"a+b", ":", `a+b`}, "3"},
		// '?' is literal in a BRE.
		{"literal question", []string{"a?", ":", `a?`}, "2"},
		// '|' is literal in a BRE (not alternation).
		{"literal pipe", []string{"a|b", ":", `a|b`}, "3"},
		// '\+' is passed through as an escaped plus, i.e. a literal '+' in RE2,
		// so it matches a literal "a+" rather than acting as a repeat operator.
		{"escaped plus literal", []string{"a+", ":", `a\+`}, "2"},
		// '\{ \}' are translated to RE2 interval braces.
		{"interval", []string{"aaa", ":", `a\{2\}`}, "2"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := expr.Eval(tt.args)
			if err != nil {
				t.Fatalf("Eval(%v) error = %v", tt.args, err)
			}
			if got != tt.want {
				t.Errorf("Eval(%v) = %q, want %q", tt.args, got, tt.want)
			}
		})
	}
}

// TestEvalSubstrNonNumeric covers parseSubstr's non-numeric pos/length branch,
// which GNU expr resolves to the empty string.
func TestEvalSubstrNonNumeric(t *testing.T) {
	t.Parallel()
	tests := [][]string{
		{"substr", "abcdef", "x", "3"},
		{"substr", "abcdef", "2", "y"},
	}
	for _, args := range tests {
		got, err := expr.Eval(args)
		if err != nil {
			t.Fatalf("Eval(%v) error = %v", args, err)
		}
		if got != "" {
			t.Errorf("Eval(%v) = %q, want empty string", args, got)
		}
	}
}

// TestEvalSubstrBoundaries covers parseSubstr's out-of-range guards: a position
// past the end, a position below one, and a length that runs past the end.
func TestEvalSubstrBoundaries(t *testing.T) {
	t.Parallel()
	tests := []struct {
		args []string
		want string
	}{
		{[]string{"substr", "abc", "0", "2"}, ""},    // pos < 1
		{[]string{"substr", "abc", "9", "2"}, ""},    // pos > len
		{[]string{"substr", "abc", "2", "-1"}, ""},   // negative length
		{[]string{"substr", "abc", "2", "10"}, "bc"}, // length clamps to end
		{[]string{"substr", "abc", "1", "0"}, ""},    // zero length
	}
	for _, tt := range tests {
		got, err := expr.Eval(tt.args)
		if err != nil {
			t.Fatalf("Eval(%v) error = %v", tt.args, err)
		}
		if got != tt.want {
			t.Errorf("Eval(%v) = %q, want %q", tt.args, got, tt.want)
		}
	}
}

// TestEvalMissingKeywordArgs covers argFor's missing-argument error branch for
// each keyword operator that consumes operands.
func TestEvalMissingKeywordArgs(t *testing.T) {
	t.Parallel()
	tests := [][]string{
		{"length"},
		{"substr", "abc"},
		{"substr", "abc", "1"},
		{"index", "abc"},
		{"match", "abc"},
	}
	for _, args := range tests {
		if _, err := expr.Eval(args); err == nil {
			t.Errorf("Eval(%v) = nil error, want missing-argument syntax error", args)
		}
	}
}

// TestEvalMissingPrimary covers parsePrimary's atEnd guard (an operator with no
// right-hand operand) and the unbalanced-parenthesis path.
func TestEvalMissingPrimary(t *testing.T) {
	t.Parallel()
	tests := [][]string{
		{"1", "+"},      // missing right operand
		{"(", "1", "+"}, // missing operand inside group
		{"(", "1"},      // unbalanced paren
	}
	for _, args := range tests {
		if _, err := expr.Eval(args); err == nil {
			t.Errorf("Eval(%v) = nil error, want syntax error", args)
		}
	}
}
