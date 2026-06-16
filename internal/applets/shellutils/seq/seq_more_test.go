package seq_test

import (
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/shellutils/seq"
)

// TestSynopsis covers the one-line description helper.
func TestSynopsis(t *testing.T) {
	t.Parallel()
	if got := seq.New().Synopsis(); got != "Print a column of numbers" {
		t.Errorf("Synopsis() = %q", got)
	}
}

// TestNegativeOperandsAfterTerminator drives escapeIfNegative: operands that
// follow an explicit "--" terminator must still be recognized as negative
// numbers rather than options.
func TestNegativeOperandsAfterTerminator(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		args []string
		want string
	}{
		// "--" then FIRST INCREMENT LAST, all negative: count up from -5 by 2.
		{"negative range after terminator", []string{"--", "-5", "2", "-1"}, "-5\n-3\n-1\n"},
		// "--" then a single negative LAST counts 1 down to -3.
		{"single negative last", []string{"--", "-3"}, ""},
		// "--" then a non-negative operand is passed through unchanged.
		{"non negative after terminator", []string{"--", "3"}, "1\n2\n3\n"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out, _, err := run(t, tt.args...)
			if err != nil {
				t.Fatalf("Run error = %v", err)
			}
			if out != tt.want {
				t.Errorf("out = %q, want %q", out, tt.want)
			}
		})
	}
}

// TestEqualWidthNoPaddingNeeded exercises padToEqualWidth where every value is
// already the maximum width, so the early-continue branch is taken.
func TestEqualWidthNoPaddingNeeded(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "-w", "10", "12")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	// 10, 11, 12 are all two digits already; no zero-padding is added.
	if out != "10\n11\n12\n" {
		t.Errorf("out = %q, want %q", out, "10\n11\n12\n")
	}
}

// TestEqualWidthNegativePadsAfterSign exercises padToEqualWidth's sign-aware
// branch: zeros are inserted after the leading minus sign.
func TestEqualWidthNegativePadsAfterSign(t *testing.T) {
	t.Parallel()
	// From -1 down to -100, step -1: the shorter values pad after the sign.
	out, _, err := run(t, "-w", "-1", "-1", "-100")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	lines := []string{"-001", "-002", "-003"}
	for _, want := range lines {
		if !contains(out, want+"\n") {
			t.Errorf("out missing zero-padded %q; got first lines:\n%.40s", want, out)
		}
	}
}

// TestFloatPrecisionFromOperands exercises precision/maxPrecision: the output
// uses the greatest fractional-digit count among the operands.
func TestFloatPrecisionFromOperands(t *testing.T) {
	t.Parallel()
	// Increment 0.25 has two fractional digits, so all output has two.
	out, _, err := run(t, "1", "0.25", "1.5")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "1.00\n1.25\n1.50\n" {
		t.Errorf("out = %q, want %q", out, "1.00\n1.25\n1.50\n")
	}
}

// TestExponentOperandPrecision exercises precision's exponent-trimming branch:
// the fractional digits are measured ignoring the exponent part.
func TestExponentOperandPrecision(t *testing.T) {
	t.Parallel()
	// 1e1 == 10; it is treated as a float operand but has zero fractional digits.
	out, _, err := run(t, "1", "1e1")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	// Float formatting with zero precision yields plain integers.
	if out != "1\n2\n3\n4\n5\n6\n7\n8\n9\n10\n" {
		t.Errorf("out = %q", out)
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
