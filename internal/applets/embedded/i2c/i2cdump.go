package i2c

import (
	"github.com/nao1215/mimixbox/internal/command"
)

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
