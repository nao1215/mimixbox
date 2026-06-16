package probe

// NewPing6 returns the ping6 applet (ICMPv6 echo). It shares the spec table,
// target validation, and privileged transport defined in probe.go.
func NewPing6() *Command { return &Command{kind: kindPing6} }
