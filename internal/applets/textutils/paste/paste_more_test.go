package paste_test

import (
	"testing"
)

// TestDelimiterEscapes covers each backslash escape understood by -d, plus the
// "default" branch for an unknown escape and the multi-character separator
// cycle. Behaviour matches GNU paste.
func TestDelimiterEscapes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		delim string
		stdin string
		want  string
	}{
		{
			// \t expands to a tab.
			name:  "tab escape",
			delim: `\t`,
			stdin: "a\nb\n",
			want:  "a\tb\n",
		},
		{
			// \\ expands to a single backslash.
			name:  "backslash escape",
			delim: `\\`,
			stdin: "a\nb\n",
			want:  "a\\b\n",
		},
		{
			// \0 means "no separator".
			name:  "null escape joins with nothing",
			delim: `\0`,
			stdin: "a\nb\n",
			want:  "ab\n",
		},
		{
			// An unrecognised escape (\q) falls through to the literal character.
			name:  "unknown escape is literal",
			delim: `\q`,
			stdin: "a\nb\n",
			want:  "aqb\n",
		},
		{
			// A trailing lone backslash is emitted literally.
			name:  "trailing backslash literal",
			delim: `\`,
			stdin: "a\nb\n",
			want:  "a\\b\n",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out, _, err := run(t, tt.stdin, "-s", "-d", tt.delim)
			if err != nil {
				t.Fatalf("Run error = %v", err)
			}
			if out != tt.want {
				t.Errorf("out = %q, want %q", out, tt.want)
			}
		})
	}
}

// TestSerialDelimiterCycle checks that a multi-character LIST is reused in a
// cycle between successive lines of a serial paste.
func TestSerialDelimiterCycle(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "1\n2\n3\n4\n", "-s", "-d", ",;")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	// Separators cycle ",", ";", "," between the four values.
	if out != "1,2;3,4\n" {
		t.Errorf("out = %q, want 1,2;3,4", out)
	}
}

// TestParallelDelimiterCycle checks the separator cycle across columns when
// merging three files in parallel.
func TestParallelDelimiterCycle(t *testing.T) {
	t.Parallel()
	f1 := writeFile(t, "a\n")
	f2 := writeFile(t, "b\n")
	f3 := writeFile(t, "c\n")
	out, _, err := run(t, "", "-d", ",;", f1, f2, f3)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	// Column separators cycle ",", ";".
	if out != "a,b;c\n" {
		t.Errorf("out = %q, want a,b;c", out)
	}
}

// TestSerialMissingFileContinues confirms a missing operand is reported but the
// remaining readable files are still pasted (firstErr path in serial).
func TestSerialMissingFileContinues(t *testing.T) {
	t.Parallel()
	good := writeFile(t, "x\ny\n")
	out, _, err := run(t, "", "-s", "/no/such/file", good)
	if err == nil {
		t.Fatal("expected failure for the missing file")
	}
	if out != "x\ty\n" {
		t.Errorf("out = %q, want the good file still pasted", out)
	}
}

// TestParallelMissingFileContinues confirms the parallel path also reports a
// missing operand while merging the readable ones.
func TestParallelMissingFileContinues(t *testing.T) {
	t.Parallel()
	good := writeFile(t, "x\n")
	out, _, err := run(t, "", "/no/such/file", good)
	if err == nil {
		t.Fatal("expected failure for the missing file")
	}
	if out != "x\n" {
		t.Errorf("out = %q, want the good file still pasted", out)
	}
}
