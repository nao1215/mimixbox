// Package arp implements the arp applet in read-only inspection mode: it prints
// the kernel ARP/neighbour cache from an injectable data source. Adding and
// deleting static entries is intentionally deferred and reported as a documented
// capability error.
package arp

import (
	"context"
	"fmt"
	"text/tabwriter"

	"github.com/nao1215/mimixbox/internal/applets/netutils/ipcmd"
	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the arp applet.
type Command struct{}

// New returns an arp command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "arp" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Show the ARP/neighbour cache (read-only)" }

// source supplies neighbour fixtures; tests replace it via SetSource.
var source = func() []ipcmd.Neighbour { return nil }

// SetSource installs fixture neighbours for a test and returns a restore func.
func SetSource(ns []ipcmd.Neighbour) (restore func()) {
	orig := source
	source = func() []ipcmd.Neighbour { return ns }
	return func() { source = orig }
}

// Run executes arp.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-n] [-a]", stdio.Err).WithHelp(command.Help{
		Description: "Print the kernel ARP cache (the IPv4 address-to-MAC neighbour table) in the " +
			"traditional columnar format (Address, HWtype, HWaddress, Flags, Iface). This slice is " +
			"read-only: adding (-s) or deleting (-d) entries is intentionally deferred and reported " +
			"as an error.",
		Examples: []command.Example{
			{Command: "arp -n", Explain: "Print the ARP cache numerically."},
		},
		ExitStatus: "0  the ARP cache was printed.\n" +
			"1  an add/delete operation was requested.",
		Notes: []string{"Adding or deleting ARP entries is not implemented in this slice."},
	})
	_ = fs.BoolP("numeric", "n", false, "show numerical addresses (always on in this slice)")
	_ = fs.BoolP("all", "a", false, "BSD-style output (accepted, table format unchanged)")
	del := fs.StringP("delete", "d", "", "delete an entry (not implemented)")
	set := fs.StringP("set", "s", "", "add a static entry (not implemented)")
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	if *del != "" || *set != "" {
		return command.Failuref("adding/deleting ARP entries is not implemented in this read-only slice")
	}

	writeTable(stdio.Out, source())
	return nil
}

// writeTable renders the ARP cache in "arp -n" style.
func writeTable(w interface{ Write([]byte) (int, error) }, ns []ipcmd.Neighbour) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "Address\tHWtype\tHWaddress\tFlags Mask\tIface")
	for _, n := range ns {
		hw, flags := "ether", "C"
		if n.MAC == "" {
			hw, flags = "", "(incomplete)"
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n", n.IP, hw, n.MAC, flags, n.Dev)
	}
	_ = tw.Flush()
}
