package sslutil

import (
	"context"

	"github.com/nao1215/mimixbox/internal/command"
)

// NewSSLClient returns an ssl_client command.
func NewSSLClient() *Command { return &Command{name: "ssl_client"} }

func (c *Command) runClient(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet("ssl_client", "-s HOST:PORT [-k]", stdio.Err).WithHelp(command.Help{
		Description: "Open a TLS connection to HOST:PORT and pipe standard input to the server and the " +
			"server's response to standard output, then close. -s names the server; -k/--insecure skips " +
			"certificate verification (needed for self-signed loopback servers). The TLS dial-and-pipe " +
			"logic (DialAndPipe) is exercised by a hermetic local-handshake test against ssl_server.",
		Examples: []command.Example{
			{Command: "ssl_client -s 127.0.0.1:8443 -k", Explain: "Connect to a self-signed loopback TLS server and pipe stdio."},
		},
		ExitStatus: "0  the exchange completed.\n1  missing server, connection, or TLS handshake error.",
	})
	server := fs.StringP("server", "s", "", "TLS server to connect to (HOST:PORT)")
	insecure := fs.BoolP("insecure", "k", false, "skip certificate verification")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	if *server == "" {
		return command.Failuref("a server address is required (-s)")
	}
	if err := DialAndPipe(*server, ClientConfig(*insecure), stdio.In, stdio.Out); err != nil {
		return command.Failuref("%v", err)
	}
	return nil
}
