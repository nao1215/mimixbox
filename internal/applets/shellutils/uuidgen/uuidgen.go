//
// mimixbox/internal/applets/shellutils/uuidgen/uuidgen.go
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

// Package uuidgen implements the uuidgen applet: print a random (version 4)
// UUID in the canonical 8-4-4-4-12 lowercase hexadecimal form.
package uuidgen

import (
	"context"
	"crypto/rand"
	"fmt"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the uuidgen applet.
type Command struct{}

// New returns a uuidgen command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "uuidgen" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Print UUID (Universally Unique IDentifier)" }

// Run executes uuidgen. It generates a random (version 4) UUID and prints it
// lowercase, followed by a newline, to stdio.Out.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]", stdio.Err).WithHelp(command.Help{
		Description: "Print a new random (version 4) UUID in the canonical 8-4-4-4-12 lowercase " +
			"hexadecimal form. Each invocation prints one UUID followed by a newline.",
		Examples: []command.Example{
			{Command: "uuidgen", Explain: "Print a random version 4 UUID."},
		},
		ExitStatus: "0  a UUID was printed.\n1  a UUID could not be generated.",
	})

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	id, err := uuidV4()
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "uuidgen: %v\n", err)
		return command.SilentFailure()
	}
	_, _ = fmt.Fprintln(stdio.Out, id)
	return nil
}

// uuidV4 returns a randomly generated version 4 UUID in the canonical
// 8-4-4-4-12 lowercase hexadecimal form.
//
//	xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx
//	              ^    ^
//	              |    |
//	              |    variant bits: one of 8, 9, a, b
//	              version, always "4"
func uuidV4() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	// Set the version (4) and variant (RFC 4122) bits.
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80

	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}
