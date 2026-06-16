package ifupdown

import "github.com/nao1215/mimixbox/internal/command"

// NewIfplugd returns an ifplugd command.
func NewIfplugd() *Command { return &Command{name: "ifplugd"} }

// runIfplugd validates the requested interface and reports the capability-gated
// link-state monitoring backend (privileged netlink access is unavailable here).
func runIfplugd(stdio command.IO, args []string) error {
	fs := command.NewFlagSet("ifplugd", "-i IFACE", stdio.Err).WithHelp(command.Help{
		Description: "Monitor an interface's link state and run ifup/ifdown when the cable is plugged or " +
			"unplugged. Link-state monitoring relies on privileged netlink/ioctl access that is not " +
			"available in this environment; this slice validates its arguments and reports a documented " +
			"capability error instead of silently doing nothing.",
		Examples: []command.Example{
			{Command: "ifplugd -i eth0", Explain: "Plan link monitoring of eth0 (capability-gated)."},
		},
		ExitStatus: "0  never in this environment.\n1  always: validated request then a documented backend error.",
	})
	iface := fs.StringP("iface", "i", "", "interface to monitor")
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	if *iface == "" {
		return command.Failuref("an interface is required (-i)")
	}
	return command.Failuref(
		"ifplugd: link-state monitoring of %q requires privileged netlink access not available in this "+
			"environment (capability-gated backend)", *iface)
}
