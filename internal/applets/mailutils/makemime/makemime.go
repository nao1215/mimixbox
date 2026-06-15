// Package makemime implements the makemime applet: build a MIME-encoded message
// from one or more attachment files. The construction is purely local; no
// network or mailbox access is involved.
package makemime

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"mime"
	"mime/multipart"
	"net/textproto"
	"os"
	"path/filepath"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the makemime applet.
type Command struct{}

// New returns a makemime command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "makemime" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Create a MIME-encoded message from files" }

// Run executes makemime.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-c TYPE] [-o OUTFILE] FILE...", stdio.Err).WithHelp(command.Help{
		Description: "Build a MIME message from one or more FILEs and write it to standard output, or to " +
			"OUTFILE with -o. With a single file a simple single-part message is produced; with " +
			"several files a multipart/mixed message is produced, one part per file. The Content-Type " +
			"of each part defaults to the value of -c (application/octet-stream when unset) and the " +
			"body is base64-encoded. This is a local construction with no network delivery.",
		Examples: []command.Example{
			{Command: "makemime -o out.eml input.txt", Explain: "Wrap input.txt in a MIME message saved to out.eml."},
			{Command: "makemime -c text/plain a.txt b.txt", Explain: "Build a multipart/mixed message from two files."},
		},
		ExitStatus: "0  success.\n1  a file could not be read or the output could not be written.",
	})
	ctype := fs.StringP("content-type", "c", "application/octet-stream", "Content-Type for each part")
	outfile := fs.StringP("output", "o", "", "write the message to this file instead of standard output")
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	files := fs.Args()
	if len(files) == 0 {
		fmt.Fprintf(stdio.Err, "%s: no input file given\n", c.Name())
		return command.SilentFailure()
	}

	msg, err := build(files, *ctype)
	if err != nil {
		fmt.Fprintf(stdio.Err, "%s: %v\n", c.Name(), err)
		return command.SilentFailure()
	}

	if *outfile != "" {
		if err := os.WriteFile(*outfile, msg, 0o600); err != nil {
			fmt.Fprintf(stdio.Err, "%s: %v\n", c.Name(), err)
			return command.SilentFailure()
		}
		return nil
	}
	_, _ = stdio.Out.Write(msg)
	return nil
}

// build assembles the MIME message bytes from files using contentType for each
// part. A single file yields a simple message; multiple files yield
// multipart/mixed.
func build(files []string, contentType string) ([]byte, error) {
	if len(files) == 1 {
		return buildSingle(files[0], contentType)
	}
	return buildMultipart(files, contentType)
}

func buildSingle(file, contentType string) ([]byte, error) {
	data, err := os.ReadFile(file) //nolint:gosec // user-named file
	if err != nil {
		return nil, err
	}
	var b bytes.Buffer
	fmt.Fprintf(&b, "MIME-Version: 1.0\r\n")
	fmt.Fprintf(&b, "Content-Type: %s\r\n", contentType)
	fmt.Fprintf(&b, "Content-Transfer-Encoding: base64\r\n")
	fmt.Fprintf(&b, "Content-Disposition: inline; filename=%q\r\n", filepath.Base(file))
	b.WriteString("\r\n")
	b.WriteString(base64Lines(data))
	return b.Bytes(), nil
}

func buildMultipart(files []string, contentType string) ([]byte, error) {
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	for _, f := range files {
		data, err := os.ReadFile(f) //nolint:gosec // user-named file
		if err != nil {
			return nil, err
		}
		h := textproto.MIMEHeader{}
		h.Set("Content-Type", contentType)
		h.Set("Content-Transfer-Encoding", "base64")
		h.Set("Content-Disposition", fmt.Sprintf("inline; filename=%q", filepath.Base(f)))
		part, err := mw.CreatePart(h)
		if err != nil {
			return nil, err
		}
		if _, err := part.Write([]byte(base64Lines(data))); err != nil {
			return nil, err
		}
	}
	if err := mw.Close(); err != nil {
		return nil, err
	}

	var b bytes.Buffer
	fmt.Fprintf(&b, "MIME-Version: 1.0\r\n")
	fmt.Fprintf(&b, "Content-Type: %s\r\n", mime.FormatMediaType("multipart/mixed", map[string]string{"boundary": mw.Boundary()}))
	b.WriteString("\r\n")
	b.Write(body.Bytes())
	return b.Bytes(), nil
}

// base64Lines base64-encodes data and wraps it at 76 columns with CRLF, the
// MIME convention.
func base64Lines(data []byte) string {
	encoded := base64.StdEncoding.EncodeToString(data)
	var b bytes.Buffer
	for len(encoded) > 76 {
		b.WriteString(encoded[:76])
		b.WriteString("\r\n")
		encoded = encoded[76:]
	}
	b.WriteString(encoded)
	b.WriteString("\r\n")
	return b.String()
}
