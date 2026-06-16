package ifupdown

// NewIfup returns an ifup command. ifup parses the interfaces file, runs the
// configured pre-up/up hooks for the named interface, and then reports the
// capability-gated bring-up; the shared driver lives in backend.go.
func NewIfup() *Command { return &Command{name: "ifup"} }
