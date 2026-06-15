//
// mimixbox/internal/applets/textutils/sqluv/tui.go
//
// Copyright 2021 Naohiro CHIKAMATSU
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sqluv

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
	"golang.org/x/term"
)

// runTUI is the minimal interactive viewer. The first port intentionally keeps
// this tiny: it loads every source, lists the resulting tables and their
// columns, then waits for the user to quit. This is enough to prove the applet
// can open a source, render tables, and exit cleanly on a pseudo-terminal
// without panicking. Richer keybindings (browse rows, run SQL, export) are a
// long-term goal handled by the headless path for now.
func runTUI(stdio command.IO, sources []string, opts options) error {
	eng, err := loadAll(sources)
	if err != nil {
		return err
	}
	defer func() { _ = eng.close() }()

	rs, err := eng.query(
		"SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name",
		true,
	)
	if err != nil {
		return err
	}

	renderTUIScreen(stdio.Out, sources, rs)

	if err := recordHistory(opts.historyFile, "[tui] opened "+strings.Join(sources, " ")); err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "sqluv: warning: could not write history: %v\n", err)
	}

	waitForQuit(stdio)
	_, _ = fmt.Fprintln(stdio.Out, "bye")
	return nil
}

// renderTUIScreen writes the minimal full-screen layout: a header, the loaded
// sources, the tables found, and a key legend.
func renderTUIScreen(w io.Writer, sources []string, tables *resultSet) {
	var b strings.Builder
	b.WriteString("sqluv (MimixBox) - minimal viewer\n")
	b.WriteString("=================================\n")
	b.WriteString("Sources:\n")
	for _, s := range sources {
		fmt.Fprintf(&b, "  - %s\n", s)
	}
	b.WriteString("Tables:\n")
	if len(tables.rows) == 0 {
		b.WriteString("  (none)\n")
	}
	for _, row := range tables.rows {
		fmt.Fprintf(&b, "  - %s\n", row[0])
	}
	b.WriteString("\nHeadless query: sqluv --execute 'SELECT ...' <sources> --output=table|csv|tsv|json\n")
	b.WriteString("Keys: q or Ctrl-C to quit.\n")
	_, _ = io.WriteString(w, b.String())
}

// waitForQuit blocks until the user asks to quit. On a real terminal it puts the
// terminal into raw mode and reads a single keystroke ('q' or Ctrl-C). When
// stdin is not a terminal (CI, pipes, the smoke test), it drains stdin and
// returns at EOF so the applet never hangs.
func waitForQuit(stdio command.IO) {
	if f, ok := stdio.In.(*os.File); ok && term.IsTerminal(int(f.Fd())) {
		oldState, err := term.MakeRaw(int(f.Fd()))
		if err == nil {
			defer func() { _ = term.Restore(int(f.Fd()), oldState) }()
			buf := make([]byte, 1)
			for {
				n, rerr := f.Read(buf)
				if rerr != nil || n == 0 {
					return
				}
				switch buf[0] {
				case 'q', 'Q', 3 /* Ctrl-C */, 4 /* Ctrl-D */ :
					return
				}
			}
		}
	}

	// Non-terminal stdin: read until 'q' or EOF so piped input still terminates.
	scanner := bufio.NewScanner(stdio.In)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "q" || line == "quit" {
			return
		}
	}
}
