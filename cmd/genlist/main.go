// Command genlist regenerates the command (applet) list in README.md from the
// single source of truth: the applet table in internal/applets. Each applet's
// own Synopsis() supplies its description, so the documented list can never
// drift from the registered commands.
//
// Run it with `make command-list` (or `go run ./cmd/genlist`) after adding or
// renaming an applet. It rewrites the block between the COMMAND_LIST markers in
// README.md and exits non-zero if those markers are missing.
package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/nao1215/mimixbox/internal/applets"
)

const (
	readmePath = "README.md"
	startMark  = "<!-- COMMAND_LIST_START -->"
	endMark    = "<!-- COMMAND_LIST_END -->"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "genlist:", err)
		os.Exit(1)
	}
}

func run() error {
	data, err := os.ReadFile(readmePath)
	if err != nil {
		return err
	}

	content := string(data)
	start := strings.Index(content, startMark)
	end := strings.Index(content, endMark)
	if start < 0 || end < 0 || end < start {
		return fmt.Errorf("markers %q / %q not found in %s", startMark, endMark, readmePath)
	}

	var b strings.Builder
	b.WriteString(startMark)
	b.WriteString("\n")
	b.WriteString(table())
	b.WriteString(endMark)

	updated := content[:start] + b.String() + content[end+len(endMark):]
	if updated == content {
		return nil // already up to date
	}
	return os.WriteFile(readmePath, []byte(updated), 0o644) //nolint:gosec // README is world-readable
}

// table renders the sorted applet list as a GitHub-flavored Markdown table.
func table() string {
	names := make([]string, 0, len(applets.Applets))
	for name := range applets.Applets {
		names = append(names, name)
	}
	sort.Strings(names)

	var b strings.Builder
	fmt.Fprintf(&b, "There are %d commands. Run `mimixbox --list` to see them on the terminal.\n\n", len(names))
	b.WriteString("| Command | Description |\n")
	b.WriteString("|:--|:--|\n")
	for _, name := range names {
		fmt.Fprintf(&b, "| %s | %s |\n", name, applets.Applets[name].Desc)
	}
	return b.String()
}
