// Package devmem implements the devmem applet: read or write a physical memory
// address through /dev/mem.
package devmem

import (
	"context"
	"fmt"
	"strconv"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the devmem applet.
type Command struct{}

// New returns a devmem command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "devmem" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Read or write physical memory" }

// memBackend performs the actual /dev/mem access. Tests inject a fake so the
// argument parsing and access plan can be checked without touching real memory.
var memBackend Backend = osBackend{}

// Backend abstracts physical-memory access so command planning is testable.
type Backend interface {
	// Read returns width bytes at the physical address.
	Read(addr uint64, width int) (uint64, error)
	// Write stores value (width bytes) at the physical address.
	Write(addr uint64, width int, value uint64) error
}

// plan is the parsed access request.
type plan struct {
	addr  uint64
	width int // bytes: 1, 2, 4, or 8
	write bool
	value uint64
}

// Run executes devmem.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "ADDRESS [WIDTH [VALUE]]", stdio.Err).WithHelp(command.Help{
		Description: "Read or write a value at physical memory ADDRESS through /dev/mem. ADDRESS is parsed " +
			"as a C-style integer (0x.. hex, 0.. octal, else decimal). WIDTH is the access width in bits: " +
			"8, 16, 32 (default), or 64. With no VALUE the address is read and printed in hex; with VALUE " +
			"the value is written. WARNING: writing to physical memory can corrupt the running system or " +
			"hardware. Access requires privilege; without it devmem fails with a documented error.",
		Examples: []command.Example{
			{Command: "devmem 0x10000000", Explain: "Read a 32-bit word at 0x10000000."},
			{Command: "devmem 0x10000000 8", Explain: "Read a single byte."},
			{Command: "devmem 0x10000000 32 0xdeadbeef", Explain: "Write a 32-bit value (destructive)."},
		},
		ExitStatus: "0  the access succeeded.\n1  bad arguments or the access was denied.",
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	p, err := parsePlan(fs.Args())
	if err != nil {
		return command.Failuref("%v", err)
	}

	if p.write {
		if err := memBackend.Write(p.addr, p.width, p.value); err != nil {
			return command.Failuref("%v", err)
		}
		return nil
	}
	v, err := memBackend.Read(p.addr, p.width)
	if err != nil {
		return command.Failuref("%v", err)
	}
	_, _ = fmt.Fprintf(stdio.Out, "0x%0*X\n", p.width*2, v)
	return nil
}

// parsePlan validates the operands into an access plan.
func parsePlan(args []string) (plan, error) {
	if len(args) < 1 || len(args) > 3 {
		return plan{}, fmt.Errorf("usage: devmem ADDRESS [WIDTH [VALUE]]")
	}
	addr, err := parseUint(args[0])
	if err != nil {
		return plan{}, fmt.Errorf("invalid address %q", args[0])
	}
	p := plan{addr: addr, width: 4}
	if len(args) >= 2 {
		bits, err := strconv.Atoi(args[1])
		if err != nil {
			return plan{}, fmt.Errorf("invalid width %q", args[1])
		}
		switch bits {
		case 8, 16, 32, 64:
			p.width = bits / 8
		default:
			return plan{}, fmt.Errorf("width must be 8, 16, 32, or 64, got %d", bits)
		}
	}
	if len(args) == 3 {
		val, err := parseUint(args[2])
		if err != nil {
			return plan{}, fmt.Errorf("invalid value %q", args[2])
		}
		if p.width < 8 && val >= uint64(1)<<(uint(p.width)*8) {
			return plan{}, fmt.Errorf("value 0x%X does not fit in %d bits", val, p.width*8)
		}
		p.write = true
		p.value = val
	}
	return p, nil
}

// parseUint parses a C-style unsigned integer (0x hex, 0 octal, else decimal).
func parseUint(s string) (uint64, error) {
	return strconv.ParseUint(s, 0, 64)
}
