package applets

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"

	"pgregory.net/rapid"
)

// gnuCommand finds GNU coreutils' name: the plain name where coreutils is
// GNU (Debian, Ubuntu up to 25.04, the CI runners), or gnu<name> where
// another implementation took the plain name (uutils on Ubuntu 26.04 keeps
// GNU there). It returns "" when neither is GNU.
func gnuCommand(name string) string {
	gnuOnce.Do(func() { gnuPaths = map[string]string{} })
	gnuMu.Lock()
	defer gnuMu.Unlock()
	if p, ok := gnuPaths[name]; ok {
		return p
	}
	found := ""
	for _, candidate := range []string{"gnu" + name, name} {
		p, err := exec.LookPath(candidate)
		if err != nil {
			continue
		}
		out, err := exec.Command(p, "--version").Output()
		if err == nil && strings.Contains(firstLine(string(out)), "GNU coreutils") {
			found = p
			break
		}
	}
	gnuPaths[name] = found
	return found
}

var (
	gnuOnce  sync.Once
	gnuMu    sync.Mutex
	gnuPaths map[string]string
)

func firstLine(s string) string {
	line, _, _ := strings.Cut(s, "\n")
	return line
}

// runGNU runs GNU's command in the C locale on stdin and returns what it
// wrote to stdout and its exit status.
func runGNU(t *rapid.T, path, stdin string, args ...string) (string, int) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, path, args...)
	cmd.Env = []string{"LC_ALL=C", "PATH=/usr/bin:/bin"}
	cmd.Stdin = strings.NewReader(stdin)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	var exitErr *exec.ExitError
	switch {
	case err == nil:
		return out.String(), 0
	case errors.As(err, &exitErr):
		return out.String(), exitErr.ExitCode()
	default:
		t.Fatalf("running %s: %v", path, err)
		return "", 0
	}
}

// textInput draws text of a few lines from an alphabet that holds what the
// line tools treat specially: separators, tabs, digits, case, repeated
// lines, empty lines, and a last line with or without its newline. With
// ascii false it also holds a two-byte character.
func textInput(ascii bool) *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		pieces := []string{"a", "b", "B", "z", "0", "7", "10", "-", " ", "  ", "\t", ":", ",", ".", "x:y"}
		if !ascii {
			pieces = append(pieces, "é")
		}
		lines := rapid.SliceOfN(rapid.SliceOfN(rapid.SampledFrom(pieces), 0, 6), 0, 8).Draw(t, "lines")
		var b strings.Builder
		for i, line := range lines {
			b.WriteString(strings.Join(line, ""))
			if i < len(lines)-1 || rapid.Bool().Draw(t, "final newline") {
				b.WriteByte('\n')
			}
		}
		return b.String()
	})
}

// The applets below claim to do what GNU coreutils does, and for these
// options there is one answer. On generated input, each must write what
// GNU writes and exit as GNU exits. Where GNU is not installed the case is
// skipped, so the test says nothing rather than something false.
func TestAppletsMatchGNUCoreutils(t *testing.T) {
	t.Parallel()
	// ascii marks the cases where GNU counts bytes in the C locale and
	// mimixbox counts characters (cut -c, expand). That is a choice about
	// multibyte text rather than a bug, so those cases are compared on
	// ASCII input only.
	tests := []struct {
		name  string
		args  []string
		ascii bool
	}{
		{"cat", nil, false},
		{"cat", []string{"-n"}, false},
		{"tac", nil, false},
		{"head", []string{"-n", "2"}, false},
		{"tail", []string{"-n", "2"}, false},
		{"sort", nil, false},
		{"sort", []string{"-r"}, false},
		{"sort", []string{"-u"}, false},
		{"uniq", nil, false},
		{"uniq", []string{"-c"}, false},
		{"wc", []string{"-l"}, false},
		{"wc", []string{"-c"}, false},
		{"cut", []string{"-d", ":", "-f", "2"}, false},
		{"cut", []string{"-c", "1-3"}, true},
		{"tr", []string{"a-z", "A-Z"}, false},
		{"fold", []string{"-w", "4"}, false},
		{"base64", nil, false},
		{"md5sum", nil, false},
		{"sha256sum", nil, false},
		{"nl", nil, false},
		{"expand", nil, true},
		{"sort", []string{"-n"}, false},
		{"sort", []string{"-k", "2"}, false},
		{"uniq", []string{"-d"}, false},
		{"uniq", []string{"-u"}, false},
		{"wc", nil, false},
		{"wc", []string{"-w"}, false},
		{"head", []string{"-c", "3"}, false},
		{"tail", []string{"-c", "3"}, false},
		{"cut", []string{"-f", "1"}, false},
		{"tr", []string{"-d", "a"}, false},
		{"tr", []string{"-s", " "}, false},
		{"paste", []string{"-s"}, false},
		{"cksum", nil, false},
		{"sum", nil, false},
		{"sha1sum", nil, false},
		{"base32", nil, false},
		{"expand", []string{"-t", "4"}, true},
		{"unexpand", []string{"-a"}, true},
		{"fold", []string{"-s", "-w", "3"}, false},
		{"nl", []string{"-b", "a"}, false},
		{"cat", []string{"-A"}, true},
		{"cat", []string{"-s"}, false},
		{"tac", []string{"-s", ":"}, false},
		{"od", []string{"-c"}, true},
		{"seq", []string{"3"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name+" "+strings.Join(tt.args, " "), func(t *testing.T) {
			t.Parallel()
			gnu := gnuCommand(tt.name)
			if gnu == "" {
				t.Skipf("GNU %s is not installed", tt.name)
			}
			rapid.Check(t, func(rt *rapid.T) {
				in := textInput(tt.ascii).Draw(rt, "stdin")
				want, wantCode := runGNU(rt, gnu, in, tt.args...)
				got, _, gotCode := runApplet(t, tt.name, in, tt.args...)
				if got != want || gotCode != wantCode {
					rt.Fatalf("%s %v on %q:\n mimixbox (exit %d): %q\n GNU      (exit %d): %q", tt.name, tt.args, in, gotCode, got, wantCode, want)
				}
			})
		})
	}
}
