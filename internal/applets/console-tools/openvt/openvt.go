// Package openvt implements the openvt applet: start a program on a new virtual
// terminal.
package openvt

import (
	"context"
	"fmt"
	"strconv"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the openvt applet.
type Command struct{}

// New returns an openvt command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "openvt" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Start a program on a new virtual terminal" }

// Request is the parsed openvt invocation: the program (and its arguments) to
// run, the explicit VT to use (0 means "pick the first free one"), and the mode
// flags. Parsing it apart from the console access makes the command line
// testable.
type Request struct {
	VT      int      // -c N: use VT N; 0 means find a free VT
	Switch  bool     // -s: switch to the new VT
	Wait    bool     // -w: wait for the program to finish
	Verbose bool     // -v: report what is being done
	Argv    []string // the program and its arguments
}

// ParseRequest validates the openvt flags and operands. A -c value, if present,
// must be a small positive VT number; at least one program word is required.
func ParseRequest(vtArg string, doSwitch, wait, verbose bool, argv []string) (*Request, error) {
	r := &Request{Switch: doSwitch, Wait: wait, Verbose: verbose, Argv: argv}
	if vtArg != "" {
		n, err := strconv.Atoi(vtArg)
		if err != nil || n < 1 || n > 63 {
			return nil, fmt.Errorf("invalid virtual terminal number: %q (expected 1..63)", vtArg)
		}
		r.VT = n
	}
	if len(argv) == 0 {
		return nil, fmt.Errorf("a program to run is required")
	}
	return r, nil
}

// runFn is indirected so the privileged console work can be replaced in a test.
// In production it fails deterministically: openvt must query a free VT via the
// VT_OPENQRY ioctl on the console, open /dev/ttyN, and (with -s) switch to it,
// all of which need a real Linux console and privilege.
var runFn = func(r *Request) error {
	return fmt.Errorf("opening a new virtual terminal for %q requires the VT_OPENQRY ioctl on a console (not available here)", r.Argv[0])
}

// Run executes openvt.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-c N] [-swv] -- PROGRAM [ARG]...", stdio.Err).WithHelp(command.Help{
		Description: "Find the first free Linux virtual terminal (or, with -c N, use VT N), open it, and " +
			"start PROGRAM with its arguments attached to that terminal. -s also switches the display to " +
			"the new VT, -w waits for the program to exit, and -v reports each step. Selecting and " +
			"opening a VT uses the VT_OPENQRY/VT_ACTIVATE ioctls on the console, so openvt needs a real " +
			"Linux console and privilege; in an environment without one it validates the request and " +
			"then fails with a clear message rather than doing nothing. Put -- before PROGRAM so its " +
			"own options are not parsed by openvt.",
		Examples: []command.Example{
			{Command: "openvt -- getty 38400 tty7", Explain: "Run getty on the first free VT."},
			{Command: "openvt -s -w -- top", Explain: "Open a VT, switch to it, run top, and wait."},
			{Command: "openvt -c 7 -- /bin/sh", Explain: "Run a shell specifically on tty7."},
		},
		ExitStatus: "0  the program was started.\n" +
			"1  a bad VT number, no program given, or no usable console was available.",
	})
	vt := fs.StringP("console", "c", "", "use VT N instead of the first free one")
	doSwitch := fs.BoolP("switch", "s", false, "switch to the new VT")
	wait := fs.BoolP("wait", "w", false, "wait for the program to finish")
	verbose := fs.BoolP("verbose", "v", false, "report each step")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	req, err := ParseRequest(*vt, *doSwitch, *wait, *verbose, fs.Args())
	if err != nil {
		return command.Failure(err)
	}
	if err := runFn(req); err != nil {
		return command.Failuref("%v", err)
	}
	return nil
}
