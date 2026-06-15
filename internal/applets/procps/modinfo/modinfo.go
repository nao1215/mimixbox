// Package modinfo implements the modinfo applet: show metadata about a Linux
// kernel module (.ko file) by reading the ".modinfo" ELF section. Parsing is
// pure and hermetic; no kernel interaction is required.
package modinfo

import (
	"bytes"
	"context"
	"debug/elf"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the modinfo applet.
type Command struct{}

// New returns a modinfo command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "modinfo" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Show information about a kernel module" }

// Run executes modinfo.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-F FIELD] [-0] FILE...", stdio.Err).WithHelp(command.Help{
		Description: "Show metadata about each kernel module FILE by reading the .modinfo section of " +
			"the ELF object (a .ko file). Without options every field is printed as 'name: value'. " +
			"With -F only the values of the named field are printed, one per line. FILE must be a " +
			"path to a module file; module-name resolution from /lib/modules is not performed.",
		Examples: []command.Example{
			{Command: "modinfo ./loop.ko", Explain: "Print all metadata fields of loop.ko."},
			{Command: "modinfo -F license ./loop.ko", Explain: "Print only the license field."},
			{Command: "modinfo -0 -F depends ./loop.ko", Explain: "Print the depends field, NUL-separated."},
		},
		ExitStatus: "0  success.\n1  a file could not be read or is not a valid ELF module.",
		Notes: []string{
			"Operates on .ko files directly; it does not resolve bare module names against /lib/modules/$(uname -r).",
		},
	})
	field := fs.StringP("field", "F", "", "print only the value(s) of the named field")
	nul := fs.BoolP("null", "0", false, "use the NUL character instead of newline to separate field values")
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	files := fs.Args()
	if len(files) == 0 {
		_, _ = fmt.Fprintf(stdio.Err, "%s: no module file given\n", c.Name())
		return command.SilentFailure()
	}

	sep := byte('\n')
	if *nul {
		sep = 0
	}

	var failed bool
	for _, name := range files {
		pairs, err := readModinfo(name)
		if err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "%s: %s: %v\n", c.Name(), name, err)
			failed = true
			continue
		}
		if *field != "" {
			for _, p := range pairs {
				if p.key == *field {
					_, _ = fmt.Fprintf(stdio.Out, "%s%c", p.value, sep)
				}
			}
			continue
		}
		printAll(stdio, name, pairs)
	}
	if failed {
		return command.SilentFailure()
	}
	return nil
}

// printAll writes every field aligned in the classic "key:value" form, with the
// filename shown first.
func printAll(stdio command.IO, name string, pairs []pair) {
	width := len("filename")
	for _, p := range pairs {
		if len(p.key) > width {
			width = len(p.key)
		}
	}
	_, _ = fmt.Fprintf(stdio.Out, "%-*s %s\n", width+1, "filename:", name)
	for _, p := range pairs {
		_, _ = fmt.Fprintf(stdio.Out, "%-*s %s\n", width+1, p.key+":", p.value)
	}
}

// pair is one key/value entry from the .modinfo section.
type pair struct {
	key   string
	value string
}

// readModinfo opens a .ko file and decodes its ".modinfo" section into ordered
// key/value pairs. The section is a sequence of NUL-terminated "key=value"
// strings.
func readModinfo(name string) ([]pair, error) {
	f, err := os.Open(name) //nolint:gosec // user-named file
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	ef, err := elf.NewFile(f)
	if err != nil {
		return nil, fmt.Errorf("not an ELF object: %w", err)
	}
	sec := ef.Section(".modinfo")
	if sec == nil {
		return nil, fmt.Errorf("no .modinfo section (not a kernel module?)")
	}
	data, err := sec.Data()
	if err != nil {
		return nil, err
	}
	return parseModinfo(data), nil
}

// parseModinfo decodes NUL-separated "key=value" records, preserving their
// order. Entries without '=' are skipped.
func parseModinfo(data []byte) []pair {
	var pairs []pair
	for _, rec := range bytes.Split(data, []byte{0}) {
		if len(rec) == 0 {
			continue
		}
		s := string(rec)
		i := strings.IndexByte(s, '=')
		if i < 0 {
			continue
		}
		pairs = append(pairs, pair{key: s[:i], value: s[i+1:]})
	}
	return pairs
}

// fieldNames returns the distinct field names present, sorted; used by tests.
func fieldNames(pairs []pair) []string {
	seen := map[string]bool{}
	var out []string
	for _, p := range pairs {
		if !seen[p.key] {
			seen[p.key] = true
			out = append(out, p.key)
		}
	}
	sort.Strings(out)
	return out
}
