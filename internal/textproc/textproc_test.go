package textproc_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/textproc"
)

func TestCountReader(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input string
		want  textproc.Count
	}{
		{
			name:  "no trailing newline",
			input: "a\nb\nc",
			want:  textproc.Count{Lines: 2, Words: 3, Runes: 5, Bytes: 5, MaxLineWidth: 1},
		},
		{
			name:  "trailing newline",
			input: "a\nb\nc\n",
			want:  textproc.Count{Lines: 3, Words: 3, Runes: 6, Bytes: 6, MaxLineWidth: 1},
		},
		{
			name:  "words and spaces",
			input: "hello  world\nfoo\n",
			want:  textproc.Count{Lines: 2, Words: 3, Runes: 17, Bytes: 17, MaxLineWidth: 12},
		},
		{
			name:  "multibyte counts runes and bytes separately",
			input: "あ\n",
			want:  textproc.Count{Lines: 1, Words: 1, Runes: 2, Bytes: 4, MaxLineWidth: 1},
		},
		{
			name:  "empty",
			input: "",
			want:  textproc.Count{},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := textproc.CountReader(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("CountReader error = %v", err)
			}
			if got != tt.want {
				t.Errorf("CountReader(%q) = %+v, want %+v", tt.input, got, tt.want)
			}
		})
	}
}

func TestCountAdd(t *testing.T) {
	t.Parallel()
	a := textproc.Count{Lines: 1, Words: 2, Runes: 3, Bytes: 3, MaxLineWidth: 4}
	b := textproc.Count{Lines: 10, Words: 20, Runes: 30, Bytes: 30, MaxLineWidth: 2}
	got := a.Add(b)
	want := textproc.Count{Lines: 11, Words: 22, Runes: 33, Bytes: 33, MaxLineWidth: 4}
	if got != want {
		t.Errorf("Add = %+v, want %+v", got, want)
	}
}

func TestReverse(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"trailing newline", "a\nb\nc\n", "c\nb\na\n"},
		{"no trailing newline", "a\nb\nc", "cb\na\n"},
		{"single line", "only\n", "only\n"},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := textproc.Reverse(tt.input, "\n"); got != tt.want {
				t.Errorf("Reverse(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestNumbererCatN(t *testing.T) {
	t.Parallel()
	n := textproc.Numberer{Style: textproc.NumberAll, Start: 1, Increment: 1, Width: 6, Separator: "\t"}
	var buf bytes.Buffer
	if err := n.WriteTo(&buf, strings.NewReader("a\n\nb\n")); err != nil {
		t.Fatalf("WriteTo error = %v", err)
	}
	want := "     1\ta\n     2\t\n     3\tb\n"
	if buf.String() != want {
		t.Errorf("cat -n = %q, want %q", buf.String(), want)
	}
}

func TestNumbererCatB(t *testing.T) {
	t.Parallel()
	n := textproc.Numberer{Style: textproc.NumberNonBlank, Start: 1, Increment: 1, Width: 6, Separator: "\t"}
	var buf bytes.Buffer
	if err := n.WriteTo(&buf, strings.NewReader("a\n\nb\n")); err != nil {
		t.Fatalf("WriteTo error = %v", err)
	}
	want := "     1\ta\n\n     2\tb\n"
	if buf.String() != want {
		t.Errorf("cat -b = %q, want %q", buf.String(), want)
	}
}

func TestNumbererNLDefault(t *testing.T) {
	t.Parallel()
	// nl default: number non-blank lines, pad blank lines so columns align.
	n := textproc.Numberer{Style: textproc.NumberNonBlank, Start: 1, Increment: 1, Width: 6, Separator: "\t", PadBlank: true}
	var buf bytes.Buffer
	if err := n.WriteTo(&buf, strings.NewReader("a\n\nb\n")); err != nil {
		t.Fatalf("WriteTo error = %v", err)
	}
	want := "     1\ta\n       \n     2\tb\n"
	if buf.String() != want {
		t.Errorf("nl = %q, want %q", buf.String(), want)
	}
}

func TestNumbererJustify(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		justify textproc.NumberJustify
		want    string
	}{
		{"right", textproc.JustifyRight, "     1\ta\n"},
		{"right-zero", textproc.JustifyRightZero, "000001\ta\n"},
		{"left", textproc.JustifyLeft, "1     \ta\n"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			n := textproc.Numberer{Style: textproc.NumberAll, Start: 1, Increment: 1, Width: 6, Separator: "\t", Justify: tt.justify}
			var buf bytes.Buffer
			if err := n.WriteTo(&buf, strings.NewReader("a\n")); err != nil {
				t.Fatalf("WriteTo error = %v", err)
			}
			if buf.String() != tt.want {
				t.Errorf("%s = %q, want %q", tt.name, buf.String(), tt.want)
			}
		})
	}
}

func TestHeadLines(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input string
		n     int
		want  string
	}{
		{"fewer than available", "a\nb\nc\nd\n", 2, "a\nb\n"},
		{"more than available", "a\nb\n", 5, "a\nb\n"},
		{"no trailing newline", "a\nb\nc", 2, "a\nb\n"},
		{"zero", "a\nb\n", 0, ""},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer
			if err := textproc.HeadLines(&buf, strings.NewReader(tt.input), tt.n); err != nil {
				t.Fatalf("HeadLines error = %v", err)
			}
			if buf.String() != tt.want {
				t.Errorf("HeadLines = %q, want %q", buf.String(), tt.want)
			}
		})
	}
}

func TestTailLines(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input string
		n     int
		want  string
	}{
		{"last two", "a\nb\nc\nd\n", 2, "c\nd\n"},
		{"more than available", "a\nb\n", 5, "a\nb\n"},
		{"no trailing newline", "a\nb\nc", 2, "b\nc"},
		{"zero", "a\nb\n", 0, ""},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer
			if err := textproc.TailLines(&buf, strings.NewReader(tt.input), tt.n); err != nil {
				t.Fatalf("TailLines error = %v", err)
			}
			if buf.String() != tt.want {
				t.Errorf("TailLines = %q, want %q", buf.String(), tt.want)
			}
		})
	}
}

func TestHeadBytes(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	if err := textproc.HeadBytes(&buf, strings.NewReader("hello world"), 5); err != nil {
		t.Fatalf("HeadBytes error = %v", err)
	}
	if buf.String() != "hello" {
		t.Errorf("HeadBytes = %q, want %q", buf.String(), "hello")
	}
}

func TestTailBytes(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	if err := textproc.TailBytes(&buf, strings.NewReader("hello world"), 5); err != nil {
		t.Fatalf("TailBytes error = %v", err)
	}
	if buf.String() != "world" {
		t.Errorf("TailBytes = %q, want %q", buf.String(), "world")
	}
}
