package wget_test

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/nao1215/mimixbox/internal/applets/shellutils/wget"
	"github.com/nao1215/mimixbox/internal/command"
)

const body = "hello from test server\n"

func run(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := wget.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func requireLoopback(t *testing.T) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skipf("loopback listen unavailable: %v", err)
	}
	_ = ln.Close()
}

func newServer(t *testing.T) *httptest.Server {
	t.Helper()
	requireLoopback(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/missing" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestDownloadToStdout(t *testing.T) {
	t.Parallel()
	srv := newServer(t)

	out, _, err := run(t, "-O", "-", srv.URL+"/file.txt")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != body {
		t.Errorf("out = %q, want %q", out, body)
	}
}

func TestDownloadToFile(t *testing.T) {
	t.Parallel()
	srv := newServer(t)
	dir := t.TempDir()
	dest := filepath.Join(dir, "downloaded.txt")

	out, _, err := run(t, "-O", dest, srv.URL+"/file.txt")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "" {
		t.Errorf("out = %q, want empty (download goes to file)", out)
	}
	got, rerr := os.ReadFile(dest)
	if rerr != nil {
		t.Fatalf("ReadFile error = %v", rerr)
	}
	if string(got) != body {
		t.Errorf("file content = %q, want %q", got, body)
	}
}

func TestDownloadQuietSuppressesStderr(t *testing.T) {
	t.Parallel()
	srv := newServer(t)

	out, errOut, err := run(t, "-q", "-O", "-", srv.URL+"/file.txt")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != body {
		t.Errorf("out = %q, want %q", out, body)
	}
	if errOut != "" {
		t.Errorf("stderr = %q, want empty with -q", errOut)
	}
}

func TestDownloadNotFound(t *testing.T) {
	t.Parallel()
	srv := newServer(t)

	_, errOut, err := run(t, "-O", "-", srv.URL+"/missing")
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
	if ec := exitCode(t, err); ec != command.ExitFailure {
		t.Errorf("exit code = %d, want %d", ec, command.ExitFailure)
	}
	if !strings.Contains(errOut, "wget: "+srv.URL+"/missing:") {
		t.Errorf("stderr = %q, want wget error prefix", errOut)
	}
}

func TestDownloadBadURL(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, "-O", "-", "not-a-url")
	if err == nil {
		t.Fatal("expected error for invalid URL")
	}
	if !strings.Contains(errOut, "wget: not-a-url:") {
		t.Errorf("stderr = %q, want wget error prefix", errOut)
	}
}

func TestMissingOperand(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t)
	if err == nil {
		t.Fatal("expected error for missing URL")
	}
	if ec := exitCode(t, err); ec != command.ExitFailure {
		t.Errorf("exit code = %d, want %d", ec, command.ExitFailure)
	}
	if !strings.Contains(errOut, "wget: missing URL") {
		t.Errorf("stderr = %q, want missing URL message", errOut)
	}
}

// exitCode extracts the process exit code that err maps to. SilentFailure and
// ExitError both carry ExitFailure; anything else is treated as a failure too.
func exitCode(t *testing.T, err error) int {
	t.Helper()
	if err == nil {
		return command.ExitSuccess
	}
	var ee *command.ExitError
	if ok := asExitError(err, &ee); ok {
		return ee.Code
	}
	return command.ExitFailure
}

func asExitError(err error, target **command.ExitError) bool {
	if e, ok := err.(*command.ExitError); ok {
		*target = e
		return true
	}
	return false
}

func TestDownloadDirectoryPrefix(t *testing.T) {
	t.Parallel()
	srv := newServer(t)
	dir := t.TempDir()

	if _, stderr, err := run(t, "-P", dir, srv.URL+"/file.txt"); err != nil {
		t.Fatalf("Run error = %v (%s)", err, stderr)
	}

	got, err := os.ReadFile(filepath.Join(dir, "file.txt"))
	if err != nil {
		t.Fatalf("-P should write under the prefix dir: %v", err)
	}
	if string(got) != body {
		t.Errorf("content = %q, want %q", got, body)
	}
}

func TestDownloadUserAgent(t *testing.T) {
	t.Parallel()
	requireLoopback(t)
	var gotUA atomic.Value
	gotUA.Store("")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA.Store(r.Header.Get("User-Agent"))
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)

	if _, stderr, err := run(t, "--user-agent", "MimixAgent/1.0", "-O", "-", srv.URL+"/f"); err != nil {
		t.Fatalf("Run error = %v (%s)", err, stderr)
	}
	if ua := gotUA.Load().(string); ua != "MimixAgent/1.0" {
		t.Errorf("server saw User-Agent %q, want %q", ua, "MimixAgent/1.0")
	}
}

func TestDownloadContinueResumes(t *testing.T) {
	t.Parallel()
	requireLoopback(t)
	const full = "0123456789ABCDEF"
	var gotRange atomic.Value
	gotRange.Store("")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rangeHdr := r.Header.Get("Range")
		gotRange.Store(rangeHdr)
		if rangeHdr != "" {
			var start int
			_, _ = fmt.Sscanf(rangeHdr, "bytes=%d-", &start)
			w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, len(full)-1, len(full)))
			w.WriteHeader(http.StatusPartialContent)
			_, _ = w.Write([]byte(full[start:]))
			return
		}
		_, _ = w.Write([]byte(full))
	}))
	t.Cleanup(srv.Close)

	dir := t.TempDir()
	dest := filepath.Join(dir, "out.bin")
	// Pretend the first 5 bytes were already downloaded.
	if err := os.WriteFile(dest, []byte(full[:5]), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, stderr, err := run(t, "-c", "-O", dest, srv.URL+"/big"); err != nil {
		t.Fatalf("Run error = %v (%s)", err, stderr)
	}

	if r := gotRange.Load().(string); r != "bytes=5-" {
		t.Errorf("server saw Range %q, want %q", r, "bytes=5-")
	}
	got, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != full {
		t.Errorf("resumed file = %q, want %q", got, full)
	}
}

func TestDownloadTimeout(t *testing.T) {
	t.Parallel()
	requireLoopback(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(2 * time.Second):
			_, _ = w.Write([]byte(body))
		case <-r.Context().Done(): // client gave up; return promptly
		}
	}))
	t.Cleanup(srv.Close)

	_, _, err := run(t, "-T", "0.1", "-O", "-", srv.URL+"/slow")
	if err == nil {
		t.Errorf("-T 0.1 against a slow server should fail")
	}
}

func TestDownloadRetries(t *testing.T) {
	t.Parallel()
	requireLoopback(t)
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)

	if _, _, err := run(t, "-t", "3", "-O", "-", srv.URL+"/fail"); err == nil {
		t.Errorf("a persistently failing server should make wget fail")
	}
	if got := atomic.LoadInt32(&hits); got != 3 {
		t.Errorf("server was hit %d times, want 3 (the -t value)", got)
	}
}
