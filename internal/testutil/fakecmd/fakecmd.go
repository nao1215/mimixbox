// Package fakecmd provides repo-local fake implementations of the handful of
// external commands (echo, printf, true, false, sh, cat, wc) that several unit
// tests need to exec. Tests historically skipped when these commands were
// absent from the host PATH; that made the tests non-deterministic on
// stripped-down hosts and CI images. Instead of depending on the host, tests
// build one small helper binary (the program in helperSource) once per test
// process and expose it under whatever command names a test asks for, in a
// temporary directory that the test prepends to PATH.
//
// The helper is a single Go program that dispatches on the base name it was
// invoked as (argv[0]). This is the standard "multi-call" trick and keeps the
// fixtures portable: it never relies on a real /bin/sh, /bin/echo, etc.
package fakecmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
)

// helperSource is the source of the multi-call fake-command binary. It is
// written to a temp directory and compiled with `go build` once per test
// process. Keeping it as source (rather than committing a binary) means it is
// built for the running platform and stays reviewable.
const helperSource = `package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func main() {
	switch filepath.Base(os.Args[0]) {
	case "echo":
		os.Exit(doEcho(os.Args[1:]))
	case "printf":
		os.Exit(doPrintf(os.Args[1:]))
	case "true":
		os.Exit(0)
	case "false":
		os.Exit(1)
	case "cat":
		os.Exit(doCat(os.Args[1:]))
	case "wc":
		os.Exit(doWc(os.Args[1:]))
	case "sh":
		os.Exit(doSh(os.Args[1:]))
	default:
		fmt.Fprintf(os.Stderr, "fakecmd: unknown command %q\n", filepath.Base(os.Args[0]))
		os.Exit(127)
	}
}

// doEcho mimics POSIX echo: join args with spaces and append a newline. The
// -n flag suppresses the trailing newline.
func doEcho(args []string) int {
	newline := true
	if len(args) > 0 && args[0] == "-n" {
		newline = false
		args = args[1:]
	}
	fmt.Print(strings.Join(args, " "))
	if newline {
		fmt.Print("\n")
	}
	return 0
}

// doPrintf supports the subset of printf the tests use: the %s conversion and
// the common backslash escapes inside the format string. Extra arguments reuse
// the format, matching printf semantics for the cases exercised here.
func doPrintf(args []string) int {
	if len(args) == 0 {
		return 0
	}
	format := unescape(args[0])
	rest := args[1:]
	verbs := strings.Count(format, "%s")
	if verbs == 0 {
		fmt.Print(format)
		return 0
	}
	i := 0
	for {
		out := format
		for v := 0; v < verbs; v++ {
			arg := ""
			if i < len(rest) {
				arg = rest[i]
				i++
			}
			out = strings.Replace(out, "%s", arg, 1)
		}
		fmt.Print(out)
		if i >= len(rest) {
			break
		}
	}
	return 0
}

func unescape(s string) string {
	r := strings.NewReplacer(
		"\\n", "\n",
		"\\t", "\t",
		"\\r", "\r",
		"\\\\", "\\",
	)
	return r.Replace(s)
}

func doCat(args []string) int {
	if len(args) == 0 {
		_, _ = io.Copy(os.Stdout, os.Stdin)
		return 0
	}
	status := 0
	for _, name := range args {
		f, err := os.Open(name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "cat: %v\n", err)
			status = 1
			continue
		}
		_, _ = io.Copy(os.Stdout, f)
		_ = f.Close()
	}
	return status
}

// doWc supports -c (byte count), -l (line count) and -w (word count) on stdin
// or a single file, which is all the tests require.
func doWc(args []string) int {
	mode := ""
	var file string
	for _, a := range args {
		switch a {
		case "-c", "-l", "-w":
			mode = a
		default:
			file = a
		}
	}
	var reader io.Reader = os.Stdin
	if file != "" {
		f, err := os.Open(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "wc: %v\n", err)
			return 1
		}
		defer f.Close()
		reader = f
	}
	data, err := io.ReadAll(reader)
	if err != nil {
		fmt.Fprintf(os.Stderr, "wc: %v\n", err)
		return 1
	}
	var n int
	switch mode {
	case "-l":
		n = strings.Count(string(data), "\n")
	case "-w":
		n = len(strings.Fields(string(data)))
	default: // -c or unspecified
		n = len(data)
	}
	fmt.Printf("%d\n", n)
	return 0
}

// doSh implements a deliberately tiny "sh -c" good enough for the limited
// command lines the tests pass. It understands a handful of builtins and
// constructs: pwd -P, exit N, printf, echo, simple expansion of $VAR and $0.
// It is NOT a general shell; it exists so tests that previously required a
// host /bin/sh can run deterministically.
func doSh(args []string) int {
	// Parse: sh -c SCRIPT [arg0 arg1 ...]
	if len(args) < 2 || args[0] != "-c" {
		return 0
	}
	script := args[1]
	positional := args[2:]
	return runShScript(script, positional)
}

func runShScript(script string, positional []string) int {
	script = strings.TrimSpace(script)
	// Expand $0 .. $9 from the positional parameters.
	for idx := 0; idx <= 9; idx++ {
		val := ""
		if idx < len(positional) {
			val = positional[idx]
		}
		script = strings.ReplaceAll(script, "$"+strconv.Itoa(idx), val)
	}
	// Expand named environment variables of the form $NAME or ${NAME}.
	script = os.Expand(script, os.Getenv)

	switch {
	case script == "pwd -P" || script == "pwd":
		wd, _ := os.Getwd()
		fmt.Println(wd)
		return 0
	case strings.HasPrefix(script, "exit"):
		code := 0
		fields := strings.Fields(script)
		if len(fields) > 1 {
			if v, err := strconv.Atoi(fields[1]); err == nil {
				code = v
			}
		}
		return code
	case strings.HasPrefix(script, "printf "):
		return doPrintf(shFields(strings.TrimPrefix(script, "printf ")))
	case strings.HasPrefix(script, "echo "):
		return doEcho(shFields(strings.TrimPrefix(script, "echo ")))
	default:
		fmt.Fprintf(os.Stderr, "fakesh: unsupported script %q\n", script)
		return 127
	}
}

// shFields splits a command tail into fields, honoring single and double
// quotes so that printf 'a b' is one argument.
func shFields(s string) []string {
	var fields []string
	var cur strings.Builder
	inSingle, inDouble, started := false, false, false
	flush := func() {
		if started {
			fields = append(fields, cur.String())
			cur.Reset()
			started = false
		}
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c == '\'' && !inDouble:
			inSingle = !inSingle
			started = true
		case c == '"' && !inSingle:
			inDouble = !inDouble
			started = true
		case c == ' ' && !inSingle && !inDouble:
			flush()
		default:
			cur.WriteByte(c)
			started = true
		}
	}
	flush()
	return fields
}
`

var (
	buildOnce sync.Once
	buildPath string
	buildErr  error
)

// helperBinary builds (once per test process) the multi-call fake-command
// helper and returns the path to the compiled binary. Subsequent calls reuse
// the cached binary.
func helperBinary(t *testing.T) string {
	t.Helper()
	buildOnce.Do(func() {
		dir, err := os.MkdirTemp("", "mimixbox-fakecmd-")
		if err != nil {
			buildErr = err
			return
		}
		src := filepath.Join(dir, "main.go")
		if err := os.WriteFile(src, []byte(helperSource), 0o600); err != nil {
			buildErr = err
			return
		}
		bin := filepath.Join(dir, "fakecmd")
		if runtime.GOOS == "windows" {
			bin += ".exe"
		}
		cmd := exec.Command("go", "build", "-o", bin, src)
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			buildErr = err
			return
		}
		buildPath = bin
	})
	if buildErr != nil {
		t.Fatalf("fakecmd: building helper binary: %v", buildErr)
	}
	return buildPath
}

// Dir builds the fake-command helper and exposes it under each of the given
// command names in a fresh temporary directory, returning that directory. Add
// it to the front of PATH so the fakes shadow any host commands. The directory
// is removed automatically when the test ends.
//
// Supported names: echo, printf, true, false, cat, wc, sh. The sh fake only
// understands the small command set the tests use; see doSh in the helper.
func Dir(t *testing.T, names ...string) string {
	t.Helper()
	bin := helperBinary(t)
	dir := t.TempDir()
	for _, name := range names {
		target := filepath.Join(dir, name)
		if runtime.GOOS == "windows" {
			target += ".exe"
		}
		linkOrCopy(t, bin, target)
	}
	return dir
}

// Prepend returns a PATH value with dir at the front of the current PATH.
func Prepend(dir string) string {
	existing := os.Getenv("PATH")
	if existing == "" {
		return dir
	}
	return dir + string(os.PathListSeparator) + existing
}

// Use builds the fakes for the given command names and points the test's PATH
// at them (via t.Setenv, so it is restored automatically). Existing PATH
// entries are kept after the fakes so that already-available real commands
// still resolve. Returns the fake directory.
func Use(t *testing.T, names ...string) string {
	t.Helper()
	dir := Dir(t, names...)
	t.Setenv("PATH", Prepend(dir))
	return dir
}

// UseOnly points PATH only at the fake directory, hiding every host command.
// Use this when a test must prove it does not depend on host PATH at all.
func UseOnly(t *testing.T, names ...string) string {
	t.Helper()
	dir := Dir(t, names...)
	t.Setenv("PATH", dir)
	return dir
}

func linkOrCopy(t *testing.T, src, dst string) {
	t.Helper()
	if err := os.Link(src, dst); err == nil {
		return
	}
	if err := os.Symlink(src, dst); err == nil {
		return
	}
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatalf("fakecmd: reading helper: %v", err)
	}
	if err := os.WriteFile(dst, data, 0o700); err != nil {
		t.Fatalf("fakecmd: writing fake %q: %v", dst, err)
	}
}
