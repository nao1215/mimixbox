package i2c

import (
	"fmt"
	"strconv"
)

// parseChipReg parses BUS CHIP-ADDR [REG]. When regRequired is true REG must be
// present; otherwise a missing REG yields -1 (current pointer). It is shared by
// the i2cget and i2cset front-ends.
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
