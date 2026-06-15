package i2c

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

// fakeBus is an in-memory I2C bus map keyed by address.
type fakeBus struct {
	regs    map[int]map[int]byte // addr -> reg -> value
	present []int
	err     error
	writes  []string
}

func (f *fakeBus) ReadReg(_, addr, reg int) (byte, error) {
	if f.err != nil {
		return 0, f.err
	}
	if reg < 0 {
		reg = 0
	}
	return f.regs[addr][reg], nil
}

func (f *fakeBus) WriteReg(_, addr, reg int, value byte) error {
	if f.err != nil {
		return f.err
	}
	f.writes = append(f.writes, formatWrite(addr, reg, value))
	return nil
}

func (f *fakeBus) Detect(_, lo, hi int) ([]int, error) {
	if f.err != nil {
		return nil, f.err
	}
	var out []int
	for _, a := range f.present {
		if a >= lo && a <= hi {
			out = append(out, a)
		}
	}
	return out, nil
}

func formatWrite(addr, reg int, v byte) string {
	var b strings.Builder
	b.WriteString("w")
	for _, n := range []int{addr, reg, int(v)} {
		b.WriteString(" ")
		b.WriteString(itoa(n))
	}
	return b.String()
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf []byte
	for n > 0 {
		buf = append([]byte{byte('0' + n%10)}, buf...)
		n /= 10
	}
	if neg {
		buf = append([]byte{'-'}, buf...)
	}
	return string(buf)
}

func withBus(t *testing.T, b Backend) {
	t.Helper()
	prev := busBackend
	busBackend = b
	t.Cleanup(func() { busBackend = prev })
}

func run(t *testing.T, c *Command, args ...string) (string, error) {
	t.Helper()
	var out, errBuf bytes.Buffer
	stdio := command.IO{In: strings.NewReader(""), Out: &out, Err: &errBuf}
	err := c.Run(context.Background(), stdio, args)
	return out.String(), err
}

func TestI2cgetReadsRegister(t *testing.T) {
	withBus(t, &fakeBus{regs: map[int]map[int]byte{0x50: {0x10: 0xAB}}})
	out, err := run(t, NewI2cget(), "-y", "1", "0x50", "0x10")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.TrimSpace(out) != "0xab" {
		t.Errorf("unexpected output: %q", out)
	}
}

func TestI2cgetBadAddress(t *testing.T) {
	withBus(t, &fakeBus{})
	if _, err := run(t, NewI2cget(), "-y", "1", "0x99", "0x10"); err == nil {
		t.Fatal("expected error for out-of-range chip address")
	}
}

func TestI2csetWrites(t *testing.T) {
	bus := &fakeBus{}
	withBus(t, bus)
	if _, err := run(t, NewI2cset(), "-y", "1", "0x50", "0x10", "0xff"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(bus.writes) != 1 || !strings.Contains(bus.writes[0], "80 16 255") {
		t.Errorf("write not planned: %v", bus.writes)
	}
}

func TestI2csetRejectsBigValue(t *testing.T) {
	withBus(t, &fakeBus{})
	if _, err := run(t, NewI2cset(), "-y", "1", "0x50", "0x10", "0x1ff"); err == nil {
		t.Fatal("expected error for value > 0xff")
	}
}

func TestI2cdetectGrid(t *testing.T) {
	withBus(t, &fakeBus{present: []int{0x50, 0x68}})
	out, err := run(t, NewI2cdetect(), "-y", "1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "50") || !strings.Contains(out, "68") {
		t.Errorf("detected addresses missing: %q", out)
	}
	if !strings.Contains(out, "--") {
		t.Errorf("absent markers missing: %q", out)
	}
}

func TestI2cdumpTable(t *testing.T) {
	regs := map[int]byte{}
	for r := 0; r < 256; r++ {
		regs[r] = byte(r)
	}
	withBus(t, &fakeBus{regs: map[int]map[int]byte{0x50: regs}})
	out, err := run(t, NewI2cdump(), "-y", "1", "0x50")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 17 { // header + 16 rows
		t.Errorf("expected 17 lines, got %d", len(lines))
	}
}

func TestI2cBackendError(t *testing.T) {
	withBus(t, &fakeBus{err: errors.New("permission denied")})
	if _, err := run(t, NewI2cget(), "-y", "1", "0x50", "0x10"); err == nil {
		t.Fatal("expected backend error to propagate")
	}
}

func TestI2cUsageErrors(t *testing.T) {
	withBus(t, &fakeBus{})
	if _, err := run(t, NewI2cdetect()); err == nil {
		t.Error("i2cdetect without bus should error")
	}
	if _, err := run(t, NewI2cset(), "-y", "1", "0x50"); err == nil {
		t.Error("i2cset with too few args should error")
	}
}
