package command

import "fmt"

// PlanGate is the reusable "validate, serialize a plan, then fail with a
// documented capability gate" flow shared by privileged applets (brctl,
// ifenslave, the SELinux mutators, the kernel-module mutators, and so on).
//
// Those applets all follow the same batch rule of "never ship a silent no-op":
// they validate their operands, serialize the requested action into a
// deterministic plan, report that plan on stdout, and then return a documented
// capability error rather than partially applying a privileged change. Holding
// that flow in one place keeps the user-visible "planned action" line and the
// "validate before gating" ordering consistent across the applets.
//
// Plan turns the parsed operands into the human- and test-readable plan string,
// or returns a validation error (which is reported before any plan line is
// printed). Gate maps a validated plan string to the documented capability
// error. Run wires it together; see (*FlagSet).Parse for the proceed/err
// contract.
type PlanGate struct {
	// Plan validates operands and returns the plan string to report.
	Plan func(operands []string) (string, error)
	// Gate returns the documented capability error for a validated plan.
	Gate func(plan string) error
}

// Run parses args with fs, then runs the plan-and-gate flow. On a parse stop
// (--help/--version or a parse error) it returns like (*FlagSet).Parse. On a
// validation error it returns that error without printing a plan. Otherwise it
// prints "<name>: planned action: <plan>\n" to stdout and returns g.Gate(plan).
func (g PlanGate) Run(fs *FlagSet, stdio IO, args []string) error {
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	plan, err := g.Plan(fs.Args())
	if err != nil {
		return Failure(err)
	}
	_, _ = fmt.Fprintf(stdio.Out, "%s: planned action: %s\n", fs.Name(), plan)
	return g.Gate(plan)
}
