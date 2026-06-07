// Package hostid implements the hostid applet: print the numeric identifier
// (in hexadecimal) of the current host. The computation is preserved from the
// original applet and, as its description warns, does not match the Coreutils
// hostid command.
package hostid

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
	mb "github.com/nao1215/mimixbox/internal/lib"
)

// Command is the hostid applet.
type Command struct{}

// New returns a hostid command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "hostid" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string {
	return "Print hostid (Host Identity Number, hex)!!!Does not work properly!!!"
}

// Run executes hostid.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]", stdio.Err)

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	ip4, err := mb.Ip4()
	if err != nil {
		fmt.Fprintf(stdio.Err, "hostid: %v\n", err)
		return command.SilentFailure()
	}

	// NOTE: The output doesn't match the Coreutils version of hostid command.
	// First, the IP address should be calculated from the hostname.
	// Next, the process of converting the IP address to hexadecimal does not
	// match. This computation is preserved from the original implementation.
	for _, ip := range ip4 {
		ipList := strings.Split(ip, ".")
		fmt.Fprintf(stdio.Out, "%02x%02x%02x%02x\n",
			atoi(ipList[1]), atoi(ipList[0]), atoi(ipList[3]), atoi(ipList[2]))
	}

	return nil
}

func atoi(decimal string) int {
	i, _ := strconv.Atoi(decimal)
	return i
}
