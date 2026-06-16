package nl_test

import (
	"strings"
	"testing"
)

// sectionInput is a header/body/footer fixture using nl's delimiter lines
// (\:\:\: header, \:\: body, \: footer). Every section is non-empty so the
// per-section STYLE choice is observable.
const sectionInput = "H1\n\\:\\:\\:\nHDR\n\\:\\:\nB1\n\\:\nF1\n"

func TestRunSectionStyles(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			// Default: only non-empty body lines are numbered; header and footer
			// styles default to "n" (no numbering). The first H1 is in the body
			// section (no header delimiter precedes it).
			name: "default header and footer unnumbered",
			args: nil,
			want: "     1\tH1\n\n       HDR\n\n     1\tB1\n\n       F1\n",
		},
		{
			// -h a -f a numbers every header and footer line too. Each section
			// delimiter resets the counter to the starting line number.
			name: "all sections numbered",
			args: []string{"-h", "a", "-b", "a", "-f", "a"},
			want: "     1\tH1\n\n     1\tHDR\n\n     1\tB1\n\n     1\tF1\n",
		},
		{
			// Mixed styles: header all, body non-empty, footer none.
			name: "mixed per-section styles",
			args: []string{"-h", "a", "-b", "t", "-f", "n"},
			want: "     1\tH1\n\n     1\tHDR\n\n     1\tB1\n\n       F1\n",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out, _, err := run(t, sectionInput, tt.args...)
			if err != nil {
				t.Fatalf("Run error = %v", err)
			}
			if out != tt.want {
				t.Errorf("out = %q, want %q", out, tt.want)
			}
		})
	}
}

func TestRunRegexpStyle(t *testing.T) {
	t.Parallel()
	// pBRE numbers only lines matching the pattern.
	out, _, err := run(t, "foo\nbar\nfoobar\n", "-b", "pfoo")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	want := "     1\tfoo\n       bar\n     2\tfoobar\n"
	if out != want {
		t.Errorf("out = %q, want %q", out, want)
	}
}

func TestRunJoinBlankLines(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			// -l 2 counts two consecutive blank lines as one numbered line:
			// the number lands on the 2nd and 4th blank.
			name: "join two blank lines",
			args: []string{"-b", "a", "-l", "2"},
			want: "     1\ta\n       \n     2\t\n       \n     3\t\n     4\tb\n",
		},
		{
			// -l 3 numbers only every third blank line.
			name: "join three blank lines",
			args: []string{"-b", "a", "-l", "3"},
			want: "     1\ta\n       \n       \n     2\t\n       \n     3\tb\n",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out, _, err := run(t, "a\n\n\n\n\nb\n", tt.args...)
			if err != nil {
				t.Fatalf("Run error = %v", err)
			}
			if out != tt.want {
				t.Errorf("out = %q, want %q", out, tt.want)
			}
		})
	}
}

func TestRunInvalidSectionStyles(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		args    []string
		wantMsg string
	}{
		{"invalid header", []string{"-h", "z"}, "invalid header numbering style"},
		{"invalid footer", []string{"-f", "z"}, "invalid footer numbering style"},
		{"invalid join", []string{"-l", "0"}, "invalid line number of blank lines"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, errOut, err := run(t, "a\n", tt.args...)
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(errOut, tt.wantMsg) {
				t.Errorf("stderr = %q, want it to mention %q", errOut, tt.wantMsg)
			}
		})
	}
}
