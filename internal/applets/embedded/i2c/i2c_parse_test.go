package i2c

import (
	"bytes"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func TestParseInt(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		in      string
		want    int
		wantErr bool
	}{
		{"hex", "0x50", 0x50, false},
		{"hex upper", "0XFF", 0xff, false},
		{"decimal", "42", 42, false},
		{"octal", "010", 8, false},
		{"zero", "0", 0, false},
		{"negative", "-1", -1, false},
		{"empty", "", 0, true},
		{"garbage", "xyz", 0, true},
		{"trailing", "0x50z", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := parseInt(tt.in)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("parseInt(%q) = %d, want error", tt.in, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseInt(%q) unexpected error: %v", tt.in, err)
			}
			if got != tt.want {
				t.Errorf("parseInt(%q) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}

func TestParseChipReg(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		args        []string
		regRequired bool
		wantBus     int
		wantAddr    int
		wantReg     int
		wantErr     bool
	}{
		{"bus addr no reg", []string{"1", "0x50"}, false, 1, 0x50, -1, false},
		{"bus addr with reg", []string{"1", "0x50", "0x10"}, false, 1, 0x50, 0x10, false},
		{"reg required present", []string{"1", "0x50", "0x10"}, true, 1, 0x50, 0x10, false},
		{"reg required missing", []string{"1", "0x50"}, true, 0, 0, 0, true},
		{"too few args", []string{"1"}, false, 0, 0, 0, true},
		{"too many args", []string{"1", "0x50", "0x10", "0x20"}, false, 0, 0, 0, true},
		{"invalid bus", []string{"bad", "0x50"}, false, 0, 0, 0, true},
		{"addr too low", []string{"1", "0x02"}, false, 0, 0, 0, true},
		{"addr too high", []string{"1", "0x78"}, false, 0, 0, 0, true},
		{"invalid reg", []string{"1", "0x50", "0x100"}, false, 0, 0, 0, true},
		{"negative reg", []string{"1", "0x50", "-1"}, false, 0, 0, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			bus, addr, reg, err := parseChipReg(tt.args, tt.regRequired)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("parseChipReg(%v, %v) = (%d,%d,%d), want error", tt.args, tt.regRequired, bus, addr, reg)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseChipReg(%v, %v) unexpected error: %v", tt.args, tt.regRequired, err)
			}
			if bus != tt.wantBus || addr != tt.wantAddr || reg != tt.wantReg {
				t.Errorf("parseChipReg(%v) = (%d,%d,%d), want (%d,%d,%d)", tt.args, bus, addr, reg, tt.wantBus, tt.wantAddr, tt.wantReg)
			}
		})
	}
}

// gridIO returns a command.IO writing to the returned buffer.
func gridIO() (command.IO, *bytes.Buffer) {
	out := &bytes.Buffer{}
	return command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}, out
}

func TestWriteDetectGrid(t *testing.T) {
	t.Parallel()
	stdio, out := gridIO()
	writeDetectGrid(stdio, []int{0x50, 0x68})
	got := out.String()

	if !strings.HasPrefix(got, "     0  1  2  3  4  5  6  7  8  9  a  b  c  d  e  f\n") {
		t.Errorf("missing/incorrect header: %q", got)
	}
	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
	if len(lines) != 9 { // header + 8 rows (0x00..0x70)
		t.Fatalf("expected 9 lines, got %d: %q", len(lines), got)
	}
	// Row 0x50 shows the detected address; 0x00-0x02 are out of range (blanks).
	row50 := lines[1+5] // header + row index 5 (0x50)
	if !strings.HasPrefix(row50, "50:") || !strings.Contains(row50, " 50") {
		t.Errorf("row 0x50 = %q, want detected 50", row50)
	}
	// Row 0x00 leaves 0x00-0x02 blank and marks the rest absent with --.
	row00 := lines[1]
	if !strings.HasPrefix(row00, "00:   ") {
		t.Errorf("row 0x00 = %q, want leading blanks for 0x00-0x02", row00)
	}
	if !strings.Contains(row00, " --") {
		t.Errorf("row 0x00 = %q, want -- absent markers", row00)
	}
}

func TestWriteDump(t *testing.T) {
	t.Parallel()
	var regs [256]byte
	for i := range regs {
		regs[i] = byte(i)
	}
	stdio, out := gridIO()
	writeDump(stdio, regs)
	got := out.String()

	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
	if len(lines) != 17 { // header + 16 rows
		t.Fatalf("expected 17 lines, got %d", len(lines))
	}
	if lines[0] != "     0  1  2  3  4  5  6  7  8  9  a  b  c  d  e  f" {
		t.Errorf("header = %q", lines[0])
	}
	// The first data row holds registers 0x00-0x0f, in order.
	if lines[1] != "00: 00 01 02 03 04 05 06 07 08 09 0a 0b 0c 0d 0e 0f" {
		t.Errorf("row 0 = %q", lines[1])
	}
	// The last data row holds registers 0xf0-0xff.
	if lines[16] != "f0: f0 f1 f2 f3 f4 f5 f6 f7 f8 f9 fa fb fc fd fe ff" {
		t.Errorf("row 15 = %q", lines[16])
	}
}

// TestInvalidOperandsAcrossApplets locks the error paths for bad bus, address,
// register, and value operands across all four front-ends. It is not parallel
// because it shares the package-global busBackend.
func TestInvalidOperandsAcrossApplets(t *testing.T) {
	withBus(t, &fakeBus{present: []int{0x50}, regs: map[int]map[int]byte{0x50: {}}})
	tests := []struct {
		name string
		cmd  func() *Command
		args []string
	}{
		{"detect bad bus", NewI2cdetect, []string{"-y", "bad"}},
		{"detect extra operand", NewI2cdetect, []string{"-y", "1", "2"}},
		{"get bad bus", NewI2cget, []string{"-y", "bad", "0x50", "0x10"}},
		{"get bad addr", NewI2cget, []string{"-y", "1", "0x99", "0x10"}},
		{"get bad reg", NewI2cget, []string{"-y", "1", "0x50", "0x100"}},
		{"set bad bus", NewI2cset, []string{"-y", "bad", "0x50", "0x10", "0x01"}},
		{"set bad addr", NewI2cset, []string{"-y", "1", "0x99", "0x10", "0x01"}},
		{"set bad reg", NewI2cset, []string{"-y", "1", "0x50", "0x100", "0x01"}},
		{"set bad value", NewI2cset, []string{"-y", "1", "0x50", "0x10", "0x1ff"}},
		{"set too few", NewI2cset, []string{"-y", "1", "0x50", "0x10"}},
		{"dump bad bus", NewI2cdump, []string{"-y", "bad", "0x50"}},
		{"dump bad addr", NewI2cdump, []string{"-y", "1", "bad"}},
		{"dump addr too high", NewI2cdump, []string{"-y", "1", "0x99"}},
		{"dump addr too low", NewI2cdump, []string{"-y", "1", "0x02"}},
		{"dump wrong arity", NewI2cdump, []string{"-y", "1"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := run(t, tt.cmd(), tt.args...); err == nil {
				t.Errorf("%s: expected error for args %v", tt.name, tt.args)
			}
		})
	}
}
