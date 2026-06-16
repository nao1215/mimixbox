package sslutil

import (
	"context"
	"crypto/tls"
	"fmt"

	"github.com/nao1215/mimixbox/internal/command"
)

// NewSSLServer returns an ssl_server command.
func NewSSLServer() *Command { return &Command{name: "ssl_server"} }

func (c *Command) runServer(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet("ssl_server", "-c CERT -k KEY [-b ADDR]", stdio.Err).WithHelp(command.Help{
		Description: "Run a TLS server that terminates connections on a loopback address. -c and -k name " +
			"the PEM certificate and private key; -b sets the listen address (default 127.0.0.1:443). The " +
			"TLS accept loop (ServeTLS) is implemented and exercised by a hermetic local-handshake test; " +
			"running it as a system service requires a certificate and key that are not provided in this " +
			"environment, so without both -c and -k the command fails with a documented error.",
		Examples: []command.Example{
			{Command: "ssl_server -c cert.pem -k key.pem -b 127.0.0.1:8443", Explain: "Terminate TLS on loopback port 8443 using the given cert/key."},
		},
		ExitStatus: "0  clean shutdown.\n1  missing cert/key, load error, or bind error.",
		Notes: []string{
			"Binds a loopback address only; it is intended for local testing, not as a public-facing TLS server.",
			"Pair it with ssl_client to exercise a TLS exchange end to end.",
		},
	})
	certFile := fs.StringP("cert", "c", "", "PEM certificate file")
	keyFile := fs.StringP("key", "k", "", "PEM private key file")
	addr := fs.StringP("bind", "b", "127.0.0.1:443", "address to listen on (HOST:PORT)")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	if *certFile == "" || *keyFile == "" {
		return command.Failuref("a certificate (-c) and key (-k) are required to start the TLS server")
	}
	cfg, err := ServerConfig(*certFile, *keyFile)
	if err != nil {
		return command.Failuref("%v", err)
	}
	ln, err := tls.Listen("tcp", *addr, cfg)
	if err != nil {
		return command.Failuref("cannot listen on %s: %v", *addr, err)
	}
	_, _ = fmt.Fprintf(stdio.Out, "ssl_server: listening on %s\n", ln.Addr().String())
	return ServeTLS(ln, EchoHandler)
}
