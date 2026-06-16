// Package cksum implements the cksum applet: print a POSIX CRC checksum and the
// byte count for each file (or standard input).
package cksum

import (
	"context"
	"fmt"
	"io"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the cksum applet.
type Command struct{}

// New returns a cksum command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "cksum" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Print CRC checksum and byte count of each file" }

// Run executes cksum.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... [FILE]...", stdio.Err).WithHelp(command.Help{
		Description: "Print a POSIX CRC checksum and the byte count for each FILE. With no FILE, or when FILE " +
			"is '-', read standard input. The output is compatible with GNU coreutils' cksum.",
		Examples: []command.Example{
			{Command: "cksum file.txt", Explain: "Print the CRC checksum, byte count, and name of file.txt."},
			{Command: "cksum a.txt b.txt", Explain: "Print a checksum line for each named file."},
			{Command: "cat file | cksum", Explain: "Checksum data read from standard input."},
		},
		ExitStatus: "0  all files were read successfully.\n1  one or more files could not be read.",
	})

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	files := fs.Args()
	if len(files) == 0 {
		return c.sumStdin(stdio)
	}

	var firstErr error
	for _, name := range files {
		if err := c.sumFile(stdio, name); err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "cksum: %s\n", command.FileError(name, err))
			if firstErr == nil {
				firstErr = command.SilentFailure()
			}
		}
	}
	return firstErr
}

// sumStdin checksums standard input and prints "crc bytes".
func (c *Command) sumStdin(stdio command.IO) error {
	data, err := io.ReadAll(stdio.In)
	if err != nil {
		return command.Failure(err)
	}
	crc, n := checksum(data)
	_, err = fmt.Fprintf(stdio.Out, "%d %d\n", crc, n)
	return err
}

// sumFile checksums name and prints "crc bytes name".
func (c *Command) sumFile(stdio command.IO, name string) error {
	r, err := command.Open(stdio, name)
	if err != nil {
		return err
	}
	defer func() { _ = r.Close() }()

	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	crc, n := checksum(data)
	_, err = fmt.Fprintf(stdio.Out, "%d %d %s\n", crc, n, name)
	return err
}

// crcPoly is the polynomial POSIX cksum uses (CRC-32/CKSUM).
const crcPoly = 0x04C11DB7

// crcTable is the precomputed lookup table for crcPoly.
var crcTable = func() [256]uint32 {
	var t [256]uint32
	for i := range t {
		crc := uint32(i) << 24
		for j := 0; j < 8; j++ {
			if crc&0x80000000 != 0 {
				crc = (crc << 1) ^ crcPoly
			} else {
				crc <<= 1
			}
		}
		t[i] = crc
	}
	return t
}()

// checksum returns the POSIX CRC checksum of data and its length in bytes. The
// length is folded into the CRC the way POSIX cksum specifies, so the result
// matches GNU coreutils' cksum.
func checksum(data []byte) (uint32, int) {
	var crc uint32
	for _, b := range data {
		crc = (crc << 8) ^ crcTable[byte(crc>>24)^b]
	}
	for n := len(data); n != 0; n >>= 8 {
		crc = (crc << 8) ^ crcTable[byte(crc>>24)^byte(n)]
	}
	return ^crc, len(data)
}
