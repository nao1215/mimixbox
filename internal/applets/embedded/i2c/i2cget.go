package i2c

import (
	"fmt"

	"github.com/nao1215/mimixbox/internal/command"
)

// runGet implements i2cget: read one byte from a device register.
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
