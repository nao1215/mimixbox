package i2c

import (
	"github.com/nao1215/mimixbox/internal/command"
)

// runDetect implements i2cdetect: scan a bus and print the address grid.
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
