package probe

// NewArping returns the arping applet (local-link ARP probe). It shares the spec
// table, target validation, and privileged transport defined in probe.go.
func NewArping() *Command { return &Command{kind: kindArping} }
