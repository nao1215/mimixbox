package tr

import (
	"reflect"
	"testing"
)

// TestExpandSetRangesAndEscapes drives expandSet through ranges, octal/C
// escapes, character classes, and the error paths.
func TestExpandSetRangesAndEscapes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		spec    string
		want    []rune
		wantErr bool
	}{
		{"plain", "abc", []rune{'a', 'b', 'c'}, false},
		{"range", "a-d", []rune{'a', 'b', 'c', 'd'}, false},
		{"digit range", "0-3", []rune{'0', '1', '2', '3'}, false},
		{"escape newline", `\n`, []rune{'\n'}, false},
		{"escape tab", `\t`, []rune{'\t'}, false},
		{"escape backslash", `\\`, []rune{'\\'}, false},
		{"escape cr", `\r`, []rune{'\r'}, false},
		{"escape bell", `\a`, []rune{'\a'}, false},
		{"octal", `\101`, []rune{'A'}, false},
		{"trailing backslash", `a\`, []rune{'a', '\\'}, false},
		{"unknown escape literal", `\q`, []rune{'q'}, false},
		{"class upper", "[:upper:]", rangeRunes('A', 'Z'), false},
		{"class blank", "[:blank:]", []rune{'\t', ' '}, false},
		{"reverse range err", "z-a", nil, true},
		{"unknown class err", "[:bogus:]", nil, true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := expandSet(tt.spec)
			if (err != nil) != tt.wantErr {
				t.Fatalf("expandSet(%q) err = %v, wantErr %v", tt.spec, err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("expandSet(%q) = %v, want %v", tt.spec, got, tt.want)
			}
		})
	}
}

// TestExpandSetClassNotAtRangeStart ensures a "[" that is not a class opener is
// treated as a literal character.
func TestExpandSetLiteralBracket(t *testing.T) {
	t.Parallel()
	got, err := expandSet("[x]")
	if err != nil {
		t.Fatalf("expandSet err = %v", err)
	}
	want := []rune{'[', 'x', ']'}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("expandSet = %v, want %v", got, want)
	}
}

// TestClassRunes covers every supported POSIX class and the unknown case.
func TestClassRunes(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"upper", "lower", "digit", "alpha", "alnum", "space", "blank"} {
		r, ok := classRunes(name)
		if !ok || len(r) == 0 {
			t.Errorf("classRunes(%q) = (%v, %v), want non-empty ok", name, r, ok)
		}
	}
	if _, ok := classRunes("bogus"); ok {
		t.Error("classRunes(bogus) ok = true, want false")
	}
}

// TestNextRune covers the escape decoding helper directly, including a
// three-digit octal sequence stopping at a non-octal digit.
func TestNextRune(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		src     string
		i       int
		wantR   rune
		wantIdx int
	}{
		{"plain", "abc", 1, 'b', 2},
		{"escape v", `\v`, 0, '\v', 2},
		{"escape f", `\f`, 0, '\f', 2},
		{"escape b", `\b`, 0, '\b', 2},
		{"octal three", `\101`, 0, 'A', 4},
		{"octal stops", `\18`, 0, '\001', 2},
		{"trailing backslash", `\`, 0, '\\', 1},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r, idx, err := nextRune([]rune(tt.src), tt.i)
			if err != nil {
				t.Fatalf("nextRune err = %v", err)
			}
			if r != tt.wantR || idx != tt.wantIdx {
				t.Errorf("nextRune(%q, %d) = (%q, %d), want (%q, %d)", tt.src, tt.i, r, idx, tt.wantR, tt.wantIdx)
			}
		})
	}
}

func TestSynopsis(t *testing.T) {
	t.Parallel()
	if New().Synopsis() == "" {
		t.Error("Synopsis() = empty")
	}
}
