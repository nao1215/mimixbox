package ipcmd

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func fixture() NetData {
	return NetData{
		Links: []Link{
			{Index: 1, Name: "lo", Flags: []string{"LOOPBACK", "UP", "LOWER_UP"}, MTU: 65536, State: "UNKNOWN",
				Addrs: []Addr{{Family: "inet", CIDR: "127.0.0.1/8", Scope: "host"}}},
			{Index: 2, Name: "eth0", Flags: []string{"BROADCAST", "MULTICAST", "UP", "LOWER_UP"}, MTU: 1500,
				MAC: "52:54:00:12:34:56", State: "UP",
				Addrs: []Addr{{Family: "inet", CIDR: "192.168.1.10/24", Scope: "global"}}},
		},
		Routes: []Route{
			{Dest: "default", Via: "192.168.1.1", Dev: "eth0", Proto: "static"},
			{Dest: "192.168.1.0/24", Dev: "eth0", Proto: "kernel", Scope: "link", Src: "192.168.1.10"},
		},
		Neighbours: []Neighbour{
			{IP: "192.168.1.1", Dev: "eth0", MAC: "52:54:00:aa:bb:cc", State: "REACHABLE"},
		},
		Rules: []Rule{
			{Priority: 0, Selector: "from all", Action: "lookup local"},
			{Priority: 32766, Selector: "from all", Action: "lookup main"},
		},
	}
}

func run(t *testing.T, cmd *Command, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := cmd.Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestIPAddrShow(t *testing.T) {
	defer SetSource(fixture())()
	out, _, err := run(t, NewIP(), "addr", "show")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	for _, want := range []string{
		"2: eth0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 state UP",
		"link/ether 52:54:00:12:34:56",
		"inet 192.168.1.10/24 scope global",
		"inet 127.0.0.1/8 scope host",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("addr show missing %q\ngot:\n%s", want, out)
		}
	}
}

func TestIPAddrAppletEquivalent(t *testing.T) {
	defer SetSource(fixture())()
	out, _, err := run(t, NewIPAddr(), "show")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(out, "inet 192.168.1.10/24") {
		t.Errorf("ipaddr show missing address: %s", out)
	}
}

func TestIPLinkShowDevFilter(t *testing.T) {
	defer SetSource(fixture())()
	out, _, err := run(t, NewIPLink(), "show", "dev", "eth0")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(out, "eth0") {
		t.Errorf("expected eth0: %s", out)
	}
	if strings.Contains(out, "lo:") {
		t.Errorf("dev filter should exclude lo: %s", out)
	}
	// link show should not print addresses.
	if strings.Contains(out, "inet ") {
		t.Errorf("link show should not include addresses: %s", out)
	}
}

func TestIPRouteShow(t *testing.T) {
	defer SetSource(fixture())()
	out, _, err := run(t, NewIPRoute(), "show")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(out, "default via 192.168.1.1 dev eth0 proto static") {
		t.Errorf("missing default route: %s", out)
	}
	if !strings.Contains(out, "192.168.1.0/24 dev eth0 proto kernel scope link src 192.168.1.10") {
		t.Errorf("missing connected route: %s", out)
	}
}

func TestIPNeighShow(t *testing.T) {
	defer SetSource(fixture())()
	out, _, err := run(t, NewIPNeigh(), "show")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(out, "192.168.1.1 dev eth0 lladdr 52:54:00:aa:bb:cc REACHABLE") {
		t.Errorf("missing neighbour: %s", out)
	}
}

func TestIPRuleShow(t *testing.T) {
	defer SetSource(fixture())()
	out, _, err := run(t, NewIPRule(), "show")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(out, "from all lookup local") || !strings.Contains(out, "32766:") {
		t.Errorf("missing rules: %s", out)
	}
}

func TestDefaultShowSubcommand(t *testing.T) {
	defer SetSource(fixture())()
	// "ip addr" with no subcommand defaults to show.
	out, _, err := run(t, NewIP(), "addr")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(out, "192.168.1.10/24") {
		t.Errorf("default-to-show failed: %s", out)
	}
}

func TestObjectPrefixMatch(t *testing.T) {
	defer SetSource(fixture())()
	// "ip a" should resolve to address.
	out, _, err := run(t, NewIP(), "a")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(out, "192.168.1.10/24") {
		t.Errorf("prefix 'a' should mean address: %s", out)
	}
}

func TestMutatingSubcommandRejected(t *testing.T) {
	defer SetSource(fixture())()
	_, _, err := run(t, NewIPAddr(), "add", "10.0.0.1/24", "dev", "eth0")
	if err == nil {
		t.Fatal("expected error for mutating subcommand")
	}
	if !strings.Contains(err.Error(), "mutating") {
		t.Errorf("err = %v, want mutating message", err)
	}
}

func TestUnknownObject(t *testing.T) {
	_, _, err := run(t, NewIP(), "bogus", "show")
	if err == nil {
		t.Fatal("expected error for unknown object")
	}
}

func TestMissingObject(t *testing.T) {
	_, _, err := run(t, NewIP())
	if err == nil {
		t.Fatal("expected error when OBJECT missing")
	}
}

func TestNamesAndSynopses(t *testing.T) {
	t.Parallel()
	cmds := map[string]*Command{
		"ip":      NewIP(),
		"ipaddr":  NewIPAddr(),
		"iplink":  NewIPLink(),
		"iproute": NewIPRoute(),
		"ipneigh": NewIPNeigh(),
		"iprule":  NewIPRule(),
	}
	for want, c := range cmds {
		if c.Name() != want {
			t.Errorf("Name() = %q, want %q", c.Name(), want)
		}
		if c.Synopsis() == "" {
			t.Errorf("%s Synopsis() is empty", want)
		}
	}
}
