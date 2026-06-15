// Package rx implements the rx applet: receive a file over a serial link using
// the XMODEM (checksum) protocol.
package rx

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the rx applet.
type Command struct{}

// New returns an rx command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "rx" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Receive a file with the XMODEM protocol" }

// XMODEM control bytes.
const (
	soh = 0x01 // start of a 128-byte data packet
	eot = 0x04 // end of transmission
	ack = 0x06 // acknowledge
	nak = 0x15 // negative acknowledge / start request
	can = 0x18 // cancel
)

const blockSize = 128

// Receive runs the XMODEM (checksum) receiver against the link r/w, writing the
// reassembled payload to out. It is the testable core of rx: pass pipes (or
// buffers preloaded with a captured transfer) instead of a serial port. The
// returned error describes any protocol violation; trailing padding bytes (the
// 0x1A SUB used to fill the last block) are not trimmed.
func Receive(ctx context.Context, r io.Reader, w io.Writer, out io.Writer) error {
	br := bufio.NewReader(r)
	if err := writeByte(w, nak); err != nil { // ask the sender to start
		return err
	}

	expected := byte(1)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		b, err := br.ReadByte()
		if err != nil {
			return fmt.Errorf("reading packet header: %w", err)
		}
		switch b {
		case eot:
			return writeByte(w, ack)
		case can:
			return fmt.Errorf("transfer cancelled by sender")
		case soh:
			// fall through to read the packet
		default:
			return fmt.Errorf("unexpected control byte 0x%02x", b)
		}

		block, err := readPacket(br)
		if err != nil {
			return err
		}
		if block.num != expected && block.num != expected-1 {
			return fmt.Errorf("block number out of sequence: got %d, expected %d", block.num, expected)
		}
		if block.num == expected {
			if _, err := out.Write(block.data); err != nil {
				return err
			}
			expected++
		}
		if err := writeByte(w, ack); err != nil {
			return err
		}
	}
}

// packet is one decoded XMODEM data block.
type packet struct {
	num  byte
	data []byte
}

// readPacket reads the bytes following an SOH: block number, its complement,
// 128 data bytes and a checksum, validating the complement and checksum.
func readPacket(br *bufio.Reader) (*packet, error) {
	num, err := br.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("reading block number: %w", err)
	}
	inv, err := br.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("reading block number complement: %w", err)
	}
	if num != 255-inv {
		return nil, fmt.Errorf("block number complement mismatch: %d vs ^%d", num, inv)
	}
	data := make([]byte, blockSize)
	if _, err := io.ReadFull(br, data); err != nil {
		return nil, fmt.Errorf("reading packet data: %w", err)
	}
	sum, err := br.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("reading checksum: %w", err)
	}
	if got := checksum(data); got != sum {
		return nil, fmt.Errorf("checksum mismatch: got 0x%02x, want 0x%02x", got, sum)
	}
	return &packet{num: num, data: data}, nil
}

// checksum is the simple 8-bit additive checksum XMODEM uses.
func checksum(data []byte) byte {
	var s byte
	for _, b := range data {
		s += b
	}
	return s
}

func writeByte(w io.Writer, b byte) error {
	_, err := w.Write([]byte{b})
	return err
}

// Run executes rx.
func (c *Command) Run(ctx context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "FILE", stdio.Err).WithHelp(command.Help{
		Description: "Receive a file over the serial link on standard input/output using the XMODEM " +
			"protocol (128-byte blocks with an additive checksum) and write it to FILE. rx sends NAK to " +
			"start the transfer, acknowledges each correct block, and stops on EOT. The protocol runs " +
			"over standard input and output, so connect them to a serial device, or, for testing, to " +
			"pipes carrying a captured transfer. The 0x1A padding in the final block is not trimmed.",
		Examples: []command.Example{
			{Command: "rx incoming.bin < /dev/ttyS0 > /dev/ttyS0", Explain: "Receive a file over ttyS0."},
		},
		ExitStatus: "0  the file was received.\n" +
			"1  no FILE was given, the transfer failed, or the file could not be written.",
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	rest := fs.Args()
	if len(rest) == 0 {
		return command.Failuref("an output file is required")
	}
	if len(rest) > 1 {
		return command.Failuref("unexpected argument: %q", rest[1])
	}

	f, err := os.Create(rest[0]) //nolint:gosec // user-named output file
	if err != nil {
		return command.Failuref("cannot create %q: %v", rest[0], err)
	}
	defer func() { _ = f.Close() }()

	if err := Receive(ctx, stdio.In, stdio.Out, f); err != nil {
		return command.Failuref("%v", err)
	}
	return nil
}
