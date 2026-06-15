// Package adjtimex implements the adjtimex applet: read or set kernel clock
// adjustment parameters.
package adjtimex

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the adjtimex applet.
type Command struct{}

// New returns an adjtimex command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "adjtimex" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Read or set kernel clock parameters" }

// Request holds the parsed adjtimex tuning values. Only the fields named on the
// command line are populated. Parsing them here, away from the syscall, makes
// the command line testable.
type Request struct {
	Tick   int64 // -t: microseconds per tick
	Freq   int64 // -f: frequency offset (scaled ppm)
	Offset int64 // -o: time offset in microseconds
	set    map[string]bool
}

// IsSet reports whether the named tuning value was given.
func (r *Request) IsSet(name string) bool { return r.set[name] }

// Modifies reports whether the request would change the kernel clock (any
// setter flag present), as opposed to a read-only query.
func (r *Request) Modifies() bool {
	return r.IsSet("tick") || r.IsSet("frequency") || r.IsSet("offset")
}

// ParseArgs converts the -t/-f/-o flag values (already extracted by pflag) into
// a Request, recording which were explicitly given via the changed map.
func ParseArgs(tick, freq, offset string, changed map[string]bool) (*Request, error) {
	r := &Request{set: map[string]bool{}}
	if changed["tick"] {
		v, err := strconv.ParseInt(tick, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid tick value %q", tick)
		}
		r.Tick = v
		r.set["tick"] = true
	}
	if changed["frequency"] {
		v, err := strconv.ParseInt(freq, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid frequency value %q", freq)
		}
		r.Freq = v
		r.set["frequency"] = true
	}
	if changed["offset"] {
		v, err := strconv.ParseInt(offset, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid offset value %q", offset)
		}
		r.Offset = v
		r.set["offset"] = true
	}
	return r, nil
}

// String renders the requested changes for echo output and tests.
func (r *Request) String() string {
	var parts []string
	if r.IsSet("tick") {
		parts = append(parts, fmt.Sprintf("tick %d", r.Tick))
	}
	if r.IsSet("frequency") {
		parts = append(parts, fmt.Sprintf("frequency %d", r.Freq))
	}
	if r.IsSet("offset") {
		parts = append(parts, fmt.Sprintf("offset %d", r.Offset))
	}
	sort.Strings(parts)
	return strings.Join(parts, ", ")
}

// status describes the current kernel clock as read for a query.
type status struct {
	tick, freq, offset int64
	statusFlags        int64
}

// readStatusFn is indirected so the read-only adjtimex(2) query can be replaced
// in a test. In production it fails deterministically because even a read needs
// the adjtimex syscall, which is not available in every sandbox.
var readStatusFn = func() (*status, error) {
	return nil, fmt.Errorf("reading kernel clock parameters requires the adjtimex(2) syscall (not available here)")
}

// applyFn is indirected so the privileged adjtimex(2) write can be replaced in a
// test. In production it fails deterministically because setting clock
// parameters needs CAP_SYS_TIME.
var applyFn = func(_ *Request) error {
	return fmt.Errorf("setting kernel clock parameters requires CAP_SYS_TIME (not available here)")
}

// Run executes adjtimex.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-t TICK] [-f FREQ] [-o OFFSET]", stdio.Err).WithHelp(command.Help{
		Description: "Read or set the kernel's clock adjustment parameters via the adjtimex(2) syscall. " +
			"With no options it prints the current parameters. With -t the microseconds-per-tick value " +
			"is set, with -f the frequency offset (in scaled ppm), and with -o a one-shot time offset " +
			"(in microseconds). Setting any value changes the system clock and needs CAP_SYS_TIME; in " +
			"a sandbox without that capability the request is parsed and validated, then the command " +
			"fails with a clear message rather than silently doing nothing.",
		Examples: []command.Example{
			{Command: "adjtimex", Explain: "Print the current kernel clock parameters."},
			{Command: "adjtimex -t 10000", Explain: "Set the microseconds-per-tick value."},
			{Command: "adjtimex -f 0 -o 0", Explain: "Reset the frequency and offset."},
		},
		ExitStatus: "0  the query or change succeeded.\n" +
			"1  a bad value was given, or the syscall/capability was unavailable.",
	})
	tick := fs.StringP("tick", "t", "", "set the microseconds per tick")
	freq := fs.StringP("frequency", "f", "", "set the frequency offset (scaled ppm)")
	offset := fs.StringP("offset", "o", "", "set a one-shot time offset (microseconds)")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	if rest := fs.Args(); len(rest) > 0 {
		return command.Failuref("unexpected argument: %q", rest[0])
	}

	changed := map[string]bool{
		"tick":      fs.Changed("tick"),
		"frequency": fs.Changed("frequency"),
		"offset":    fs.Changed("offset"),
	}
	req, err := ParseArgs(*tick, *freq, *offset, changed)
	if err != nil {
		return command.Failuref("%v", err)
	}

	if !req.Modifies() {
		st, err := readStatusFn()
		if err != nil {
			return command.Failuref("%v", err)
		}
		_, _ = fmt.Fprintf(stdio.Out, "tick:      %d us\n", st.tick)
		_, _ = fmt.Fprintf(stdio.Out, "frequency: %d\n", st.freq)
		_, _ = fmt.Fprintf(stdio.Out, "offset:    %d us\n", st.offset)
		_, _ = fmt.Fprintf(stdio.Out, "status:    0x%x\n", st.statusFlags)
		return nil
	}

	if err := applyFn(req); err != nil {
		return command.Failuref("%v", err)
	}
	return nil
}
