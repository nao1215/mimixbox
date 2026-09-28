package printf_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/shellutils/printf"
)

// TestFormatEscapes drives the backslash escapes interpreted directly in the
// FORMAT string by formatEscape (including octal \NNN and hex \xHH), plus the
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
		{"octal escape", []string{`\101`}, "A"},        // 101 octal = 65 = 'A'
		{"octal one digit", []string{`\1x`}, "\x01x"},   // \1, then a literal x
		{"octal leading zero", []string{`\0101`}, "\b1"}, // \010, then a literal 1
		{"eight is not octal", []string{`\8`}, `\8`},
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
		{"octal without leading zero", []string{"%b", `\101`}, "A"},
		{"octal one digit", []string{"%b", `\1x`}, "\x01x"},
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

// encodeEscapes writes every byte of data as a printf escape (or a literal),
// choosing the form per byte from styles. inFormat selects the FORMAT dialect
// (\NNN only, and '%' written as %%) instead of the %b one (\0NNN or \NNN).
// Every form is written at full width so a following literal digit is never
// absorbed, and decoding must give data back.
func encodeEscapes(data, styles []byte, inFormat bool) string {
	named := map[byte]string{'\a': `\a`, '\b': `\b`, '\f': `\f`, '\n': `\n`, '\r': `\r`, '\t': `\t`, '\v': `\v`, '\\': `\\`}
	var b strings.Builder
	for i, d := range data {
		style := byte(0)
		if len(styles) > 0 {
			style = styles[i%len(styles)] % 5
		}
		switch {
		case inFormat && d == '%' && style == 0:
			b.WriteString("%%")
		case style == 0 && d != '\\' && d != '%':
			b.WriteByte(d)
		case style == 2 && (inFormat || d >= 0o100):
			fmt.Fprintf(&b, `\%03o`, d)
		case style == 3:
			fmt.Fprintf(&b, `\x%02x`, d)
		case style == 4 && named[d] != "":
			b.WriteString(named[d])
		case inFormat:
			fmt.Fprintf(&b, `\%03o`, d)
		default:
			fmt.Fprintf(&b, `\0%03o`, d)
		}
	}
	return b.String()
}

// FuzzEscapes checks the escapes printf interprets in its FORMAT and in %b
// arguments. Arbitrary input must never panic; a FORMAT without '\' or '%'
// prints unchanged; %b never grows its argument. Bytes encoded with the
// octal, hex and named escapes GNU printf accepts in each place must decode
// back to themselves.
func FuzzEscapes(f *testing.F) {
	f.Add(`a\tb%%\101\0101\1x\8`, []byte("A\x00\\%\n\xff7"), []byte{0, 1, 2, 3, 4})
	f.Add(`\x9z\xz\c%b%5s%`, []byte("0101"), []byte{2})
	f.Add(`end\`, []byte{}, []byte{})
	f.Fuzz(func(t *testing.T, raw string, data, styles []byte) {
		out, _, err := run(t, raw)
		if err != nil {
			t.Fatalf("Run(%q) error = %v", raw, err)
		}
		if !strings.ContainsAny(raw, `\%`) && out != raw {
			t.Fatalf("Run(%q) = %q, want the format unchanged", raw, out)
		}
		out, _, err = run(t, "%b", raw)
		if err != nil {
			t.Fatalf("Run(%%b, %q) error = %v", raw, err)
		}
		if len(out) > len(raw) {
			t.Fatalf("Run(%%b, %q) = %q grew the argument", raw, out)
		}
		if !strings.Contains(raw, `\`) && out != raw {
			t.Fatalf("Run(%%b, %q) = %q, want the argument unchanged", raw, out)
		}

		if enc := encodeEscapes(data, styles, true); len(enc) > 0 {
			out, _, err = run(t, enc)
			if err != nil {
				t.Fatalf("Run(%q) error = %v", enc, err)
			}
			if out != string(data) {
				t.Fatalf("Run(%q) = %q, want %q", enc, out, data)
			}
		}
		enc := encodeEscapes(data, styles, false)
		out, _, err = run(t, "%b", enc)
		if err != nil {
			t.Fatalf("Run(%%b, %q) error = %v", enc, err)
		}
		if out != string(data) {
			t.Fatalf("Run(%%b, %q) = %q, want %q", enc, out, data)
		}
	})
}
