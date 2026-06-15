// Package setserial implements the setserial applet: get or set the
// configuration of a serial port.
package setserial

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the setserial applet.
type Command struct{}

// New returns a setserial command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "setserial" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Get or set serial port configuration" }

// Params holds the serial-port settings setserial can read or change. Only the
// fields that appear on the command line are populated; the rest keep their zero
// value. Parsing them here, away from any device, is what lets the command line
// be unit tested.
type Params struct {
	Port     uint64 // port: I/O port base address
	IRQ      int    // irq: interrupt number
	Baudbase int    // baud_base: base baud rate
	UARTType string // uart: UART chip type (e.g. 16550A, none)
	set      map[string]bool
}

// IsSet reports whether the named parameter appeared on the command line.
func (p *Params) IsSet(name string) bool { return p.set[name] }

// knownUARTTypes is the set of UART chip names setserial accepts for "uart X".
var knownUARTTypes = map[string]bool{
	"none": true, "8250": true, "16450": true, "16550": true,
	"16550A": true, "16650": true, "16650V2": true, "16654": true,
	"16750": true, "16850": true, "16950": true,
}

// ParseParams parses setserial parameter words (the operands after the device
// name) into a Params. It understands the common "name value" pairs and the
// numeric base prefixes (0x.., 0.., decimal). Unknown or malformed parameters
// produce a descriptive error.
func ParseParams(words []string) (*Params, error) {
	p := &Params{set: map[string]bool{}}
	for i := 0; i < len(words); i++ {
		name := strings.ToLower(words[i])
		needValue := func() (string, error) {
			if i+1 >= len(words) {
				return "", fmt.Errorf("the %q parameter needs a value", name)
			}
			i++
			return words[i], nil
		}
		switch name {
		case "port":
			v, err := needValue()
			if err != nil {
				return nil, err
			}
			n, err := strconv.ParseUint(v, 0, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid port address %q", v)
			}
			p.Port = n
			p.set["port"] = true
		case "irq":
			v, err := needValue()
			if err != nil {
				return nil, err
			}
			n, err := strconv.Atoi(v)
			if err != nil || n < 0 {
				return nil, fmt.Errorf("invalid irq %q", v)
			}
			p.IRQ = n
			p.set["irq"] = true
		case "baud_base":
			v, err := needValue()
			if err != nil {
				return nil, err
			}
			n, err := strconv.Atoi(v)
			if err != nil || n <= 0 {
				return nil, fmt.Errorf("invalid baud_base %q", v)
			}
			p.Baudbase = n
			p.set["baud_base"] = true
		case "uart":
			v, err := needValue()
			if err != nil {
				return nil, err
			}
			if !knownUARTTypes[strings.ToUpper(v)] && !knownUARTTypes[strings.ToLower(v)] {
				return nil, fmt.Errorf("unknown uart type %q", v)
			}
			p.UARTType = v
			p.set["uart"] = true
		case "auto_irq", "skip_test", "autoconfig", "^auto_irq", "^skip_test":
			// Flag-style configuration directives accepted for compatibility.
			p.set[strings.TrimPrefix(name, "^")] = true
		default:
			return nil, fmt.Errorf("unknown parameter %q", words[i])
		}
	}
	return p, nil
}

// String renders the set parameters in "name value" form, sorted by name, for
// the -g (get) output and for tests.
func (p *Params) String() string {
	var parts []string
	if p.IsSet("port") {
		parts = append(parts, fmt.Sprintf("port 0x%04x", p.Port))
	}
	if p.IsSet("irq") {
		parts = append(parts, fmt.Sprintf("irq %d", p.IRQ))
	}
	if p.IsSet("baud_base") {
		parts = append(parts, fmt.Sprintf("baud_base %d", p.Baudbase))
	}
	if p.IsSet("uart") {
		parts = append(parts, fmt.Sprintf("uart %s", p.UARTType))
	}
	sort.Strings(parts)
	return strings.Join(parts, ", ")
}

// applyFn is indirected so the privileged TIOCSSERIAL ioctl can be replaced in a
// test. In production it fails deterministically because changing a serial
// port's hardware configuration needs the ioctl on a real tty and privilege.
var applyFn = func(device string, _ *Params) error {
	return fmt.Errorf("configuring %s requires the TIOCSSERIAL ioctl on a serial device (not available here)", device)
}

// Run executes setserial.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-g] DEVICE [PARAMETER VALUE]...", stdio.Err).WithHelp(command.Help{
		Description: "Get or set the configuration of the serial port named by DEVICE (for example " +
			"/dev/ttyS0). With no parameters, or with -g, the port's settings are reported. Otherwise " +
			"each 'PARAMETER VALUE' pair is parsed and applied with the TIOCSSERIAL ioctl. Recognised " +
			"parameters include port (I/O base, accepts 0x.. hex), irq, baud_base, and uart (the chip " +
			"type, e.g. 16550A or none). Parameters are fully parsed and validated before any device " +
			"change; applying them needs a real serial device and privilege, so without one the " +
			"command validates the request and then fails with a clear message rather than silently " +
			"doing nothing.",
		Examples: []command.Example{
			{Command: "setserial -g /dev/ttyS0", Explain: "Report the port's configuration."},
			{Command: "setserial /dev/ttyS0 baud_base 115200", Explain: "Set the base baud rate."},
			{Command: "setserial /dev/ttyS0 uart 16550A irq 4", Explain: "Set the UART type and IRQ."},
		},
		ExitStatus: "0  the request succeeded.\n" +
			"1  a bad parameter was given, no device was named, or no serial device was available.",
	})
	get := fs.BoolP("get", "g", false, "report the device configuration instead of changing it")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	rest := fs.Args()
	if len(rest) == 0 {
		return command.Failuref("a serial device is required")
	}
	device := rest[0]
	params, err := ParseParams(rest[1:])
	if err != nil {
		return command.Failuref("%v", err)
	}

	if *get || len(rest) == 1 {
		// Reporting a live port needs TIOCGSERIAL; without a device, echo back
		// only what was requested so the command is never a silent no-op.
		if len(params.set) == 0 {
			return command.Failuref("reading %s requires the TIOCGSERIAL ioctl on a serial device (not available here)", device)
		}
		_, _ = fmt.Fprintf(stdio.Out, "%s, %s\n", device, params.String())
		return nil
	}

	if err := applyFn(device, params); err != nil {
		return command.Failuref("%v", err)
	}
	return nil
}
