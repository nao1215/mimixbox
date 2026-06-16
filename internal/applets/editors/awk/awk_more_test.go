package awk_test

import (
	"strings"
	"testing"
)

// TestLexicalComparisons drives compare()'s string (non-numeric) branch for
// each relational operator.
func TestLexicalComparisons(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		stdin   string
		program string
		want    string
	}{
		{"string less than", "apple\nzebra\n", `$1 < "m"`, "apple\n"},
		{"string greater than", "apple\nzebra\n", `$1 > "m"`, "zebra\n"},
		{"string le", "abc\nabd\n", `$1 <= "abc"`, "abc\n"},
		{"string ge", "abc\nabd\n", `$1 >= "abd"`, "abd\n"},
		{"string ne", "cat\ndog\n", `$1 != "cat"`, "dog\n"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out, _, err := run(t, tt.stdin, tt.program)
			if err != nil {
				t.Fatalf("err = %v", err)
			}
			if out != tt.want {
				t.Errorf("out = %q, want %q", out, tt.want)
			}
		})
	}
}

// TestNumericLeGe covers compare()'s numeric <=/>= arms.
func TestNumericLeGe(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "1\n2\n3\n", "$1 <= 2")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if out != "1\n2\n" {
		t.Errorf("out = %q, want 1,2", out)
	}
	out, _, err = run(t, "1\n2\n3\n", "$1 >= 2")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if out != "2\n3\n" {
		t.Errorf("out = %q, want 2,3", out)
	}
}

// TestBareValuePattern covers evalCondition's bare-value truthiness: a non-zero
// field is true, a zero field is false.
func TestBareValuePattern(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "0\n5\n0\n", "$1")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if out != "5\n" {
		t.Errorf("out = %q, want only the non-zero line", out)
	}
}

// TestRegexFieldSeparator covers setLine's multi-character (regular expression)
// field-separator branch.
func TestRegexFieldSeparator(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "a12b345c\n", "-F", "[0-9]+", "{print $2}")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if out != "b\n" {
		t.Errorf("out = %q, want b", out)
	}
}

// TestMultiCharLiteralSeparator covers setLine's multi-char separator that is
// not a valid regexp, exercising the strings.Split fallback. "[" alone is an
// invalid regexp, so the literal split path is taken.
func TestMultiCharLiteralSeparator(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "a[b\n", "-F", "[", "{print $2}")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if out != "b\n" {
		t.Errorf("out = %q, want b", out)
	}
}

// TestVarReferenceMiss covers evalPrimary's bare-token branch when a referenced
// name is not a -v variable: it is returned verbatim.
func TestVarReferenceMiss(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "x\n", "{print undefined_name}")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if out != "undefined_name\n" {
		t.Errorf("out = %q, want the bare token", out)
	}
}

// TestDollarNR covers evalPrimary's "$NR" branch ($ followed by NR).
func TestDollarNR(t *testing.T) {
	t.Parallel()
	// On record 1, $NR == $1; on record 2, $NR == $2.
	out, _, err := run(t, "a b c\nd e f\n", "{print $NR}")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if out != "a\ne\n" {
		t.Errorf("out = %q, want a,e", out)
	}
}

// TestDollarNonNumeric covers evalPrimary's "$<non-number>" branch, which
// resolves to the empty field.
func TestDollarNonNumeric(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "a b\n", "{print $foo}")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if out != "\n" {
		t.Errorf("out = %q, want empty line", out)
	}
}

// TestFieldOutOfRange covers field()'s out-of-range branch returning "".
func TestFieldOutOfRange(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "a b\n", "{print $9}")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if out != "\n" {
		t.Errorf("out = %q, want empty line", out)
	}
}

// TestBeginOnly covers hasMainOrEnd's false branch: a program with only a BEGIN
// block does not read input.
func TestBeginOnly(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "ignored input\n", `BEGIN{print "hi"}`)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if out != "hi\n" {
		t.Errorf("out = %q, want hi", out)
	}
}

// TestUnbalancedBraces covers parseProgram's unbalanced-braces error.
func TestUnbalancedBraces(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, "x\n", "{print")
	if err == nil {
		t.Fatal("expected error for unbalanced braces")
	}
	if !strings.Contains(errOut, "awk:") {
		t.Errorf("stderr = %q, want awk diagnostic", errOut)
	}
}

// TestEmptyProgram covers parseProgram's empty-program error.
func TestEmptyProgram(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, "x\n", "   ")
	if err == nil {
		t.Fatal("expected error for empty program")
	}
	if !strings.Contains(errOut, "awk:") {
		t.Errorf("stderr = %q, want awk diagnostic", errOut)
	}
}

// TestPrintfMultipleAndEscapes covers doPrintf with multiple values and the
// unescape() handling of \t and \n.
func TestPrintfMultipleAndEscapes(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "a b\n", `{printf "%s\t%s\n", $1, $2}`)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if out != "a\tb\n" {
		t.Errorf("out = %q, want a<tab>b", out)
	}
}

// TestPrintNRandNF covers evalPrimary's NR/NF arms used inside print.
func TestPrintNRandNF(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "x y z\n", "{print NR, NF}")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if out != "1 3\n" {
		t.Errorf("out = %q, want '1 3'", out)
	}
}

// TestInvalidRegexPattern covers match()'s regexp-compile-error branch, where a
// malformed /regex/ matches nothing rather than failing the run.
func TestInvalidRegexPattern(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "abc\n", "/[/{print}")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if out != "" {
		t.Errorf("out = %q, want no matches for invalid regex", out)
	}
}
