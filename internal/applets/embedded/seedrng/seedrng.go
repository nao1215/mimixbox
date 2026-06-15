// Package seedrng implements the seedrng applet: carry random-number-generator
// seed entropy across reboots using a seed file.
package seedrng

import (
	"context"
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the seedrng applet.
type Command struct{}

// New returns a seedrng command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "seedrng" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Seed the RNG from a persistent seed file" }

// seedDir is the directory holding the persistent seed; overridable in tests.
var seedDir = "/var/lib/seedrng"

const seedFile = "seed"

// seeder feeds seed bytes into the kernel RNG. It is injected so the seed-file
// handling can be tested without crediting real kernel entropy.
var seeder Seeder = osSeeder{}

// Seeder abstracts the privileged "add entropy to the kernel RNG" operation.
type Seeder interface {
	// Credit feeds seed into the kernel pool. credit reports whether the
	// bytes should count toward the entropy estimate.
	Credit(seed []byte, credit bool) error
}

// Run executes seedrng.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-n]", stdio.Err).WithHelp(command.Help{
		Description: "Carry RNG seed entropy across reboots. On each run the existing seed file is read and " +
			"fed into the kernel random pool, then a fresh seed is drawn from the kernel and written back " +
			"for next boot. By default the restored seed credits the entropy estimate; -n adds it without " +
			"crediting. Feeding the kernel pool requires privilege; without it seedrng fails with a " +
			"documented error after the seed file has been refreshed.",
		Examples: []command.Example{
			{Command: "seedrng", Explain: "Restore and refresh the boot seed, crediting entropy."},
			{Command: "seedrng -n", Explain: "Restore the seed without crediting entropy."},
		},
		ExitStatus: "0  the seed was restored and refreshed.\n1  the seed file or the kernel pool was inaccessible.",
		Notes: []string{
			"The seed file lives under /var/lib/seedrng; the directory is created if missing.",
		},
	})
	noCredit := fs.BoolP("skip-credit", "n", false, "add the seed without crediting the entropy estimate")
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	path := filepath.Join(seedDir, seedFile)
	if err := os.MkdirAll(seedDir, 0o700); err != nil {
		return command.Failuref("%s", command.FileError(seedDir, err))
	}

	// Read the existing seed, if any, and feed it to the kernel.
	old, readErr := os.ReadFile(path) //nolint:gosec // fixed seed path under seedDir

	// Always refresh the seed file first so a privilege failure does not leave
	// the system without a fresh seed for next boot.
	fresh := make([]byte, 256)
	if _, err := rand.Read(fresh); err != nil {
		return command.Failuref("draw fresh seed: %v", err)
	}
	if err := os.WriteFile(path, fresh, 0o600); err != nil {
		return command.Failuref("%s", command.FileError(path, err))
	}

	if readErr != nil {
		if os.IsNotExist(readErr) {
			_, _ = fmt.Fprintln(stdio.Err, "seedrng: no existing seed; wrote a fresh one")
			return nil
		}
		return command.Failuref("%s", command.FileError(path, readErr))
	}
	if err := seeder.Credit(old, !*noCredit); err != nil {
		return command.Failuref("seed kernel RNG: %v", err)
	}
	return nil
}
