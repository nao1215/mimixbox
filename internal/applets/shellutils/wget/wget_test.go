package wget_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

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

func newServer(t *testing.T) *httptest.Server {
	t.Helper()
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
