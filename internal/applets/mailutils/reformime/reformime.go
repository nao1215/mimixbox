// Package reformime implements the reformime applet: parse a MIME message read
// from standard input and either list its structure or extract a decoded part.
// Parsing is purely local and hermetic.
package reformime

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/mail"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the reformime applet.
type Command struct{}

// New returns a reformime command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "reformime" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Parse a MIME message and list or extract its parts" }

// Run executes reformime.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-x N]", stdio.Err).WithHelp(command.Help{
		Description: "Read a MIME message from standard input. By default it lists each part's index and " +
			"Content-Type, one per line. With -x N it instead decodes part N (counting from 1) and " +
			"writes the decoded body to standard output. A non-multipart message counts as a single " +
			"part (1). No mailbox or network access is performed.",
		Examples: []command.Example{
			{Command: "reformime < message.eml", Explain: "List the parts of message.eml."},
			{Command: "reformime -x 1 < message.eml", Explain: "Decode and print part 1."},
		},
		ExitStatus: "0  success.\n1  the input is not a valid message or the requested part does not exist.",
	})
	extract := fs.IntP("extract", "x", 0, "decode part N (1-based) to standard output")
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	parts, err := parse(stdio.In)
	if err != nil {
		fmt.Fprintf(stdio.Err, "%s: %v\n", c.Name(), err)
		return command.SilentFailure()
	}

	if *extract > 0 {
		if *extract > len(parts) {
			fmt.Fprintf(stdio.Err, "%s: no such part: %d\n", c.Name(), *extract)
			return command.SilentFailure()
		}
		_, _ = stdio.Out.Write(parts[*extract-1].body)
		return nil
	}

	for i, p := range parts {
		fmt.Fprintf(stdio.Out, "%d\t%s\n", i+1, p.contentType)
	}
	return nil
}

// part is one decoded MIME part.
type part struct {
	contentType string
	body        []byte
}

// parse reads an RFC 5322 message and returns its MIME parts. For a multipart
// message each sub-part is returned; otherwise the whole body is a single part.
func parse(r io.Reader) ([]part, error) {
	msg, err := mail.ReadMessage(bufio.NewReader(r))
	if err != nil {
		return nil, err
	}
	ctype := msg.Header.Get("Content-Type")
	if ctype == "" {
		ctype = "text/plain"
	}
	mediatype, params, err := mime.ParseMediaType(ctype)
	if err != nil {
		// Treat an unparseable Content-Type as a single opaque part, but still
		// surface a body read error instead of returning an empty part.
		body, rerr := io.ReadAll(msg.Body)
		if rerr != nil {
			return nil, rerr
		}
		return []part{{contentType: ctype, body: body}}, nil
	}

	if !strings.HasPrefix(mediatype, "multipart/") {
		body, err := io.ReadAll(msg.Body)
		if err != nil {
			return nil, err
		}
		body = decodeBody(msg.Header.Get("Content-Transfer-Encoding"), body)
		return []part{{contentType: mediatype, body: body}}, nil
	}

	boundary := params["boundary"]
	if boundary == "" {
		return nil, fmt.Errorf("multipart message without boundary")
	}
	mr := multipart.NewReader(msg.Body, boundary)
	var parts []part
	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		raw, err := io.ReadAll(p)
		if err != nil {
			return nil, err
		}
		pct := p.Header.Get("Content-Type")
		if pct == "" {
			pct = "text/plain"
		}
		raw = decodeBody(p.Header.Get("Content-Transfer-Encoding"), raw)
		parts = append(parts, part{contentType: pct, body: raw})
	}
	return parts, nil
}

// decodeBody decodes a part body according to its Content-Transfer-Encoding.
// Unknown or identity encodings are returned unchanged.
func decodeBody(encoding string, body []byte) []byte {
	switch strings.ToLower(strings.TrimSpace(encoding)) {
	case "base64":
		clean := bytes.ReplaceAll(body, []byte("\r\n"), nil)
		clean = bytes.ReplaceAll(clean, []byte("\n"), nil)
		dec, err := base64.StdEncoding.DecodeString(string(clean))
		if err != nil {
			return body
		}
		return dec
	default:
		return body
	}
}
