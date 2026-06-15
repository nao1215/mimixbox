package devmem

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

// fakeBackend records accesses and returns canned read values.
type fakeBackend struct {
	readVal   uint64
	readErr   error
	writeErr  error
	lastRead  *plan
	lastWrite *plan
}

func (f *fakeBackend) Read(addr uint64, width int) (uint64, error) {
	f.lastRead = &plan{addr: addr, width: width}
	return f.readVal, f.readErr
}

func (f *fakeBackend) Write(addr uint64, width int, value uint64) error {
	f.lastWrite = &plan{addr: addr, width: width, value: value, write: true}
	return f.writeErr
}

func withBackend(t *testing.T, b Backend) {
	t.Helper()
	prev := memBackend
	memBackend = b
	t.Cleanup(func() { memBackend = prev })
}

func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	var out, errBuf bytes.Buffer
	stdio := command.IO{In: strings.NewReader(""), Out: &out, Err: &errBuf}
	err := New().Run(context.Background(), stdio, args)
	return out.String(), err
}

func TestParsePlanRead(t *testing.T) {
	p, err := parsePlan([]string{"0x1000"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.addr != 0x1000 || p.width != 4 || p.write {
		t.Errorf("unexpected plan: %+v", p)
	}
}

func TestParsePlanWidths(t *testing.T) {
	for bits, width := range map[string]int{"8": 1, "16": 2, "32": 4, "64": 8} {
		p, err := parsePlan([]string{"0x10", bits})
		if err != nil {
			t.Fatalf("width %s: %v", bits, err)
		}
		if p.width != width {
			t.Errorf("bits %s => width %d, want %d", bits, p.width, width)
		}
	}
}

func TestParsePlanWrite(t *testing.T) {
	p, err := parsePlan([]string{"0x10", "32", "0xdeadbeef"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !p.write || p.value != 0xdeadbeef {
		t.Errorf("unexpected plan: %+v", p)
	}
}

func TestParsePlanErrors(t *testing.T) {
	cases := [][]string{
		{},
		{"a", "b", "c", "d"},
		{"notanaddr"},
		{"0x10", "7"},       // bad width
		{"0x10", "8", "0x1FF"}, // value too wide for 8 bits
	}
	for _, args := range cases {
		if _, err := parsePlan(args); err == nil {
			t.Errorf("expected error for %v", args)
		}
	}
}

func TestRunRead(t *testing.T) {
	fake := &fakeBackend{readVal: 0xABCD}
	withBackend(t, fake)
	out, err := run(t, "0x2000", "32")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.TrimSpace(out) != "0x0000ABCD" {
		t.Errorf("unexpected output: %q", out)
	}
	if fake.lastRead.addr != 0x2000 || fake.lastRead.width != 4 {
		t.Errorf("read not planned correctly: %+v", fake.lastRead)
	}
}

func TestRunWrite(t *testing.T) {
	fake := &fakeBackend{}
	withBackend(t, fake)
	if _, err := run(t, "0x2000", "16", "0x42"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fake.lastWrite == nil || fake.lastWrite.value != 0x42 || fake.lastWrite.width != 2 {
		t.Errorf("write not planned correctly: %+v", fake.lastWrite)
	}
}

func TestRunBackendError(t *testing.T) {
	withBackend(t, &fakeBackend{readErr: errors.New("permission denied")})
	if _, err := run(t, "0x10"); err == nil {
		t.Fatal("expected error when backend fails")
	}
}
