package sortcmd_test

import (
	"strings"
	"testing"

	sortcmd "github.com/nao1215/mimixbox/internal/applets/shellutils/sort"
)

// TestSynopsis covers the one-line description helper.
func TestSynopsis(t *testing.T) {
	t.Parallel()
	if got := sortcmd.New().Synopsis(); got != "Sort lines of text files" {
		t.Errorf("Synopsis() = %q", got)
	}
}

// TestKeyDefVariants drives parseKey's acceptance of richer KEYDEFs: a
// "start,end" range (the end is accepted but ignored), and a trailing column
// offset / ordering flag such as "2.3" or "2n" that is stripped to the field.
func TestKeyDefVariants(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		stdin string
		args  []string
		want  string
	}{
		{"range keydef", "x 3\ny 1\nz 2\n", []string{"-k", "2,2"}, "y 1\nz 2\nx 3\n"},
		{"column offset stripped", "x 3\ny 1\nz 2\n", []string{"-k", "2.1"}, "y 1\nz 2\nx 3\n"},
		{"ordering flag stripped", "x 30\ny 1\nz 200\n", []string{"-n", "-k", "2n"}, "y 1\nx 30\nz 200\n"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out, _, err := run(t, tt.stdin, tt.args...)
			if err != nil {
				t.Fatalf("Run error = %v", err)
			}
			if out != tt.want {
				t.Errorf("out = %q, want %q", out, tt.want)
			}
		})
	}
}

// TestInvalidKeyDef drives parseKey's rejection paths: a non-numeric or
// out-of-range start field is an error.
func TestInvalidKeyDef(t *testing.T) {
	t.Parallel()
	for _, def := range []string{"abc", "0", ".5", ","} {
		def := def
		t.Run(def, func(t *testing.T) {
			t.Parallel()
			_, _, err := run(t, "a\nb\n", "-k", def)
			if err == nil {
				t.Fatalf("expected error for -k %q", def)
			}
			if !strings.Contains(err.Error(), "invalid key definition") {
				t.Errorf("err = %v", err)
			}
		})
	}
}

// TestEmptyInput drives splitLines' empty branch and the empty-output path: no
// input yields no output and no error.
func TestEmptyInput(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "" {
		t.Errorf("out = %q, want empty", out)
	}
}

// TestNumericNonNumericFields drives leadingNumber: lines whose key has no
// number sort as zero, ahead of lines with positive numbers.
func TestNumericNonNumericFields(t *testing.T) {
	t.Parallel()
	// "apple" has no leading number (0); "5x" parses as 5; "-3y" parses as -3.
	out, _, err := run(t, "5x\napple\n-3y\n", "-n")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "-3y\napple\n5x\n" {
		t.Errorf("out = %q, want %q", out, "-3y\napple\n5x\n")
	}
}

// TestSeparatorMultiFieldKey drives field()'s separator join branch: with -t and
// -k the key spans from the chosen field to the end of the line, rejoined with
// the separator.
func TestSeparatorMultiFieldKey(t *testing.T) {
	t.Parallel()
	// Sort on field 2 onward, separated by ':'. Keys: "b:z", "a:y", "c:x".
	out, _, err := run(t, "1:b:z\n2:a:y\n3:c:x\n", "-t", ":", "-k", "2")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "2:a:y\n1:b:z\n3:c:x\n" {
		t.Errorf("out = %q, want %q", out, "2:a:y\n1:b:z\n3:c:x\n")
	}
}

// TestKeyFieldBeyondLine drives field()'s out-of-range branch: a key field past
// the number of fields yields an empty key, so such lines sort first and stably.
func TestKeyFieldBeyondLine(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "lonely\nx y\n", "-k", "2")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	// "lonely" has no field 2 (empty key) and sorts before "x y" (key "y").
	if out != "lonely\nx y\n" {
		t.Errorf("out = %q, want %q", out, "lonely\nx y\n")
	}
}
