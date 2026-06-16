// Package i2c implements the BusyBox I2C applets (i2cdetect, i2cget, i2cset,
// i2cdump) over a shared, injectable bus backend so the argument parsing and
// command planning are unit testable without real I2C hardware.
//
// The package is split so that backend-independent parsing (parse.go) and
// rendering (render.go) live apart from the per-applet front-ends
// (i2cdetect.go, i2cget.go, i2cset.go, i2cdump.go); this file holds only the
// shared Command type, its constructors, and the name-based dispatch tables.
package i2c

import (
	"context"

	"github.com/nao1215/mimixbox/internal/command"
)

// command name constants for the multi-applet package.
const (
	cmdI2cdetect = "i2cdetect"
	cmdI2cget    = "i2cget"
	cmdI2cset    = "i2cset"
	cmdI2cdump   = "i2cdump"
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
func NewI2cdetect() *Command { return &Command{name: cmdI2cdetect} }

// NewI2cget returns the i2cget applet.
func NewI2cget() *Command { return &Command{name: cmdI2cget} }

// NewI2cset returns the i2cset applet.
func NewI2cset() *Command { return &Command{name: cmdI2cset} }

// NewI2cdump returns the i2cdump applet.
func NewI2cdump() *Command { return &Command{name: cmdI2cdump} }

// Name returns the command name.
func (c *Command) Name() string { return c.name }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string {
	switch c.name {
	case cmdI2cdetect:
		return "Detect I2C chips on a bus"
	case cmdI2cget:
		return "Read a byte from an I2C device"
	case cmdI2cset:
		return "Write a byte to an I2C device"
	case cmdI2cdump:
		return "Dump the registers of an I2C device"
	}
	return "I2C bus tool"
}

// Run dispatches to the per-command implementation.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	switch c.name {
	case cmdI2cdetect:
		return c.runDetect(stdio, args)
	case cmdI2cget:
		return c.runGet(stdio, args)
	case cmdI2cset:
		return c.runSet(stdio, args)
	case cmdI2cdump:
		return c.runDump(stdio, args)
	}
	return command.Failuref("unknown i2c command %q", c.name)
}
