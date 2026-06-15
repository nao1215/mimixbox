// Package etherwake implements the ether-wake applet: send a Wake-on-LAN "magic
// packet" to a target MAC address. The magic-packet construction is pure and
// fully tested; the actual transmission goes through an injectable sender so
// tests never put a frame on the wire. Raw-socket (link-layer) transmission is
// intentionally deferred; this slice sends the magic packet as a UDP broadcast.
package etherwake

import (
	"context"
	"fmt"
	"net"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the ether-wake applet.
type Command struct{}

// New returns an ether-wake command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "ether-wake" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Send a Wake-on-LAN magic packet to a MAC address" }

// send transmits the magic packet payload to a broadcast destination. Tests
// replace it so no packet leaves the process.
var send = udpBroadcast

// Run executes ether-wake.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-b BROADCAST] MAC", stdio.Err).WithHelp(command.Help{
		Description: "Build a Wake-on-LAN magic packet for the target MAC address and send it. The " +
			"magic packet is the 6-byte sync stream FF:FF:FF:FF:FF:FF followed by the target MAC " +
			"repeated 16 times. This slice transmits the packet as a UDP broadcast (default " +
			"255.255.255.255:9), which needs no special privileges; raw link-layer (-i INTERFACE) " +
			"transmission is intentionally deferred and not yet implemented.",
		Examples: []command.Example{
			{Command: "ether-wake 11:22:33:44:55:66", Explain: "Wake a host by MAC via UDP broadcast."},
			{Command: "ether-wake -b 192.168.1.255 11:22:33:44:55:66", Explain: "Use a subnet broadcast."},
		},
		ExitStatus: "0  the magic packet was sent.\n" +
			"1  the MAC was invalid or the packet could not be sent.",
		Notes: []string{"Raw link-layer transmission (-i INTERFACE) is not implemented in this slice."},
	})
	broadcast := fs.StringP("broadcast", "b", "255.255.255.255", "broadcast address to send to")
	iface := fs.StringP("interface", "i", "", "send via raw socket on INTERFACE (not implemented)")
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	if *iface != "" {
		return command.Failuref("raw link-layer transmission via -i is not implemented in this slice")
	}

	operands := fs.Args()
	if len(operands) != 1 {
		return command.Failuref("exactly one target MAC address is required")
	}

	packet, err := magicPacket(operands[0])
	if err != nil {
		return command.Failuref("%v", err)
	}
	if err := send(*broadcast, packet); err != nil {
		return command.Failuref("cannot send magic packet: %v", err)
	}
	fmt.Fprintf(stdio.Out, "Sent magic packet to %s via %s\n", operands[0], *broadcast)
	return nil
}

// magicPacket builds the 102-byte Wake-on-LAN payload for mac: 6 bytes of 0xFF
// followed by the 6-byte MAC repeated 16 times.
func magicPacket(mac string) ([]byte, error) {
	hw, err := net.ParseMAC(mac)
	if err != nil {
		return nil, fmt.Errorf("invalid MAC address %q: %v", mac, err)
	}
	if len(hw) != 6 {
		return nil, fmt.Errorf("only 6-byte (EUI-48) MAC addresses are supported, got %d bytes", len(hw))
	}
	packet := make([]byte, 0, 6+16*6)
	for i := 0; i < 6; i++ {
		packet = append(packet, 0xFF)
	}
	for i := 0; i < 16; i++ {
		packet = append(packet, hw...)
	}
	return packet, nil
}

// udpBroadcast sends payload as a UDP datagram to broadcast:9.
func udpBroadcast(broadcast string, payload []byte) error {
	addr, err := net.ResolveUDPAddr("udp", net.JoinHostPort(broadcast, "9"))
	if err != nil {
		return err
	}
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()
	_, err = conn.Write(payload)
	return err
}
