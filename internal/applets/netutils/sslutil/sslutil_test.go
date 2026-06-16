package sslutil

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
)

// selfSigned builds a self-signed certificate valid for 127.0.0.1.
func selfSigned(t *testing.T) tls.Certificate {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "localhost"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	der, err := x509.CreateCertificate(rand.Reader, &tmpl, &tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	return tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key}
}

func TestLocalTLSHandshakeEcho(t *testing.T) {
	t.Parallel()
	cert := selfSigned(t)
	ln, err := tls.Listen("tcp", "127.0.0.1:0", &tls.Config{Certificates: []tls.Certificate{cert}})
	if err != nil {
		t.Skipf("loopback TLS listen unavailable: %v", err)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = ServeTLS(ln, EchoHandler)
	}()

	out := &bytes.Buffer{}
	clientCfg := &tls.Config{InsecureSkipVerify: true} //nolint:gosec // self-signed test cert
	if err := DialAndPipe(ln.Addr().String(), clientCfg, strings.NewReader("secret payload"), out); err != nil {
		t.Fatalf("DialAndPipe error: %v", err)
	}
	if out.String() != "secret payload" {
		t.Errorf("echo = %q, want 'secret payload'", out.String())
	}

	_ = ln.Close()
	wg.Wait()
}

func TestServerRequiresCertAndKey(t *testing.T) {
	t.Parallel()
	stdio := command.IO{In: bytes.NewReader(nil), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	if err := NewSSLServer().Run(context.Background(), stdio, []string{"-b", "127.0.0.1:0"}); err == nil {
		t.Error("ssl_server should fail without cert/key")
	}
}

func TestClientRequiresServer(t *testing.T) {
	t.Parallel()
	stdio := command.IO{In: bytes.NewReader(nil), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	if err := NewSSLClient().Run(context.Background(), stdio, nil); err == nil {
		t.Error("ssl_client should fail without -s")
	}
}

// TestClientConfig checks the shared client tls.Config helper honors the
// insecure (-k) flag and otherwise verifies certificates by default.
func TestClientConfig(t *testing.T) {
	t.Parallel()
	if got := ClientConfig(true); !got.InsecureSkipVerify {
		t.Error("ClientConfig(true) should skip verification")
	}
	if got := ClientConfig(false); got.InsecureSkipVerify {
		t.Error("ClientConfig(false) should verify certificates")
	}
}

// TestServerConfig checks the shared server setup loads a PEM cert/key into a
// usable tls.Config and reports an error for a missing/invalid pair.
func TestServerConfig(t *testing.T) {
	t.Parallel()
	cert := selfSigned(t)
	keyDER, err := x509.MarshalPKCS8PrivateKey(cert.PrivateKey)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	certPath := filepath.Join(dir, "cert.pem")
	keyPath := filepath.Join(dir, "key.pem")
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Certificate[0]})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER})
	if err := os.WriteFile(certPath, certPEM, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(keyPath, keyPEM, 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := ServerConfig(certPath, keyPath)
	if err != nil {
		t.Fatalf("ServerConfig error: %v", err)
	}
	if len(cfg.Certificates) != 1 {
		t.Errorf("ServerConfig loaded %d certificates, want 1", len(cfg.Certificates))
	}
	if _, err := ServerConfig(filepath.Join(dir, "missing.pem"), keyPath); err == nil {
		t.Error("ServerConfig should fail on a missing certificate")
	}
}

// TestServerHelpNotes asserts ssl_server --help documents a Notes section.
func TestServerHelpNotes(t *testing.T) {
	t.Parallel()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := NewSSLServer().Run(context.Background(), io, []string{"--help"}); err != nil {
		t.Fatalf("--help err = %v", err)
	}
	if !strings.Contains(out.String(), "Notes:") {
		t.Errorf("ssl_server --help missing Notes section: %q", out.String())
	}
}
