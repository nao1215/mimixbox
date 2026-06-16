package ifupdown

// NewIfdown returns an ifdown command. ifdown parses the interfaces file, runs
// the configured down/post-down hooks for the named interface, and then reports
// the capability-gated take-down; the shared driver lives in backend.go.
func NewIfdown() *Command { return &Command{name: "ifdown"} }
