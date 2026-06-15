// Package i2c implements the BusyBox I2C applets (i2cdetect, i2cget, i2cset,
// i2cdump, i2ctransfer) over a shared, injectable bus backend so the argument
// parsing and command planning are unit testable without real I2C hardware.
package i2c

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Backend abstracts an I2C bus so commands can be planned and tested without
// touching /dev/i2c-*.
type Backend interface {
	// ReadReg reads one byte from the device at addr on the given bus,
	// either from a specific register (reg >= 0) or the current pointer
	// (reg < 0).
	ReadReg(bus, addr, reg int) (byte, error)
	// WriteReg writes one byte to register reg of the device at addr.
	WriteReg(bus, addr, reg int, value byte) error
	// Detect probes addresses lo..hi on the bus and returns those that ACK.
	Detect(bus, lo, hi int) ([]int, error)
}

// busBackend is the active backend; tests inject a fake.
var busBackend Backend = osBackend{}

// Command is one of the I2C applets, distinguished by name.
type Command struct{ name string }

// NewI2cdetect returns the i2cdetect applet.
func NewI2cdetect() *Command { return &Command{name: "i2cdetect"} }

// NewI2cget returns the i2cget applet.
func NewI2cget() *Command { return &Command{name: "i2cget"} }

// NewI2cset returns the i2cset applet.
func NewI2cset() *Command { return &Command{name: "i2cset"} }

// NewI2cdump returns the i2cdump applet.
func NewI2cdump() *Command { return &Command{name: "i2cdump"} }

// Name returns the command name.
func (c *Command) Name() string { return c.name }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string {
	switch c.name {
	case "i2cdetect":
		return "Detect I2C chips on a bus"
	case "i2cget":
		return "Read a byte from an I2C device"
	case "i2cset":
		return "Write a byte to an I2C device"
	case "i2cdump":
		return "Dump the registers of an I2C device"
	}
	return "I2C bus tool"
}

// Run dispatches to the per-command implementation.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	switch c.name {
	case "i2cdetect":
		return c.runDetect(stdio, args)
	case "i2cget":
		return c.runGet(stdio, args)
	case "i2cset":
		return c.runSet(stdio, args)
	case "i2cdump":
		return c.runDump(stdio, args)
	}
	return command.Failuref("unknown i2c command %q", c.name)
}

// runDetect implements i2cdetect.
func (c *Command) runDetect(stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.name, "[-y] BUS", stdio.Err).WithHelp(command.Help{
		Description: "Scan I2C BUS (a bus number, e.g. 1 for /dev/i2c-1) and print a grid of the 7-bit " +
			"addresses 0x03-0x77 that respond. Probing reads from devices and is normally safe, but on " +
			"some chips a quick-write probe can have side effects. Access needs privilege; without it the " +
			"command fails with a documented error.",
		Examples: []command.Example{
			{Command: "i2cdetect -y 1", Explain: "Scan bus 1 without the interactive confirmation."},
		},
		ExitStatus: "0  the scan completed.\n1  bad arguments or the bus could not be opened.",
	})
	_ = fs.BoolP("yes", "y", false, "skip the interactive confirmation")
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	rest := fs.Args()
	if len(rest) != 1 {
		return command.Failuref("usage: i2cdetect [-y] BUS")
	}
	bus, err := parseInt(rest[0])
	if err != nil {
		return command.Failuref("invalid bus %q", rest[0])
	}
	found, err := busBackend.Detect(bus, 0x03, 0x77)
	if err != nil {
		return command.Failuref("%v", err)
	}
	writeDetectGrid(stdio, found)
	return nil
}

// runGet implements i2cget.
func (c *Command) runGet(stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.name, "[-y] BUS CHIP-ADDR [REG]", stdio.Err).WithHelp(command.Help{
		Description: "Read one byte from the I2C device at CHIP-ADDR on BUS. With REG the byte is read from " +
			"that register; without it the byte is read from the current pointer. Values may be given in " +
			"hex (0x..), decimal, or octal. The result is printed in hex. Access needs privilege.",
		Examples: []command.Example{
			{Command: "i2cget -y 1 0x50 0x00", Explain: "Read register 0x00 of chip 0x50 on bus 1."},
		},
		ExitStatus: "0  the read succeeded.\n1  bad arguments or the read failed.",
	})
	_ = fs.BoolP("yes", "y", false, "skip the interactive confirmation")
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	bus, addr, reg, err := parseChipReg(fs.Args(), false)
	if err != nil {
		return command.Failuref("%v", err)
	}
	v, err := busBackend.ReadReg(bus, addr, reg)
	if err != nil {
		return command.Failuref("%v", err)
	}
	_, _ = fmt.Fprintf(stdio.Out, "0x%02x\n", v)
	return nil
}

// runSet implements i2cset.
func (c *Command) runSet(stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.name, "[-y] BUS CHIP-ADDR REG VALUE", stdio.Err).WithHelp(command.Help{
		Description: "Write one byte VALUE to register REG of the I2C device at CHIP-ADDR on BUS. WARNING: " +
			"writing to a device register can change hardware state irreversibly. Access needs privilege; " +
			"without it the command fails with a documented error.",
		Examples: []command.Example{
			{Command: "i2cset -y 1 0x50 0x10 0xff", Explain: "Write 0xff to register 0x10 of chip 0x50 (destructive)."},
		},
		ExitStatus: "0  the write succeeded.\n1  bad arguments or the write failed.",
	})
	_ = fs.BoolP("yes", "y", false, "skip the interactive confirmation")
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	rest := fs.Args()
	if len(rest) != 4 {
		return command.Failuref("usage: i2cset [-y] BUS CHIP-ADDR REG VALUE")
	}
	bus, addr, reg, err := parseChipReg(rest[:3], true)
	if err != nil {
		return command.Failuref("%v", err)
	}
	val, err := parseInt(rest[3])
	if err != nil || val < 0 || val > 0xff {
		return command.Failuref("invalid byte value %q", rest[3])
	}
	if err := busBackend.WriteReg(bus, addr, reg, byte(val)); err != nil {
		return command.Failuref("%v", err)
	}
	return nil
}

// runDump implements i2cdump by reading every register 0x00-0xff.
func (c *Command) runDump(stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.name, "[-y] BUS CHIP-ADDR", stdio.Err).WithHelp(command.Help{
		Description: "Read and print all 256 registers (0x00-0xff) of the I2C device at CHIP-ADDR on BUS as " +
			"a hex table. Each register is read individually. Access needs privilege; without it the " +
			"command fails with a documented error.",
		Examples: []command.Example{
			{Command: "i2cdump -y 1 0x50", Explain: "Dump every register of chip 0x50 on bus 1."},
		},
		ExitStatus: "0  the dump completed.\n1  bad arguments or a read failed.",
	})
	_ = fs.BoolP("yes", "y", false, "skip the interactive confirmation")
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	rest := fs.Args()
	if len(rest) != 2 {
		return command.Failuref("usage: i2cdump [-y] BUS CHIP-ADDR")
	}
	bus, err := parseInt(rest[0])
	if err != nil {
		return command.Failuref("invalid bus %q", rest[0])
	}
	addr, err := parseInt(rest[1])
	if err != nil {
		return command.Failuref("invalid chip address %q", rest[1])
	}
	var regs [256]byte
	for reg := 0; reg < 256; reg++ {
		v, err := busBackend.ReadReg(bus, addr, reg)
		if err != nil {
			return command.Failuref("%v", err)
		}
		regs[reg] = v
	}
	writeDump(stdio, regs)
	return nil
}

// parseChipReg parses BUS CHIP-ADDR [REG]. When regRequired is true REG must be
// present; otherwise a missing REG yields -1 (current pointer).
func parseChipReg(args []string, regRequired bool) (bus, addr, reg int, err error) {
	if len(args) < 2 || len(args) > 3 {
		return 0, 0, 0, fmt.Errorf("expected BUS CHIP-ADDR [REG]")
	}
	bus, err = parseInt(args[0])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid bus %q", args[0])
	}
	addr, err = parseInt(args[1])
	if err != nil || addr < 0x03 || addr > 0x77 {
		return 0, 0, 0, fmt.Errorf("invalid chip address %q (want 0x03-0x77)", args[1])
	}
	reg = -1
	if len(args) == 3 {
		reg, err = parseInt(args[2])
		if err != nil || reg < 0 || reg > 0xff {
			return 0, 0, 0, fmt.Errorf("invalid register %q", args[2])
		}
	} else if regRequired {
		return 0, 0, 0, fmt.Errorf("a register operand is required")
	}
	return bus, addr, reg, nil
}

// parseInt parses a C-style signed integer (0x hex, 0 octal, else decimal).
func parseInt(s string) (int, error) {
	n, err := strconv.ParseInt(s, 0, 32)
	return int(n), err
}

// writeDetectGrid renders the i2cdetect address grid.
func writeDetectGrid(stdio command.IO, found []int) {
	set := make(map[int]bool, len(found))
	for _, a := range found {
		set[a] = true
	}
	_, _ = fmt.Fprintln(stdio.Out, "     0  1  2  3  4  5  6  7  8  9  a  b  c  d  e  f")
	for row := 0; row < 8; row++ {
		var b strings.Builder
		fmt.Fprintf(&b, "%02x:", row*16)
		for col := 0; col < 16; col++ {
			addr := row*16 + col
			switch {
			case addr < 0x03 || addr > 0x77:
				b.WriteString("   ")
			case set[addr]:
				fmt.Fprintf(&b, " %02x", addr)
			default:
				b.WriteString(" --")
			}
		}
		_, _ = fmt.Fprintln(stdio.Out, b.String())
	}
}

// writeDump renders the 256-register hex table.
func writeDump(stdio command.IO, regs [256]byte) {
	_, _ = fmt.Fprintln(stdio.Out, "     0  1  2  3  4  5  6  7  8  9  a  b  c  d  e  f")
	for row := 0; row < 16; row++ {
		var b strings.Builder
		fmt.Fprintf(&b, "%02x:", row*16)
		for col := 0; col < 16; col++ {
			fmt.Fprintf(&b, " %02x", regs[row*16+col])
		}
		_, _ = fmt.Fprintln(stdio.Out, b.String())
	}
}
