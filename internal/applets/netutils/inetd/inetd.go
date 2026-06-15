// Package inetd implements the inetd applet: a minimal internet super-server.
//
// The config grammar parser is a pure function so it can be table-tested, and
// the server listens on each configured service in the foreground, wiring each
// accepted connection to the configured program's stdin/stdout. This lets a
// hermetic test bring inetd up on a loopback port with a fixture service and
// shut it down cleanly.
package inetd

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the inetd applet.
type Command struct{}

// New returns an inetd command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "inetd" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Internet super-server (minimal)" }

// Service is one configured inetd service line.
type Service struct {
	Port     int    // service port (numeric services only in this slice)
	Socket   string // "stream" (TCP) or "dgram" (UDP)
	Protocol string // "tcp" or "udp"
	Wait     bool   // wait/nowait flag
	User     string // user the program should run as (advisory in this slice)
	Program  string // absolute path or command to run
	Args     []string
}

// Run executes inetd.
func (c *Command) Run(ctx context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-f] CONFIG", stdio.Err).WithHelp(command.Help{
		Description: "Read a configuration file and listen on each configured service, running the named " +
			"program for every incoming connection with the socket wired to its standard input and output. " +
			"-f keeps inetd in the foreground (the only supported mode in this slice). Each config line is: " +
			"'PORT SOCKETTYPE PROTOCOL WAIT USER PROGRAM [ARGS...]', where SOCKETTYPE is stream or dgram, " +
			"PROTOCOL is tcp or udp, and WAIT is wait or nowait. Blank lines and '#' comments are ignored.",
		Examples: []command.Example{
			{Command: "inetd -f inetd.conf", Explain: "Run services from inetd.conf in the foreground."},
		},
		ExitStatus: "0  clean shutdown.\n1  config error or a service could not bind.",
		Notes: []string{
			"Only numeric ports and tcp/udp protocols are supported; named services from /etc/services are not resolved.",
			"Background/daemon mode is not implemented; pass -f.",
		},
	})
	foreground := fs.BoolP("foreground", "f", false, "run in the foreground (required in this slice)")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	if !*foreground {
		return command.Failuref("only foreground mode is implemented; pass -f")
	}
	rest := fs.Args()
	if len(rest) < 1 {
		return command.Failuref("usage: inetd -f CONFIG")
	}

	f, err := os.Open(rest[0])
	if err != nil {
		return command.Failuref("cannot open config %q: %v", rest[0], err)
	}
	defer func() { _ = f.Close() }()

	services, err := ParseConfig(f)
	if err != nil {
		return command.Failuref("%s: %v", rest[0], err)
	}
	if len(services) == 0 {
		return command.Failuref("%s: no services configured", rest[0])
	}

	return Serve(ctx, stdio, services, runProgram)
}

// ParseConfig parses the minimal inetd config grammar from r.
func ParseConfig(r io.Reader) ([]Service, error) {
	var services []Service
	sc := bufio.NewScanner(r)
	line := 0
	for sc.Scan() {
		line++
		text := strings.TrimSpace(sc.Text())
		if text == "" || strings.HasPrefix(text, "#") {
			continue
		}
		f := strings.Fields(text)
		if len(f) < 6 {
			return nil, fmt.Errorf("line %d: expected 'PORT SOCKET PROTO WAIT USER PROGRAM ...', got %q", line, text)
		}
		port, err := strconv.Atoi(f[0])
		if err != nil || port < 1 || port > 65535 {
			return nil, fmt.Errorf("line %d: invalid port %q", line, f[0])
		}
		socket := f[1]
		if socket != "stream" && socket != "dgram" {
			return nil, fmt.Errorf("line %d: socket type must be stream or dgram, got %q", line, socket)
		}
		proto := f[2]
		if proto != "tcp" && proto != "udp" {
			return nil, fmt.Errorf("line %d: protocol must be tcp or udp, got %q", line, proto)
		}
		var wait bool
		switch f[3] {
		case "wait":
			wait = true
		case "nowait":
			wait = false
		default:
			return nil, fmt.Errorf("line %d: wait flag must be wait or nowait, got %q", line, f[3])
		}
		services = append(services, Service{
			Port:     port,
			Socket:   socket,
			Protocol: proto,
			Wait:     wait,
			User:     f[4],
			Program:  f[5],
			Args:     f[6:],
		})
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return services, nil
}

// Runner runs svc's program with conn wired to its stdin/stdout.
type Runner func(ctx context.Context, svc Service, conn net.Conn) error

// runProgram is the production Runner: it forks the configured program.
func runProgram(ctx context.Context, svc Service, conn net.Conn) error {
	cmd := exec.CommandContext(ctx, svc.Program, svc.Args...)
	cmd.Stdin = conn
	cmd.Stdout = conn
	return cmd.Run()
}

// Serve listens on every TCP service (dgram services are accepted but not yet
// looped) and dispatches each connection to runner until ctx is cancelled.
func Serve(ctx context.Context, stdio command.IO, services []Service, runner Runner) error {
	var wg sync.WaitGroup
	var firstErr error
	var mu sync.Mutex
	listeners := make([]net.Listener, 0, len(services))

	for _, svc := range services {
		if svc.Protocol != "tcp" {
			// dgram/udp super-serving is not implemented in this slice; skip but report.
			_, _ = fmt.Fprintf(stdio.Err, "inetd: skipping unsupported %s service on port %d\n", svc.Protocol, svc.Port)
			continue
		}
		addr := net.JoinHostPort("127.0.0.1", strconv.Itoa(svc.Port))
		ln, err := net.Listen("tcp", addr)
		if err != nil {
			closeAll(listeners)
			return command.Failuref("cannot listen on %s: %v", addr, err)
		}
		listeners = append(listeners, ln)
		_, _ = fmt.Fprintf(stdio.Out, "inetd: listening on %s for %s\n", ln.Addr().String(), svc.Program)

		wg.Add(1)
		go func(ln net.Listener, svc Service) {
			defer wg.Done()
			if err := acceptLoop(ctx, ln, svc, runner); err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				mu.Unlock()
			}
		}(ln, svc)
	}

	if len(listeners) == 0 {
		return command.Failuref("no supported (tcp) services to serve")
	}

	go func() {
		<-ctx.Done()
		closeAll(listeners)
	}()

	wg.Wait()
	return firstErr
}

func acceptLoop(ctx context.Context, ln net.Listener, svc Service, runner Runner) error {
	var wg sync.WaitGroup
	for {
		conn, err := ln.Accept()
		if err != nil {
			if ctx.Err() != nil {
				wg.Wait()
				return nil
			}
			return command.Failuref("accept on port %d: %v", svc.Port, err)
		}
		wg.Add(1)
		go func(c net.Conn) {
			defer wg.Done()
			defer func() { _ = c.Close() }()
			_ = runner(ctx, svc, c)
		}(conn)
	}
}

func closeAll(lns []net.Listener) {
	for _, ln := range lns {
		_ = ln.Close()
	}
}
