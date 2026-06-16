package i2c

import (
	"github.com/nao1215/mimixbox/internal/command"
)

// runSet implements i2cset: write one byte to a device register.
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
