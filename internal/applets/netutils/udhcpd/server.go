package udhcpd

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Config is a parsed udhcpd configuration.
type Config struct {
	Interface string
	Start     net.IP
	End       net.IP
	ServerID  net.IP
	SubnetMask net.IP
	Router    net.IP
	DNS       net.IP
	LeaseSec  uint32
}

// ParseConfig parses the udhcpd config grammar from r. Recognised keys: start,
// end, interface, opt/option (subnet, router, dns), server_id, and lease.
// Unknown keys are ignored to stay forward-compatible; blank lines and '#'
// comments are skipped.
func ParseConfig(r io.Reader) (*Config, error) {
	cfg := &Config{LeaseSec: 86400}
	sc := bufio.NewScanner(r)
	line := 0
	for sc.Scan() {
		line++
		text := strings.TrimSpace(sc.Text())
		if text == "" || strings.HasPrefix(text, "#") {
			continue
		}
		f := strings.Fields(text)
		key := strings.ToLower(f[0])
		val := ""
		if len(f) > 1 {
			val = f[1]
		}
		switch key {
		case "interface":
			cfg.Interface = val
		case "start":
			if cfg.Start = net.ParseIP(val).To4(); cfg.Start == nil {
				return nil, fmt.Errorf("line %d: invalid start IP %q", line, val)
			}
		case "end":
			if cfg.End = net.ParseIP(val).To4(); cfg.End == nil {
				return nil, fmt.Errorf("line %d: invalid end IP %q", line, val)
			}
		case "server_id", "siaddr":
			cfg.ServerID = net.ParseIP(val).To4()
		case "lease":
			n, err := strconv.ParseUint(val, 10, 32)
			if err != nil {
				return nil, fmt.Errorf("line %d: invalid lease %q", line, val)
			}
			cfg.LeaseSec = uint32(n)
		case "opt", "option":
			if len(f) < 3 {
				return nil, fmt.Errorf("line %d: option needs a name and value", line)
			}
			if err := cfg.applyOption(strings.ToLower(f[1]), f[2]); err != nil {
				return nil, fmt.Errorf("line %d: %v", line, err)
			}
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	if cfg.Start == nil || cfg.End == nil {
		return nil, fmt.Errorf("config must set both 'start' and 'end'")
	}
	return cfg, nil
}

func (c *Config) applyOption(name, val string) error {
	ip := net.ParseIP(val).To4()
	switch name {
	case "subnet":
		c.SubnetMask = ip
	case "router":
		c.Router = ip
	case "dns":
		c.DNS = ip
	default:
		// Ignore unknown options for forward compatibility.
	}
	return nil
}

// Transport sends and receives DHCP packets. Injecting it lets tests drive the
// server without a real UDP socket.
type Transport interface {
	Recv() (packet []byte, from net.Addr, err error)
	Send(packet []byte, to net.Addr) error
}

// Allocator hands out IPs from the configured pool.
type Allocator struct {
	next net.IP
	end  net.IP
}

// NewAllocator returns an Allocator over [start,end].
func NewAllocator(start, end net.IP) *Allocator {
	return &Allocator{next: append(net.IP{}, start.To4()...), end: end.To4()}
}

// Allocate returns the next free IP, or nil when the pool is exhausted.
func (a *Allocator) Allocate() net.IP {
	if a.next == nil || ipGreater(a.next, a.end) {
		return nil
	}
	ip := append(net.IP{}, a.next...)
	a.next = incIP(a.next)
	return ip
}

// HandlePacket builds the reply for one received DHCP request given cfg and an
// allocator. It returns the reply message and the IP it allocated (if any). A
// nil reply means the packet should be ignored.
func HandlePacket(req *Message, cfg *Config, alloc *Allocator) *Message {
	switch req.Type() {
	case Discover:
		ip := alloc.Allocate()
		if ip == nil {
			return nil
		}
		return buildReply(req, cfg, ip, Offer)
	case Request:
		ip := req.YIAddr
		if v, ok := req.Options[OptRequestedIP]; ok && len(v) == 4 {
			ip = net.IP(v)
		}
		if ip == nil || ip.To4() == nil {
			ip = alloc.Allocate()
		}
		if ip == nil {
			return buildReply(req, cfg, nil, NAK)
		}
		return buildReply(req, cfg, ip.To4(), ACK)
	default:
		return nil
	}
}

func buildReply(req *Message, cfg *Config, ip net.IP, msgType byte) *Message {
	opts := map[byte][]byte{OptMessageType: {msgType}}
	if cfg.ServerID != nil {
		opts[OptServerID] = cfg.ServerID
	}
	if msgType != NAK {
		opts[OptLeaseTime] = LeaseTimeOption(cfg.LeaseSec)
		if cfg.SubnetMask != nil {
			opts[OptSubnetMask] = cfg.SubnetMask
		}
		if cfg.Router != nil {
			opts[OptRouter] = cfg.Router
		}
		if cfg.DNS != nil {
			opts[OptDNSServer] = cfg.DNS
		}
	}
	return &Message{
		Op:      OpBootReply,
		XID:     req.XID,
		YIAddr:  ip,
		SIAddr:  cfg.ServerID,
		GIAddr:  req.GIAddr,
		CHAddr:  req.CHAddr,
		Options: opts,
	}
}

// ServeOnce reads one packet from t, processes it, and (when a reply is built)
// sends the reply back. It returns io.EOF when the transport is exhausted.
func ServeOnce(cfg *Config, alloc *Allocator, t Transport) error {
	packet, from, err := t.Recv()
	if err != nil {
		return err
	}
	req, err := Unmarshal(packet)
	if err != nil {
		return nil // ignore malformed packets
	}
	reply := HandlePacket(req, cfg, alloc)
	if reply == nil {
		return nil
	}
	return t.Send(reply.Marshal(), from)
}

// Serve loops ServeOnce until the transport returns io.EOF or ctx is cancelled.
func Serve(ctx context.Context, cfg *Config, t Transport) error {
	alloc := NewAllocator(cfg.Start, cfg.End)
	for {
		if ctx.Err() != nil {
			return nil
		}
		if err := ServeOnce(cfg, alloc, t); err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
	}
}

func ipGreater(a, b net.IP) bool {
	a, b = a.To4(), b.To4()
	for i := 0; i < 4; i++ {
		if a[i] != b[i] {
			return a[i] > b[i]
		}
	}
	return false
}

func incIP(ip net.IP) net.IP {
	out := append(net.IP{}, ip.To4()...)
	for i := 3; i >= 0; i-- {
		out[i]++
		if out[i] != 0 {
			break
		}
	}
	return out
}

// udpTransport is the production Transport backed by a UDP socket.
type udpTransport struct{ pc net.PacketConn }

func (u *udpTransport) Recv() ([]byte, net.Addr, error) {
	buf := make([]byte, 1500)
	n, addr, err := u.pc.ReadFrom(buf)
	if err != nil {
		return nil, nil, err
	}
	return buf[:n], addr, nil
}

func (u *udpTransport) Send(p []byte, to net.Addr) error {
	_, err := u.pc.WriteTo(p, to)
	return err
}

// runServer wires a real UDP socket and serves until ctx is cancelled.
func runServer(ctx context.Context, cfg *Config, addr string, out io.Writer) error {
	pc, err := net.ListenPacket("udp", addr)
	if err != nil {
		return command.Failuref("cannot listen on %s: %v", addr, err)
	}
	go func() {
		<-ctx.Done()
		_ = pc.Close()
	}()
	_, _ = fmt.Fprintf(out, "udhcpd: serving %s-%s on %s\n", cfg.Start, cfg.End, pc.LocalAddr())
	err = Serve(ctx, cfg, &udpTransport{pc: pc})
	if err != nil && ctx.Err() == nil {
		return command.Failure(err)
	}
	return nil
}
