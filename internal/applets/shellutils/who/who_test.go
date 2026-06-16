package who_test

import (
	"bytes"
	"context"
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/nao1215/mimixbox/internal/applets/shellutils/who"
	"github.com/nao1215/mimixbox/internal/command"
)

// ut_type values mirrored from who.go for fixture construction.
const (
	bootTime    = 2
	userProcess = 7
)

const utmpRecordSize = 384

// utmpRecord builds one 384-byte struct utmp record with the layout who reads.
func utmpRecord(utType int16, user, line, host string, sec int32) []byte {
	rec := make([]byte, utmpRecordSize)
	binary.LittleEndian.PutUint16(rec[0:], uint16(utType)) // ut_type   int16 @0
	copy(rec[8:40], line)                                  // ut_line   char[32] @8
	copy(rec[44:76], user)                                 // ut_user   char[32] @44
	copy(rec[76:332], host)                                // ut_host   char[256] @76
	binary.LittleEndian.PutUint32(rec[340:], uint32(sec))  // ut_tv.tv_sec int32 @340
	return rec
}

// writeFixture writes a utmp file with one USER_PROCESS and one BOOT_TIME
// record and returns its path. The login time is fixed for determinism.
func writeFixture(t *testing.T) string {
	t.Helper()
	// 2021-01-02 03:04 UTC -> set TZ to UTC so the formatted output matches.
	loginSec := int32(time.Date(2021, 1, 2, 3, 4, 0, 0, time.UTC).Unix())
	bootSec := int32(time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC).Unix())

	var buf bytes.Buffer
	buf.Write(utmpRecord(userProcess, "alice", "tty1", "localhost", loginSec))
	buf.Write(utmpRecord(bootTime, "reboot", "~", "", bootSec))

	path := filepath.Join(t.TempDir(), "utmp")
	if err := os.WriteFile(path, buf.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

// run invokes who with args, passing path as the trailing FILE operand so the
// applet reads the fixture instead of /var/run/utmp.
func run(t *testing.T, path string, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	if path != "" {
		args = append(append([]string{}, args...), path)
	}
	err := who.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestWhoDefault(t *testing.T) {
	t.Setenv("TZ", "UTC")
	path := writeFixture(t)
	restore := who.SetUtmpFileForTest(path)
	defer restore()

	out, errOut, err := run(t, path)
	if err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}
	if !strings.Contains(out, "alice") {
		t.Errorf("out = %q, want NAME alice", out)
	}
	if !strings.Contains(out, "tty1") {
		t.Errorf("out = %q, want LINE tty1", out)
	}
	if !strings.Contains(out, "2021-01-02 03:04") {
		t.Errorf("out = %q, want login time", out)
	}
	// The boot record must not appear in the default listing.
	if strings.Contains(out, "system") {
		t.Errorf("default out should not contain boot line: %q", out)
	}
}

func TestWhoBoot(t *testing.T) {
	t.Setenv("TZ", "UTC")
	path := writeFixture(t)

	out, errOut, err := run(t, path, "-b")
	if err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}
	if !strings.Contains(out, "system boot") {
		t.Errorf("out = %q, want 'system boot'", out)
	}
	if !strings.Contains(out, "2021-01-01 00:00") {
		t.Errorf("out = %q, want boot time", out)
	}
	if strings.Contains(out, "alice") {
		t.Errorf("-b out should not list users: %q", out)
	}
}

func TestWhoCount(t *testing.T) {
	path := writeFixture(t)

	out, errOut, err := run(t, path, "-q")
	if err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}
	if !strings.Contains(out, "alice") {
		t.Errorf("-q out = %q, want name alice", out)
	}
	if !strings.Contains(out, "# users=1") {
		t.Errorf("-q out = %q, want '# users=1'", out)
	}
}

func TestWhoHeading(t *testing.T) {
	t.Setenv("TZ", "UTC")
	path := writeFixture(t)

	out, errOut, err := run(t, path, "-H")
	if err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}
	if !strings.Contains(out, "NAME") || !strings.Contains(out, "LINE") || !strings.Contains(out, "TIME") {
		t.Errorf("-H out = %q, want heading", out)
	}
	if !strings.Contains(out, "alice") {
		t.Errorf("-H out = %q, want user line too", out)
	}
}

func TestWhoMissingFile(t *testing.T) {
	out, errOut, err := run(t, "/no/such/utmp")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	if out != "" {
		t.Errorf("out = %q, want empty", out)
	}
	if !strings.Contains(errOut, "who:") {
		t.Errorf("stderr = %q, want who: prefix", errOut)
	}
}

func TestWhoHelp(t *testing.T) {
	out, _, err := run(t, "", "--help")
	if err != nil {
		t.Fatalf("--help error = %v", err)
	}
	if !strings.Contains(out, "Usage: who") {
		t.Errorf("--help out = %q", out)
	}
}

func TestHelpSections(t *testing.T) {
	t.Parallel()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := who.New().Run(context.Background(), io, []string{"--help"}); err != nil {
		t.Fatalf("help err = %v", err)
	}
	for _, want := range []string{"Examples:", "Exit status:"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("--help output missing %q:\n%s", want, out.String())
		}
	}
}
