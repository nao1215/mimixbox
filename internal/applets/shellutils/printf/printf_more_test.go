package printf_test

import (
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/shellutils/printf"
)

// TestFormatEscapes drives the backslash escapes interpreted directly in the
// FORMAT string by formatEscape (including octal \0NNN and hex \xHH), plus the
// "unknown escape" fall-through that emits a literal backslash. Outputs match
// GNU printf.
func TestFormatEscapes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"bell", []string{`\a`}, "\a"},
		{"backspace", []string{`\b`}, "\b"},
		{"formfeed", []string{`\f`}, "\f"},
		{"carriage return", []string{`\r`}, "\r"},
		{"vertical tab", []string{`\v`}, "\v"},
		{"literal backslash", []string{`\\`}, "\\"},
		{"octal escape", []string{`\0101`}, "A"},      // 0101 octal = 65 = 'A'
		{"bare null escape", []string{`\0`}, "\x00"},  // \0 with no digits = NUL
		{"hex escape", []string{`\x41`}, "A"},         // 0x41 = 'A'
		{"hex single digit", []string{`\x9z`}, "\tz"}, // \x9 = tab, then literal z
		{"unknown escape literal", []string{`\q`}, `\q`},
		{"trailing backslash literal", []string{`\`}, `\`},
		{"bad hex emits literal", []string{`\xz`}, `\xz`},
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

// TestPercentBEscapes drives expandEscapes via the %b conversion, covering every
// escape it understands plus the \c "stop output" escape and octal/hex forms.
func TestPercentBEscapes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"newline", []string{"%b", `a\nb`}, "a\nb"},
		{"tab", []string{"%b", `a\tb`}, "a\tb"},
		{"bell", []string{"%b", `\a`}, "\a"},
		{"backspace", []string{"%b", `\b`}, "\b"},
		{"formfeed", []string{"%b", `\f`}, "\f"},
		{"carriage return", []string{"%b", `\r`}, "\r"},
		{"vertical tab", []string{"%b", `\v`}, "\v"},
		{"backslash", []string{"%b", `\\`}, "\\"},
		{"octal", []string{"%b", `\0101`}, "A"},
		{"hex", []string{"%b", `\x41`}, "A"},
		{"bad hex keeps literal", []string{"%b", `\xz`}, `\xz`},
		{"unknown escape literal", []string{"%b", `\q`}, `\q`},
		{"c stops output", []string{"%b", `ab\ccd`}, "ab"},
		{"trailing backslash literal", []string{"%b", `a\`}, `a\`},
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

// TestUnsignedConversions covers toUint's two parse paths: an unsigned literal
// and the signed fallback for a negative argument (reinterpreted as unsigned).
func TestUnsignedConversions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"hex of negative reinterpreted", []string{"%x", "-1"}, "ffffffffffffffff"},
		{"upper hex", []string{"%X", "255"}, "FF"},
		{"unsigned decimal", []string{"%u", "42"}, "42"},
		{"unsigned of empty is zero", []string{"%u"}, "0"},
		{"unsigned of garbage is zero", []string{"%u", "notnum"}, "0"},
		{"hex literal input", []string{"%d", "0x1f"}, "31"},
		{"decimal garbage is zero", []string{"%d", "nope"}, "0"},
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

// TestConversionEdgeCases covers the trailing-'%'-with-no-verb path, an unknown
// verb emitted literally, and an empty %c argument.
func TestConversionEdgeCases(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"trailing percent no verb", []string{"abc%"}, "abc%"},
		{"unknown verb literal", []string{"%q", "x"}, "%q"},
		{"empty c arg produces nothing", []string{"[%c]"}, "[]"},
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

// TestSynopsis covers the Synopsis accessor.
func TestSynopsis(t *testing.T) {
	t.Parallel()
	c := printf.New()
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
	if c.Name() != "printf" {
		t.Errorf("Name() = %q", c.Name())
	}
}
