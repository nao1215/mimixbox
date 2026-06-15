package setserial

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, args []string) (string, string, error) {
	t.Helper()
	var out, errBuf bytes.Buffer
	io := command.IO{In: strings.NewReader(""), Out: &out, Err: &errBuf}
	err := New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestParseParams(t *testing.T) {
	t.Parallel()
	p, err := ParseParams([]string{"port", "0x3f8", "irq", "4", "baud_base", "115200", "uart", "16550A"})
	if err != nil {
		t.Fatalf("ParseParams: %v", err)
	}
	if p.Port != 0x3f8 {
		t.Errorf("Port = %#x", p.Port)
	}
	if p.IRQ != 4 {
		t.Errorf("IRQ = %d", p.IRQ)
	}
	if p.Baudbase != 115200 {
		t.Errorf("Baudbase = %d", p.Baudbase)
	}
	if p.UARTType != "16550A" {
		t.Errorf("UARTType = %q", p.UARTType)
	}
	if !p.IsSet("port") || !p.IsSet("uart") {
		t.Error("IsSet flags not recorded")
	}
}

func TestParseParamsErrors(t *testing.T) {
	t.Parallel()
	cases := [][]string{
		{"port"},               // missing value
		{"port", "nothex"},     // bad number
		{"irq", "-1"},          // negative
		{"baud_base", "0"},     // non-positive
		{"uart", "99999"},      // unknown uart
		{"bogus", "1"},         // unknown parameter
	}
	for _, words := range cases {
		if _, err := ParseParams(words); err == nil {
			t.Errorf("ParseParams(%v) expected error", words)
		}
	}
}

func TestParamsString(t *testing.T) {
	t.Parallel()
	p, _ := ParseParams([]string{"port", "0x3f8", "irq", "4"})
	got := p.String()
	if !strings.Contains(got, "port 0x03f8") || !strings.Contains(got, "irq 4") {
		t.Errorf("String = %q", got)
	}
}

func TestRunNoDevice(t *testing.T) {
	t.Parallel()
	if _, _, err := run(t, nil); err == nil {
		t.Error("expected error when no device given")
	}
}

func TestRunBadParam(t *testing.T) {
	t.Parallel()
	if _, _, err := run(t, []string{"/dev/ttyS0", "bogus", "1"}); err == nil {
		t.Error("expected error for bad parameter")
	}
}

func TestRunGetEchoesRequest(t *testing.T) {
	t.Parallel()
	// "-g" with explicit params echoes the parsed request rather than hitting a
	// device, so it is a deterministic non-no-op success path.
	out, _, err := run(t, []string{"-g", "/dev/ttyS0", "baud_base", "115200"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "/dev/ttyS0") || !strings.Contains(out, "baud_base 115200") {
		t.Errorf("get output = %q", out)
	}
}

func TestRunGetNoParamsCapabilityError(t *testing.T) {
	t.Parallel()
	if _, _, err := run(t, []string{"-g", "/dev/ttyS0"}); err == nil {
		t.Error("expected capability error reading a live device")
	}
}

func TestRunSetCapabilityError(t *testing.T) {
	t.Parallel()
	if _, _, err := run(t, []string{"/dev/ttyS0", "baud_base", "115200"}); err == nil {
		t.Error("expected capability error applying to a live device")
	}
}

// TestRunSetInjected exercises the apply path through the injected transport so
// the device name and parsed parameters are passed through without touching a
// serial port.
func TestRunSetInjected(t *testing.T) {
	orig := applyFn
	var gotDevice string
	var gotParams *Params
	applyFn = func(device string, p *Params) error {
		gotDevice = device
		gotParams = p
		return nil
	}
	defer func() { applyFn = orig }()

	if _, _, err := run(t, []string{"/dev/ttyS0", "baud_base", "115200", "uart", "16550A"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotDevice != "/dev/ttyS0" {
		t.Errorf("device = %q", gotDevice)
	}
	if gotParams == nil || gotParams.Baudbase != 115200 || gotParams.UARTType != "16550A" {
		t.Errorf("params not passed through: %+v", gotParams)
	}
}
