package probe

// NewTraceroute returns the traceroute applet (IPv4). It shares the spec table,
// target validation, and privileged transport defined in probe.go.
func NewTraceroute() *Command { return &Command{kind: kindTraceroute} }

// NewTraceroute6 returns the traceroute6 applet (IPv6). It shares the spec
// table, target validation, and privileged transport defined in probe.go.
func NewTraceroute6() *Command { return &Command{kind: kindTraceroute6} }
