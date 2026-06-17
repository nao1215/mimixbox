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

	"github.com/nao1215/mimixbox/internal/applets/netutils/internal/memnet"
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

// fakeServer is a minimal single-session FTP server for tests. It runs over
// in-memory pipes (no loopback socket): the control connection and each PASV
// data connection are buffered in-memory conn pairs handed to the applet via the
// dialTCP seam. serveContent is returned for RETR; stored captures STOR data.
type fakeServer struct {
	serveContent []byte
	stored       []byte
	dataCh       chan net.Conn // data-connection server ends, one per PASV
}

// pasvAddr is the synthetic data address the fixture advertises; the dialTCP
// dispatcher recognizes it and produces an in-memory data connection.
const pasvAddr = "127.0.0.1:258" // 1,2 -> 1*256+2

// startFTP installs the dialTCP seam and starts the fixture control loop over an
// in-memory pipe. No real socket is bound.
func startFTP(t *testing.T, content []byte) *fakeServer {
	t.Helper()
	s := &fakeServer{serveContent: content, dataCh: make(chan net.Conn, 1)}

	controlClient, controlServer := memnet.BufferedConn()
	orig := dialTCP
	dialTCP = func(address string) (net.Conn, error) {
		if address == pasvAddr {
			dataClient, dataServer := memnet.BufferedConn()
			s.dataCh <- dataServer
			return dataClient, nil
		}
		return controlClient, nil // control connection
	}
	t.Cleanup(func() { dialTCP = orig })

	go s.serve(controlServer)
	return s
}

func (s *fakeServer) serve(conn net.Conn) {
	defer func() { _ = conn.Close() }()
	r := bufio.NewReader(conn)
	w := func(line string) { _, _ = fmt.Fprintf(conn, "%s\r\n", line) }

	w("220 fake FTP ready")
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
			w("227 Entering Passive Mode (127,0,0,1,1,2)")
		case "RETR":
			w("150 opening data connection")
			dc := <-s.dataCh
			_, _ = dc.Write(s.serveContent)
			_ = dc.Close()
			w("226 transfer complete")
		case "STOR":
			w("150 opening data connection")
			dc := <-s.dataCh
			var buf bytes.Buffer
			_, _ = buf.ReadFrom(dc)
			s.stored = buf.Bytes()
			_ = dc.Close()
			w("226 transfer complete")
		case "QUIT":
			w("221 bye")
			return
		default:
			w("200 ok")
		}
	}
}

func TestFtpget(t *testing.T) {
	content := bytes.Repeat([]byte("X"), 1500)
	s := startFTP(t, content)

	var written []byte
	orig := writeLocal
	writeLocal = func(_ string, data []byte, _ os.FileMode) error { written = data; return nil }
	t.Cleanup(func() { writeLocal = orig })

	out, _, err := run(t, NewFtpget(), "ftp.test", "out.bin", "remote.bin")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !bytes.Equal(written, content) {
		t.Errorf("downloaded %d bytes, want %d", len(written), len(content))
	}
	if !strings.Contains(out, "Downloaded 1500 bytes") {
		t.Errorf("out = %q", out)
	}
	_ = s
}

func TestFtpput(t *testing.T) {
	content := []byte("upload me")
	s := startFTP(t, nil)

	orig := readLocal
	readLocal = func(string) ([]byte, error) { return content, nil }
	t.Cleanup(func() { readLocal = orig })

	out, _, err := run(t, NewFtpput(), "ftp.test", "in.bin", "remote.bin")
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
