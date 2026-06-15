// Package httpd implements the httpd applet: a small static-file HTTP server.
//
// This is a clean-room reimplementation inspired by BusyBox httpd. The first
// slice serves a document root over loopback in foreground mode so that it can
// be exercised by hermetic tests. Daemonization (dropping the foreground) is
// intentionally not implemented and is reported with a deterministic error.
package httpd

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the httpd applet.
type Command struct{}

// New returns an httpd command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "httpd" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Serve static files over HTTP" }

// Run executes httpd.
func (c *Command) Run(ctx context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "-f [-p ADDR] [-h DIR]", stdio.Err).WithHelp(command.Help{
		Description: "Serve static files from a document root over HTTP. The first slice runs only in " +
			"foreground mode (-f) bound to the address given by -p; the default address is 127.0.0.1:80. " +
			"-h selects the document root (default: the current directory). The server runs until the " +
			"process receives a termination signal or its context is cancelled, then shuts down cleanly.",
		Examples: []command.Example{
			{Command: "httpd -f -p 127.0.0.1:8080 -h ./public", Explain: "Serve ./public on loopback port 8080 in foreground."},
			{Command: "httpd -f -p 127.0.0.1:0 -h .", Explain: "Serve the current directory on an OS-chosen port."},
		},
		ExitStatus: "0  clean shutdown.\n1  bind error, missing document root, or unsupported mode.",
		Notes: []string{
			"Foreground mode (-f) is implemented; background/daemon mode is intentionally not implemented and fails with a documented error.",
			"CGI, authentication, and .htpasswd handling are not implemented in this slice.",
		},
	})
	foreground := fs.BoolP("foreground", "f", false, "run in the foreground (required in this slice)")
	addr := fs.StringP("port", "p", "127.0.0.1:80", "address to listen on (HOST:PORT)")
	home := fs.StringP("home", "h", ".", "document root directory")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	if !*foreground {
		return command.Failuref("background/daemon mode is not implemented; pass -f to run in the foreground")
	}

	srv, ln, err := newServer(*addr, *home)
	if err != nil {
		return command.Failure(err)
	}

	_, _ = fmt.Fprintf(stdio.Out, "httpd: serving %s on http://%s\n", *home, ln.Addr().String())
	return serve(ctx, srv, ln)
}

// newServer builds an http.Server that serves dir and a listener bound to addr.
func newServer(addr, dir string) (*http.Server, net.Listener, error) {
	if strings.TrimSpace(dir) == "" {
		return nil, nil, errors.New("document root must not be empty")
	}
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(dir)))
	srv := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot listen on %s: %w", addr, err)
	}
	return srv, ln, nil
}

// serve runs srv on ln until ctx is cancelled, then shuts down gracefully.
func serve(ctx context.Context, srv *http.Server, ln net.Listener) error {
	errCh := make(chan error, 1)
	go func() {
		err := srv.Serve(ln)
		if errors.Is(err, http.ErrServerClosed) {
			err = nil
		}
		errCh <- err
	}()

	select {
	case <-ctx.Done():
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutCtx)
		<-errCh
		return nil
	case err := <-errCh:
		if err != nil {
			return command.Failure(err)
		}
		return nil
	}
}
