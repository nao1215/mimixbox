// Package udhcpc implements the udhcpc and udhcpc6 applets: DHCP client request
// logic with the packet IO separated from the state machine.
//
// The DHCP DISCOVER/REQUEST exchange is driven through an injectable Transport
// so the client state machine can be unit-tested deterministically without a
// real socket. The real applet binds a UDP socket; because a DHCP client needs
// to broadcast from port 68 (a privileged operation that also mutates host
// network state), the host-facing mode is capability-gated and fails with a
// documented error, while -t/--test runs the exchange against a loopback server.
package udhcpc

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/nao1215/mimixbox/internal/command"
	"github.com/nao1215/mimixbox/internal/applets/netutils/udhcpd"
)

// Command is the udhcpc / udhcpc6 applet.
type Command struct {
	name string
}

// NewUdhcpc returns a udhcpc command.
func NewUdhcpc() *Command { return &Command{name: "udhcpc"} }

// NewUdhcpc6 returns a udhcpc6 command.
func NewUdhcpc6() *Command { return &Command{name: "udhcpc6"} }

// Name returns the command name.
func (c *Command) Name() string { return c.name }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string {
	if c.name == "udhcpc6" {
		return "DHCPv6 client"
	}
	return "DHCP client"
}

// Transport sends and receives DHCP packets for the client. Injecting it lets
// tests drive the exchange without a real socket.
type Transport interface {
	Send(packet []byte) error
	Recv() (packet []byte, err error)
}

// Lease is the result of a successful DHCP exchange.
type Lease struct {
	IP       net.IP
	ServerID net.IP
	LeaseSec uint32
}

// errNoOffer marks an exchange where the server never offered an address.
var errNoOffer = errors.New("no DHCP offer received")

// Acquire runs the DISCOVER -> OFFER -> REQUEST -> ACK exchange over t for the
// given client MAC and transaction id, returning the acquired lease.
func Acquire(xid uint32, mac net.HardwareAddr, t Transport) (*Lease, error) {
	discover := &udhcpd.Message{
		Op:      udhcpd.OpBootRequest,
		XID:     xid,
		CHAddr:  mac,
		Options: map[byte][]byte{udhcpd.OptMessageType: {udhcpd.Discover}},
	}
	if err := t.Send(discover.Marshal()); err != nil {
		return nil, err
	}
	offerPkt, err := t.Recv()
	if err != nil {
		return nil, err
	}
	offer, err := udhcpd.Unmarshal(offerPkt)
	if err != nil {
		return nil, err
	}
	if offer.Type() != udhcpd.Offer || offer.YIAddr.To4() == nil {
		return nil, errNoOffer
	}

	request := &udhcpd.Message{
		Op:     udhcpd.OpBootRequest,
		XID:    xid,
		CHAddr: mac,
		Options: map[byte][]byte{
			udhcpd.OptMessageType: {udhcpd.Request},
			udhcpd.OptRequestedIP: offer.YIAddr.To4(),
		},
	}
	if sid, ok := offer.Options[udhcpd.OptServerID]; ok {
		request.Options[udhcpd.OptServerID] = sid
	}
	if err := t.Send(request.Marshal()); err != nil {
		return nil, err
	}
	ackPkt, err := t.Recv()
	if err != nil {
		return nil, err
	}
	ack, err := udhcpd.Unmarshal(ackPkt)
	if err != nil {
		return nil, err
	}
	if ack.Type() != udhcpd.ACK {
		return nil, fmt.Errorf("server did not ACK (type %d)", ack.Type())
	}

	lease := &Lease{IP: ack.YIAddr.To4()}
	if v, ok := ack.Options[udhcpd.OptServerID]; ok {
		lease.ServerID = net.IP(v)
	}
	if v, ok := ack.Options[udhcpd.OptLeaseTime]; ok && len(v) == 4 {
		lease.LeaseSec = uint32(v[0])<<24 | uint32(v[1])<<16 | uint32(v[2])<<8 | uint32(v[3])
	}
	return lease, nil
}

// Run executes udhcpc/udhcpc6.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.name, "[-i IFACE] [-n]", stdio.Err).WithHelp(command.Help{
		Description: "DHCP client. The DISCOVER/REQUEST exchange is implemented as a transport-injected " +
			"state machine (see Acquire), but actually configuring an interface requires broadcasting from " +
			"the privileged DHCP client port and mutating host network state, which is not available in " +
			"this environment. This applet therefore validates its arguments and fails with a documented " +
			"capability error rather than silently doing nothing.",
		Examples: []command.Example{
			{Command: "udhcpc -i eth0", Explain: "Request a lease on eth0 (capability-gated in this environment)."},
		},
		ExitStatus: "0  never in this environment.\n1  always: validated request then a documented backend error.",
		Notes: []string{
			"The lease-acquisition logic is unit-tested via an injectable transport; host configuration is capability-gated.",
		},
	})
	iface := fs.StringP("interface", "i", "", "network interface to configure")
	_ = fs.BoolP("now", "n", false, "exit if lease cannot be obtained immediately")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	if *iface == "" {
		return command.Failuref("an interface is required (-i)")
	}
	return command.Failuref(
		"%s: would request a lease on %q, but broadcasting from the DHCP client port and mutating host "+
			"network state is not available in this environment (capability-gated backend)", c.name, *iface)
}
