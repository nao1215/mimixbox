package ghrdc_test

import (
	"bytes"
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/shellutils/ghrdc"
	"github.com/nao1215/mimixbox/internal/command"
)

// requireLoopback skips the test when a loopback listen socket cannot be
// created (e.g. a sandbox without networking), since httptest needs one.
func requireLoopback(t *testing.T) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skipf("loopback listen unavailable: %v", err)
	}
	_ = ln.Close()
}

// cannedReleases is a trimmed GitHub releases API response with two releases.
// The first release has one binary asset and one source asset; the second has
// a single binary asset. Asset URLs containing "source"/"src" are counted as
// source code, everything else as binary.
const cannedReleases = `[
  {
    "name": "Version 1.0.0",
    "published_at": "2021-11-20T04:00:23Z",
    "assets": [
      {"download_count": 10, "browser_download_url": "https://example.com/mybin_linux"},
      {"download_count": 3,  "browser_download_url": "https://example.com/source.tar.gz"}
    ]
  },
  {
    "name": "Version 0.9.0",
    "published_at": "2021-11-19T07:27:19Z",
    "assets": [
      {"download_count": 5, "browser_download_url": "https://example.com/mybin_darwin"}
    ]
  }
]`

func newServer(t *testing.T, status int, body string) string {
	t.Helper()
	requireLoopback(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	// SetAPIBaseURL appends "USER/REPOSITORY/releases", so end with a slash.
	restore := ghrdc.SetAPIBaseURL(srv.URL + "/")
	t.Cleanup(restore)
	return srv.URL
}

func run(args ...string) (string, string, error) {
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: &bytes.Buffer{}, Out: out, Err: errBuf}
	err := ghrdc.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestRunLatestRelease(t *testing.T) {
	newServer(t, http.StatusOK, cannedReleases)

	out, _, err := run("nao/repo")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	// Default: only the latest release, binary count 10, source count 3.
	want := "[Name(Version)]             :Version 1.0.0\n" +
		"[Binary Download Count]     :10\n" +
		"[Source Code Download Count]:3\n"
	for _, fragment := range strings.Split(want, "\n") {
		if fragment == "" {
			continue
		}
		if !strings.Contains(out, fragment) {
			t.Errorf("out missing %q\nfull output:\n%s", fragment, out)
		}
	}
	// The second release must NOT appear in default mode.
	if strings.Contains(out, "Version 0.9.0") {
		t.Errorf("default mode should only show latest release, got:\n%s", out)
	}
}

func TestRunAll(t *testing.T) {
	newServer(t, http.StatusOK, cannedReleases)

	out, _, err := run("-a", "nao/repo")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(out, "Version 1.0.0") || !strings.Contains(out, "Version 0.9.0") {
		t.Errorf("-a should show every release, got:\n%s", out)
	}
}

func TestRunTotal(t *testing.T) {
	newServer(t, http.StatusOK, cannedReleases)

	out, _, err := run("-t", "nao/repo")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	// Totals: binary 10+5 = 15, source 3.
	wants := []string{
		"[Name(Version)]                    :All release",
		"[Binary Download Count(total)]     :15",
		"[Source Code Download Count(total)]:3",
	}
	for _, w := range wants {
		if !strings.Contains(out, w) {
			t.Errorf("total output missing %q, got:\n%s", w, out)
		}
	}
}

func TestRunMissingOperand(t *testing.T) {
	out, errOut, err := run()
	if err == nil {
		t.Fatal("expected error for missing operand")
	}
	if out != "" {
		t.Errorf("out = %q, want empty", out)
	}
	if !strings.Contains(errOut, "ghrdc: missing operand") {
		t.Errorf("stderr = %q, want missing operand message", errOut)
	}
}

func TestRunEmptyReleaseData(t *testing.T) {
	newServer(t, http.StatusOK, `[]`)

	out, errOut, err := run("nao/repo")
	if err == nil {
		t.Fatal("expected error for empty release data")
	}
	if out != "" {
		t.Errorf("out = %q, want empty", out)
	}
	if !strings.Contains(errOut, "Release Data is nothing") {
		t.Errorf("stderr = %q, want empty-data message", errOut)
	}
}

func TestRunAPIError(t *testing.T) {
	// A non-JSON body (e.g. an HTML error page) must surface a parse error.
	newServer(t, http.StatusNotFound, "Not Found")

	out, errOut, err := run("nao/repo")
	if err == nil {
		t.Fatal("expected error for bad API response")
	}
	if out != "" {
		t.Errorf("out = %q, want empty", out)
	}
	if !strings.Contains(errOut, "ghrdc:") {
		t.Errorf("stderr = %q, want ghrdc error prefix", errOut)
	}
}

func TestHelp(t *testing.T) {
	out, _, err := run("--help")
	if err != nil {
		t.Fatalf("--help error = %v", err)
	}
	if !strings.Contains(out, "Usage: ghrdc") {
		t.Errorf("--help out = %q", out)
	}
}
