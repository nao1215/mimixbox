//
// mimixbox/internal/applets/debianutils/remove-shell/remove-shell.go
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

// Package removeShell implements the remove-shell applet: remove shell names
// from /etc/shells.
package removeShell

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// shellsPath is the file remove-shell operates on by default.
const shellsPath = "/etc/shells"

// Command is the remove-shell applet.
type Command struct{}

// New returns a remove-shell command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "remove-shell" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Remove shell name from /etc/shells" }

// Run executes remove-shell.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "SHELLNAME...", stdio.Err)

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	shells := fs.Args()
	if len(shells) == 0 {
		fmt.Fprintf(stdio.Err, "%s: shellname [shellname ...]\n", c.Name())
		return command.SilentFailure()
	}

	if err := removeShells(shellsPath, shells); err != nil {
		fmt.Fprintf(stdio.Err, "%s: %v\n", c.Name(), err)
		return command.SilentFailure()
	}
	return nil
}

// removeShells drops every named shell from the file at path and rewrites it
// with the remaining lines (in their original order).
func removeShells(path string, shells []string) error {
	lines, err := readShells(path)
	if err != nil {
		return err
	}

	drop := make(map[string]bool, len(shells))
	for _, s := range shells {
		drop[s] = true
	}

	kept := make([]string, 0, len(lines))
	for _, line := range lines {
		if drop[line] {
			continue
		}
		kept = append(kept, line)
	}

	f, err := os.OpenFile(path, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644) //nolint:gosec // /etc/shells is world-readable
	if err != nil {
		return err
	}
	defer f.Close()

	for _, line := range kept {
		if _, err := fmt.Fprintln(f, line); err != nil {
			return err
		}
	}
	return nil
}

// readShells returns the non-empty, whitespace-trimmed lines of the file at
// path. A missing file is treated as an empty list.
func readShells(path string) ([]string, error) {
	f, err := os.Open(path) //nolint:gosec // operating on the named shells file is the point
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var lines []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		lines = append(lines, line)
	}
	return lines, sc.Err()
}
