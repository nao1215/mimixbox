// Package ftpd implements the ftpd applet: a minimal read-only FTP server.
//
// The command-parsing and reply-formatting logic is factored into pure helpers
// so it can be table-tested, and the control loop runs over loopback for
// hermetic integration tests. The first slice supports anonymous login plus
// SYST, TYPE, PWD, CWD, PASV, LIST, and RETR for read-only access to a root
// directory; mutating commands are refused with a documented error reply.
package ftpd

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the ftpd applet.
type Command struct{}

// New returns an ftpd command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "ftpd" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Minimal read-only FTP server (foreground)" }

// Run executes ftpd.
func (c *Command) Run(ctx context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "-f [-b ADDR] [DIR]", stdio.Err).WithHelp(command.Help{
		Description: "Run a read-only FTP server rooted at DIR (default: current directory). -f keeps ftpd " +
			"in the foreground; -b sets the control listen address (default 127.0.0.1:21). Anonymous login " +
			"is accepted; passive mode (PASV) is used for data transfers. Supported commands: USER, PASS, " +
			"SYST, TYPE, PWD, CWD, CDUP, PASV, LIST, NLST, RETR, QUIT. Write commands are refused with 550.",
		Examples: []command.Example{
			{Command: "ftpd -f -b 127.0.0.1:2121 ./pub", Explain: "Serve ./pub read-only over FTP on loopback port 2121."},
		},
		ExitStatus: "0  clean shutdown.\n1  bad arguments or bind error.",
		Notes: []string{
			"Uploads (STOR), deletes, and renames are intentionally not implemented; this slice is read-only.",
		},
	})
	foreground := fs.BoolP("foreground", "f", false, "run in the foreground (required in this slice)")
	addr := fs.StringP("bind", "b", "127.0.0.1:21", "control address to listen on (HOST:PORT)")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	if !*foreground {
		return command.Failuref("only foreground mode is implemented; pass -f")
	}
	rootArg := "."
	if rest := fs.Args(); len(rest) > 0 {
		rootArg = rest[0]
	}
	root, err := filepath.Abs(rootArg)
	if err != nil {
		return command.Failuref("invalid root %q: %v", rootArg, err)
	}

	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		return command.Failuref("cannot listen on %s: %v", *addr, err)
	}
	_, _ = fmt.Fprintf(stdio.Out, "ftpd: serving %s on %s\n", root, ln.Addr().String())
	return Serve(ctx, ln, root)
}

// Serve runs the FTP control accept loop on ln until ctx is cancelled.
func Serve(ctx context.Context, ln net.Listener, root string) error {
	go func() {
		<-ctx.Done()
		_ = ln.Close()
	}()
	var wg sync.WaitGroup
	for {
		conn, err := ln.Accept()
		if err != nil {
			if ctx.Err() != nil {
				wg.Wait()
				return nil
			}
			return command.Failuref("accept: %v", err)
		}
		wg.Add(1)
		go func(c net.Conn) {
			defer wg.Done()
			defer func() { _ = c.Close() }()
			newSession(c, root).serve()
		}(conn)
	}
}

// session holds the per-connection FTP state.
type session struct {
	conn   net.Conn
	root   string // absolute server root
	cwd    string // virtual cwd, always starts with "/"
	dataLn net.Listener
	binary bool
}

func newSession(conn net.Conn, root string) *session {
	return &session{conn: conn, root: root, cwd: "/"}
}

func (s *session) reply(code int, msg string) {
	_, _ = fmt.Fprintf(s.conn, "%d %s\r\n", code, msg)
}

func (s *session) serve() {
	s.reply(220, "mimixbox ftpd ready")
	sc := bufio.NewScanner(s.conn)
	for sc.Scan() {
		verb, arg := SplitCommand(sc.Text())
		if quit := s.dispatch(verb, arg); quit {
			return
		}
	}
}

// dispatch handles one command; it returns true when the connection should close.
func (s *session) dispatch(verb, arg string) (quit bool) {
	switch verb {
	case "USER":
		s.reply(331, "anonymous login ok, send your email as password")
	case "PASS":
		s.reply(230, "login successful")
	case "SYST":
		s.reply(215, "UNIX Type: L8")
	case "TYPE":
		s.binary = strings.EqualFold(arg, "I")
		s.reply(200, "type set")
	case "PWD", "XPWD":
		s.reply(257, fmt.Sprintf("%q is the current directory", s.cwd))
	case "CWD":
		s.cwd = ResolvePath(s.cwd, arg)
		s.reply(250, "directory changed to "+s.cwd)
	case "CDUP":
		s.cwd = ResolvePath(s.cwd, "..")
		s.reply(250, "directory changed to "+s.cwd)
	case "PASV":
		s.handlePASV()
	case "LIST", "NLST":
		s.handleList()
	case "RETR":
		s.handleRetr(arg)
	case "NOOP":
		s.reply(200, "ok")
	case "QUIT":
		s.reply(221, "goodbye")
		return true
	case "STOR", "DELE", "RNFR", "RNTO", "MKD", "RMD":
		s.reply(550, "read-only server: command not permitted")
	default:
		s.reply(502, "command not implemented")
	}
	return false
}

// newDataListener opens a passive-mode data listener and reports the port to
// advertise in the PASV reply. It is a package-level seam so tests can supply an
// in-memory listener (and learn its dial handle) instead of binding a loopback
// socket. Production binds 127.0.0.1:0.
var newDataListener = func() (ln net.Listener, port int, err error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, 0, err
	}
	return l, l.Addr().(*net.TCPAddr).Port, nil
}

// handlePASV opens a data listener and announces it in PASV format.
func (s *session) handlePASV() {
	ln, p, err := newDataListener()
	if err != nil {
		s.reply(425, "cannot open data connection")
		return
	}
	if s.dataLn != nil {
		_ = s.dataLn.Close()
	}
	s.dataLn = ln
	s.reply(227, fmt.Sprintf("Entering Passive Mode (127,0,0,1,%d,%d)", p>>8, p&0xff))
}

func (s *session) acceptData() net.Conn {
	if s.dataLn == nil {
		return nil
	}
	conn, err := s.dataLn.Accept()
	_ = s.dataLn.Close()
	s.dataLn = nil
	if err != nil {
		return nil
	}
	return conn
}

func (s *session) handleList() {
	dataConn := s.acceptData()
	if dataConn == nil {
		s.reply(425, "use PASV first")
		return
	}
	defer func() { _ = dataConn.Close() }()
	s.reply(150, "here comes the directory listing")
	dir, err := s.realPath(s.cwd)
	if err == nil {
		entries, _ := os.ReadDir(dir)
		for _, e := range entries {
			_, _ = fmt.Fprintf(dataConn, "%s\r\n", FormatListEntry(e))
		}
	}
	s.reply(226, "directory send ok")
}

func (s *session) handleRetr(name string) {
	virt := ResolvePath(s.cwd, name)
	real, err := s.realPath(virt)
	if err != nil {
		s.reply(550, "no such file")
		return
	}
	data, err := os.ReadFile(real)
	if err != nil {
		s.reply(550, "cannot read file")
		return
	}
	dataConn := s.acceptData()
	if dataConn == nil {
		s.reply(425, "use PASV first")
		return
	}
	defer func() { _ = dataConn.Close() }()
	s.reply(150, "opening data connection")
	_, _ = dataConn.Write(data)
	s.reply(226, "transfer complete")
}

// realPath maps a virtual path to a filesystem path inside root, rejecting
// traversal outside root.
func (s *session) realPath(virt string) (string, error) {
	clean := path.Clean("/" + strings.TrimPrefix(virt, "/"))
	full := filepath.Join(s.root, filepath.FromSlash(clean))
	if full != s.root && !strings.HasPrefix(full, s.root+string(os.PathSeparator)) {
		return "", os.ErrPermission
	}
	return full, nil
}

// SplitCommand splits an FTP command line into an upper-cased verb and its
// argument.
func SplitCommand(line string) (verb, arg string) {
	line = strings.TrimRight(line, "\r\n")
	parts := strings.SplitN(line, " ", 2)
	verb = strings.ToUpper(strings.TrimSpace(parts[0]))
	if len(parts) > 1 {
		arg = strings.TrimSpace(parts[1])
	}
	return verb, arg
}

// ResolvePath resolves arg against the virtual working directory cwd, returning
// a cleaned absolute virtual path. Paths never escape "/".
func ResolvePath(cwd, arg string) string {
	if arg == "" {
		return cwd
	}
	var joined string
	if strings.HasPrefix(arg, "/") {
		joined = arg
	} else {
		joined = path.Join(cwd, arg)
	}
	clean := path.Clean("/" + strings.TrimPrefix(joined, "/"))
	return clean
}

// FormatListEntry renders a single LIST line for a directory entry in a
// simplified ls -l style.
func FormatListEntry(e os.DirEntry) string {
	typ := "-"
	if e.IsDir() {
		typ = "d"
	}
	var size int64
	if info, err := e.Info(); err == nil {
		size = info.Size()
	}
	return fmt.Sprintf("%srw-r--r-- 1 ftp ftp %8d Jan  1 00:00 %s", typ, size, e.Name())
}
