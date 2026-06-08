// Package fortune implements the fortune applet: print a random, hopefully
// interesting, adage from a built-in collection.
package fortune

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the fortune applet.
type Command struct{}

// New returns a fortune command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "fortune" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Print a random, hopefully interesting, adage" }

// shortLimit is the longest a fortune may be to count as "short" (-s).
const shortLimit = 40

// fortunes is the built-in collection of adages.
var fortunes = []string{
	"The early bird gets the worm, but the second mouse gets the cheese.",
	"A journey of a thousand miles begins with a single step.",
	"Premature optimization is the root of all evil.",
	"There are only two hard things in computer science: cache invalidation and naming things.",
	"Simplicity is the ultimate sophistication.",
	"Talk is cheap. Show me the code.",
	"Weeks of coding can save you hours of planning.",
	"It works on my machine.",
	"Make it work, make it right, make it fast.",
	"The best way to predict the future is to invent it.",
	"Real programmers count from zero.",
	"To iterate is human, to recurse divine.",
}

// Run executes fortune.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]...", stdio.Err)
	short := fs.BoolP("short", "s", false, "only print short fortunes")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	pool := candidates(*short)
	if len(pool) == 0 {
		return command.Failuref("no fortunes available")
	}
	choice := pool[rand.Intn(len(pool))] //nolint:gosec // a fortune need not be cryptographically random
	if _, err := fmt.Fprintln(stdio.Out, choice); err != nil {
		return command.Failure(err)
	}
	return nil
}

// candidates returns the fortunes eligible to be printed, restricting to short
// ones when short is set.
func candidates(short bool) []string {
	if !short {
		return fortunes
	}
	var out []string
	for _, f := range fortunes {
		if len(f) <= shortLimit {
			out = append(out, f)
		}
	}
	return out
}
