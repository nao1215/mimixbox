package i2c

import (
	"fmt"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// gridHeader is the column header shared by the detect grid and the dump table.
const gridHeader = "     0  1  2  3  4  5  6  7  8  9  a  b  c  d  e  f"

// writeDetectGrid renders the i2cdetect address grid.
func writeDetectGrid(stdio command.IO, found []int) {
	set := make(map[int]bool, len(found))
	for _, a := range found {
		set[a] = true
	}
	_, _ = fmt.Fprintln(stdio.Out, gridHeader)
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
	_, _ = fmt.Fprintln(stdio.Out, gridHeader)
	for row := 0; row < 16; row++ {
		var b strings.Builder
		fmt.Fprintf(&b, "%02x:", row*16)
		for col := 0; col < 16; col++ {
			fmt.Fprintf(&b, " %02x", regs[row*16+col])
		}
		_, _ = fmt.Fprintln(stdio.Out, b.String())
	}
}
