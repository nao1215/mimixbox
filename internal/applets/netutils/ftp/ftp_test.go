package ftp

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, cmd *Command, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := cmd.Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestParsePasv(t *testing.T) {
	t.Parallel()
	addr, err := parsePasv("Entering Passive Mode (127,0,0,1,200,100).")
	if err != nil {
		t.Fatalf("parsePasv error = %v", err)
	}
	if addr != "127.0.0.1:51300" {
		t.Errorf("addr = %q, want 127.0.0.1:51300", addr)
	}
	if _, err := parsePasv("garbage"); err == nil {
		t.Error("expected error for malformed response")
	}
}

// fakeServer is a minimal single-session FTP server for tests. serveContent is
// returned for RETR; stored captures STOR data.
type fakeServer struct {
	listener     net.Listener
	serveContent []byte
	stored       []byte
}

func startFTP(t *testing.T, content []byte) *fakeServer {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skipf("loopback TCP unavailable: %v", err)
	}
	s := &fakeServer{listener: ln, serveContent: content}
	t.Cleanup(func() { _ = ln.Close() })
	go s.accept(t)
	return s
}

func (s *fakeServer) addr() (host, port string) {
	host, port, _ = net.SplitHostPort(s.listener.Addr().String())
	return host, port
}

func (s *fakeServer) accept(t *testing.T) {
	conn, err := s.listener.Accept()
	if err != nil {
		return
	}
	defer func() { _ = conn.Close() }()
	r := bufio.NewReader(conn)
	w := func(line string) { _, _ = fmt.Fprintf(conn, "%s\r\n", line) }

	w("220 fake FTP ready")
	var dataLn net.Listener
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return
		}
		fields := strings.Fields(strings.TrimRight(line, "\r\n"))
		if len(fields) == 0 {
			continue
		}
		switch strings.ToUpper(fields[0]) {
		case "USER":
			w("331 need password")
		case "PASS":
			w("230 logged in")
		case "TYPE":
			w("200 type set")
		case "PASV":
			dl, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				w("425 cannot open data connection")
				continue
			}
			dataLn = dl
			_, p, _ := net.SplitHostPort(dl.Addr().String())
			port := atoi(t, p)
			w(fmt.Sprintf("227 Entering Passive Mode (127,0,0,1,%d,%d)", port/256, port%256))
		case "RETR":
			w("150 opening data connection")
			dc, err := dataLn.Accept()
			if err == nil {
				_, _ = dc.Write(s.serveContent)
				_ = dc.Close()
			}
			w("226 transfer complete")
		case "STOR":
			w("150 opening data connection")
			dc, err := dataLn.Accept()
			if err == nil {
				var buf bytes.Buffer
				_, _ = buf.ReadFrom(dc)
				s.stored = buf.Bytes()
				_ = dc.Close()
			}
			w("226 transfer complete")
		case "QUIT":
			w("221 bye")
			return
		default:
			w("200 ok")
		}
	}
}

func atoi(t *testing.T, s string) int {
	t.Helper()
	n := 0
	for _, ch := range s {
		n = n*10 + int(ch-'0')
	}
	return n
}

func TestFtpget(t *testing.T) {
	content := bytes.Repeat([]byte("X"), 1500)
	s := startFTP(t, content)
	host, port := s.addr()

	var written []byte
	orig := writeLocal
	writeLocal = func(_ string, data []byte, _ os.FileMode) error { written = data; return nil }
	t.Cleanup(func() { writeLocal = orig })

	out, _, err := run(t, NewFtpget(), "-P", port, host, "out.bin", "remote.bin")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !bytes.Equal(written, content) {
		t.Errorf("downloaded %d bytes, want %d", len(written), len(content))
	}
	if !strings.Contains(out, "Downloaded 1500 bytes") {
		t.Errorf("out = %q", out)
	}
}

func TestFtpput(t *testing.T) {
	content := []byte("upload me")
	s := startFTP(t, nil)
	host, port := s.addr()

	orig := readLocal
	readLocal = func(string) ([]byte, error) { return content, nil }
	t.Cleanup(func() { readLocal = orig })

	out, _, err := run(t, NewFtpput(), "-P", port, host, "in.bin", "remote.bin")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !bytes.Equal(s.stored, content) {
		t.Errorf("server stored %q, want %q", s.stored, content)
	}
	if !strings.Contains(out, "Uploaded 9 bytes") {
		t.Errorf("out = %q", out)
	}
}

func TestArgValidation(t *testing.T) {
	t.Parallel()
	if _, _, err := run(t, NewFtpget(), "onlyhost"); err == nil {
		t.Error("expected error with missing LOCAL")
	}
}

func TestNamesAndSynopses(t *testing.T) {
	t.Parallel()
	if NewFtpget().Name() != "ftpget" || NewFtpput().Name() != "ftpput" {
		t.Error("names wrong")
	}
	if NewFtpget().Synopsis() == "" || NewFtpput().Synopsis() == "" {
		t.Error("synopsis empty")
	}
}
