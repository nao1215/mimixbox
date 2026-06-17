// Package comp's tests exercise the shared compressor/decompressor frontend
// (Run, runStream, processFile) with a trivial in-package codec so they never
// depend on a real gzip/xz/bzip2 implementation. The codec is the identity
// transform: it copies bytes through unchanged, which is enough to verify all
// of the stream/file plumbing the frontend owns (-c, -k, -f, optional -t,
// stdin/stdout mode, mixed "-" operands, per-file error aggregation,
// partial-output cleanup, and the default/overridden ExistsErr and WrapFileErr
// paths).
package comp

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

const testSuffix = ".id"

// identityTransform copies r to w verbatim. The decompress flag is ignored: the
// identity codec is its own inverse, so the frontend's wiring is what we test,
// not the bytes.
func identityTransform(r io.Reader, w io.Writer, _ bool) error {
	_, err := io.Copy(w, r)
	return err
}

// failingTransform always fails after writing a few bytes, so tests can observe
// partial-output cleanup behavior.
func failingTransform(_ io.Reader, w io.Writer, _ bool) error {
	_, _ = io.WriteString(w, "partial")
	return errors.New("boom")
}

// suffixOutputName adds testSuffix when compressing and strips it when
// decompressing, returning an error for an unknown suffix on decompress.
func suffixOutputName(name string, decompress bool) (string, error) {
	if decompress {
		if !strings.HasSuffix(name, testSuffix) {
			return "", fmt.Errorf("%s: unknown suffix", name)
		}
		return strings.TrimSuffix(name, testSuffix), nil
	}
	return name + testSuffix, nil
}

// newStdio builds a command.IO backed by in-memory buffers.
func newStdio(in string) (command.IO, *bytes.Buffer, *bytes.Buffer) {
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	return command.IO{In: strings.NewReader(in), Out: out, Err: errBuf}, out, errBuf
}

// baseConfig returns the identity-codec Config every happy-path test starts from.
func baseConfig() *Config {
	return &Config{
		Name:       "idcomp",
		Transform:  identityTransform,
		OutputName: suffixOutputName,
	}
}

// writeTempFile creates a file with content under dir and returns its path.
func writeTempFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile(%s): %v", path, err)
	}
	return path
}

func isSilentFailure(err error) bool {
	if err == nil {
		return false
	}
	return err.Error() == "" // command.SilentFailure carries an empty message.
}

// --- runStream / stdin-stdout mode -----------------------------------------

func TestRun_StdinStdout(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		files []string
		opts  Options
	}{
		{name: "no files", files: nil},
		{name: "empty slice", files: []string{}},
		{name: "single dash", files: []string{"-"}},
		{name: "single dash decompress", files: []string{"-"}, opts: Options{Decompress: true}},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			stdio, out, errBuf := newStdio("hello stream")
			cfg := baseConfig()
			if err := cfg.Run(stdio, tt.opts, tt.files); err != nil {
				t.Fatalf("Run returned error: %v", err)
			}
			if out.String() != "hello stream" {
				t.Errorf("stdout = %q, want %q", out.String(), "hello stream")
			}
			if errBuf.Len() != 0 {
				t.Errorf("stderr = %q, want empty", errBuf.String())
			}
		})
	}
}

func TestRun_StreamTransformError(t *testing.T) {
	t.Parallel()
	stdio, _, errBuf := newStdio("data")
	cfg := baseConfig()
	cfg.Transform = failingTransform

	err := cfg.Run(stdio, Options{}, nil)
	if !isSilentFailure(err) {
		t.Fatalf("Run error = %v, want SilentFailure", err)
	}
	if !strings.Contains(errBuf.String(), "idcomp: boom") {
		t.Errorf("stderr = %q, want it to mention the transform error", errBuf.String())
	}
}

func TestRunStream_TestModeSuccess(t *testing.T) {
	t.Parallel()
	stdio, out, errBuf := newStdio("anything")
	cfg := baseConfig()
	var sawTest bool
	cfg.Test = func(r io.Reader) error {
		sawTest = true
		_, _ = io.Copy(io.Discard, r)
		return nil
	}

	if err := cfg.Run(stdio, Options{Test: true}, nil); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if !sawTest {
		t.Error("Test func was not called in -t stream mode")
	}
	if out.Len() != 0 {
		t.Errorf("stdout = %q, want empty in test mode", out.String())
	}
	if errBuf.Len() != 0 {
		t.Errorf("stderr = %q, want empty", errBuf.String())
	}
}

func TestRunStream_TestModeFailure(t *testing.T) {
	t.Parallel()
	stdio, _, errBuf := newStdio("anything")
	cfg := baseConfig()
	cfg.Test = func(_ io.Reader) error { return errors.New("corrupt") }

	err := cfg.Run(stdio, Options{Test: true}, nil)
	if !isSilentFailure(err) {
		t.Fatalf("Run error = %v, want SilentFailure", err)
	}
	if !strings.Contains(errBuf.String(), "idcomp: corrupt") {
		t.Errorf("stderr = %q, want it to mention the test error", errBuf.String())
	}
}

func TestRunStream_TestOptionWithoutTestFunc(t *testing.T) {
	t.Parallel()
	// -t requested but the codec has no Test func: the frontend must fall
	// through to a normal transform rather than calling a nil Test.
	stdio, out, _ := newStdio("payload")
	cfg := baseConfig() // Test is nil

	if err := cfg.Run(stdio, Options{Test: true}, nil); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if out.String() != "payload" {
		t.Errorf("stdout = %q, want %q", out.String(), "payload")
	}
}

// --- processFile: stdout mode (-c) -----------------------------------------

func TestProcessFile_StdoutKeepsInput(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := writeTempFile(t, dir, "a.txt", "body")
	stdio, out, errBuf := newStdio("")
	cfg := baseConfig()

	if err := cfg.Run(stdio, Options{Stdout: true}, []string{src}); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if out.String() != "body" {
		t.Errorf("stdout = %q, want %q", out.String(), "body")
	}
	if errBuf.Len() != 0 {
		t.Errorf("stderr = %q, want empty", errBuf.String())
	}
	if _, err := os.Stat(src); err != nil {
		t.Errorf("input file should remain with -c: %v", err)
	}
}

func TestProcessFile_StdoutMissingFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	missing := filepath.Join(dir, "nope.txt")
	stdio, _, errBuf := newStdio("")
	cfg := baseConfig()

	err := cfg.Run(stdio, Options{Stdout: true}, []string{missing})
	if !isSilentFailure(err) {
		t.Fatalf("Run error = %v, want SilentFailure", err)
	}
	if !strings.Contains(errBuf.String(), "idcomp:") {
		t.Errorf("stderr = %q, want it to be prefixed with the applet name", errBuf.String())
	}
}

// --- processFile: in-place compress ----------------------------------------

func TestProcessFile_InPlaceCompressRemovesInput(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := writeTempFile(t, dir, "data.txt", "compress me")
	stdio, _, _ := newStdio("")
	cfg := baseConfig()

	if err := cfg.Run(stdio, Options{}, []string{src}); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	out := src + testSuffix
	got, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("expected output file %s: %v", out, err)
	}
	if string(got) != "compress me" {
		t.Errorf("output = %q, want %q", got, "compress me")
	}
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Errorf("input should be removed by default, stat err = %v", err)
	}
}

func TestProcessFile_InPlaceKeepInput(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := writeTempFile(t, dir, "data.txt", "keep me")
	stdio, _, _ := newStdio("")
	cfg := baseConfig()

	if err := cfg.Run(stdio, Options{Keep: true}, []string{src}); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if _, err := os.Stat(src); err != nil {
		t.Errorf("input should remain with -k: %v", err)
	}
	if _, err := os.Stat(src + testSuffix); err != nil {
		t.Errorf("output should exist: %v", err)
	}
}

func TestProcessFile_InPlaceDecompressStripsSuffix(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := writeTempFile(t, dir, "data.txt"+testSuffix, "decoded")
	stdio, _, _ := newStdio("")
	cfg := baseConfig()

	if err := cfg.Run(stdio, Options{Decompress: true}, []string{src}); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	out := filepath.Join(dir, "data.txt")
	got, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("expected de-suffixed output %s: %v", out, err)
	}
	if string(got) != "decoded" {
		t.Errorf("output = %q, want %q", got, "decoded")
	}
}

func TestProcessFile_OutputNameError(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	// Decompressing a file without the suffix makes suffixOutputName fail.
	src := writeTempFile(t, dir, "plain.txt", "x")
	stdio, _, errBuf := newStdio("")
	cfg := baseConfig()

	err := cfg.Run(stdio, Options{Decompress: true}, []string{src})
	if !isSilentFailure(err) {
		t.Fatalf("Run error = %v, want SilentFailure", err)
	}
	if !strings.Contains(errBuf.String(), "unknown suffix") {
		t.Errorf("stderr = %q, want it to mention the OutputName error", errBuf.String())
	}
	// The input must be untouched when OutputName rejects it.
	if _, err := os.Stat(src); err != nil {
		t.Errorf("input should remain on OutputName error: %v", err)
	}
}

// --- existing output: -f and ExistsErr -------------------------------------

func TestProcessFile_ExistsDefaultError(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := writeTempFile(t, dir, "data.txt", "new")
	existing := writeTempFile(t, dir, "data.txt"+testSuffix, "old")
	stdio, _, errBuf := newStdio("")
	cfg := baseConfig()

	err := cfg.Run(stdio, Options{}, []string{src})
	if !isSilentFailure(err) {
		t.Fatalf("Run error = %v, want SilentFailure", err)
	}
	if !strings.Contains(errBuf.String(), "already exists; use -f to overwrite") {
		t.Errorf("stderr = %q, want the default exists message", errBuf.String())
	}
	// Output must be left untouched (not overwritten) without -f.
	got, _ := os.ReadFile(existing)
	if string(got) != "old" {
		t.Errorf("existing output overwritten without -f: %q", got)
	}
	if _, err := os.Stat(src); err != nil {
		t.Errorf("input should remain when output exists: %v", err)
	}
}

func TestProcessFile_ExistsCustomError(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := writeTempFile(t, dir, "data.txt", "new")
	writeTempFile(t, dir, "data.txt"+testSuffix, "old")
	stdio, _, errBuf := newStdio("")
	cfg := baseConfig()
	cfg.ExistsErr = func(out string) error {
		return fmt.Errorf("custom-exists: %s", out)
	}

	err := cfg.Run(stdio, Options{}, []string{src})
	if !isSilentFailure(err) {
		t.Fatalf("Run error = %v, want SilentFailure", err)
	}
	if !strings.Contains(errBuf.String(), "custom-exists:") {
		t.Errorf("stderr = %q, want the overridden ExistsErr message", errBuf.String())
	}
}

func TestProcessFile_ForceOverwrites(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := writeTempFile(t, dir, "data.txt", "fresh")
	out := writeTempFile(t, dir, "data.txt"+testSuffix, "stale")
	stdio, _, _ := newStdio("")
	cfg := baseConfig()

	if err := cfg.Run(stdio, Options{Force: true, Keep: true}, []string{src}); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	got, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(got) != "fresh" {
		t.Errorf("output = %q, want it overwritten to %q", got, "fresh")
	}
}

// --- RemoveOutputOnError ----------------------------------------------------

func TestProcessFile_RemoveOutputOnError(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := writeTempFile(t, dir, "data.txt", "input")
	stdio, _, errBuf := newStdio("")
	cfg := baseConfig()
	cfg.Transform = failingTransform
	cfg.RemoveOutputOnError = true

	err := cfg.Run(stdio, Options{}, []string{src})
	if !isSilentFailure(err) {
		t.Fatalf("Run error = %v, want SilentFailure", err)
	}
	if !strings.Contains(errBuf.String(), "boom") {
		t.Errorf("stderr = %q, want the transform error", errBuf.String())
	}
	if _, err := os.Stat(src + testSuffix); !os.IsNotExist(err) {
		t.Errorf("partial output should be removed, stat err = %v", err)
	}
	// Input must be preserved on failure (it is only removed after success).
	if _, err := os.Stat(src); err != nil {
		t.Errorf("input should remain on transform failure: %v", err)
	}
}

func TestProcessFile_KeepsPartialOutputWhenNotConfigured(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := writeTempFile(t, dir, "data.txt", "input")
	stdio, _, _ := newStdio("")
	cfg := baseConfig()
	cfg.Transform = failingTransform
	// RemoveOutputOnError stays false.

	if err := cfg.Run(stdio, Options{}, []string{src}); !isSilentFailure(err) {
		t.Fatalf("Run error = %v, want SilentFailure", err)
	}
	got, err := os.ReadFile(src + testSuffix)
	if err != nil {
		t.Fatalf("partial output should remain when RemoveOutputOnError is false: %v", err)
	}
	if string(got) != "partial" {
		t.Errorf("partial output = %q, want %q", got, "partial")
	}
}

// --- WrapFileErr ------------------------------------------------------------

func TestProcessFile_WrapFileErr(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	missing := filepath.Join(dir, "ghost.txt")
	stdio, _, errBuf := newStdio("")
	cfg := baseConfig()
	cfg.WrapFileErr = func(name string, err error) error {
		return fmt.Errorf("wrapped[%s]: %w", filepath.Base(name), err)
	}

	err := cfg.Run(stdio, Options{Stdout: true}, []string{missing})
	if !isSilentFailure(err) {
		t.Fatalf("Run error = %v, want SilentFailure", err)
	}
	if !strings.Contains(errBuf.String(), "wrapped[ghost.txt]:") {
		t.Errorf("stderr = %q, want the WrapFileErr decoration", errBuf.String())
	}
}

// --- test mode against files -----------------------------------------------

func TestProcessFile_TestModeSuccess(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := writeTempFile(t, dir, "data.txt", "valid")
	stdio, out, errBuf := newStdio("")
	cfg := baseConfig()
	cfg.Test = func(r io.Reader) error {
		b, _ := io.ReadAll(r)
		if string(b) != "valid" {
			return errors.New("unexpected content")
		}
		return nil
	}

	if err := cfg.Run(stdio, Options{Test: true}, []string{src}); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if out.Len() != 0 {
		t.Errorf("stdout = %q, want empty in test mode", out.String())
	}
	if errBuf.Len() != 0 {
		t.Errorf("stderr = %q, want empty", errBuf.String())
	}
	// Test mode must not remove or rewrite the file.
	if _, err := os.Stat(src); err != nil {
		t.Errorf("input should remain in test mode: %v", err)
	}
}

func TestProcessFile_TestModeFailure(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := writeTempFile(t, dir, "data.txt", "x")
	stdio, _, errBuf := newStdio("")
	cfg := baseConfig()
	cfg.Test = func(_ io.Reader) error { return errors.New("bad stream") }

	err := cfg.Run(stdio, Options{Test: true}, []string{src})
	if !isSilentFailure(err) {
		t.Fatalf("Run error = %v, want SilentFailure", err)
	}
	if !strings.Contains(errBuf.String(), "bad stream") {
		t.Errorf("stderr = %q, want the test error", errBuf.String())
	}
}

func TestProcessFile_TestModeMissingFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	missing := filepath.Join(dir, "absent.txt")
	stdio, _, errBuf := newStdio("")
	cfg := baseConfig()
	cfg.Test = func(_ io.Reader) error { return nil }

	err := cfg.Run(stdio, Options{Test: true}, []string{missing})
	if !isSilentFailure(err) {
		t.Fatalf("Run error = %v, want SilentFailure", err)
	}
	if !strings.Contains(errBuf.String(), "idcomp:") {
		t.Errorf("stderr = %q, want the applet-prefixed open error", errBuf.String())
	}
}

// --- mixed operands and aggregated failures --------------------------------

func TestRun_MixedDashOperands(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := writeTempFile(t, dir, "f.txt", "FILE")
	stdio, out, errBuf := newStdio("STDIN")
	cfg := baseConfig()

	// A "-" operand among real files streams stdin to stdout; the real file is
	// written to stdout too because -c is set, so both land in `out` in order.
	if err := cfg.Run(stdio, Options{Stdout: true}, []string{"-", src}); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if out.String() != "STDINFILE" {
		t.Errorf("stdout = %q, want %q (stdin then file)", out.String(), "STDINFILE")
	}
	if errBuf.Len() != 0 {
		t.Errorf("stderr = %q, want empty", errBuf.String())
	}
}

func TestRun_DashStreamFailureSetsExitCode(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := writeTempFile(t, dir, "f.txt", "ok")
	stdio, _, _ := newStdio("data")
	cfg := baseConfig()
	cfg.Transform = failingTransform

	// With multiple operands a failing "-" stream must flip the aggregate exit
	// code without aborting the loop.
	err := cfg.Run(stdio, Options{Stdout: true}, []string{"-", src})
	if !isSilentFailure(err) {
		t.Fatalf("Run error = %v, want SilentFailure", err)
	}
}

func TestRun_AggregatesPerFileFailures(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	good := writeTempFile(t, dir, "good.txt", "good")
	missing := filepath.Join(dir, "missing.txt")
	good2 := writeTempFile(t, dir, "good2.txt", "good2")
	stdio, _, errBuf := newStdio("")
	cfg := baseConfig()

	// Process good, missing (fails), good2: the run keeps going past the
	// failure, reports it, and still returns a failure exit code.
	err := cfg.Run(stdio, Options{Keep: true}, []string{good, missing, good2})
	if !isSilentFailure(err) {
		t.Fatalf("Run error = %v, want SilentFailure", err)
	}
	if !strings.Contains(errBuf.String(), "idcomp:") {
		t.Errorf("stderr = %q, want the failure reported", errBuf.String())
	}
	// Both good files were still processed despite the middle failure.
	if _, err := os.Stat(good + testSuffix); err != nil {
		t.Errorf("good.txt should have been processed: %v", err)
	}
	if _, err := os.Stat(good2 + testSuffix); err != nil {
		t.Errorf("good2.txt should have been processed: %v", err)
	}
}

func TestRun_AllSucceedReturnsNil(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	a := writeTempFile(t, dir, "a.txt", "A")
	b := writeTempFile(t, dir, "b.txt", "B")
	stdio, _, errBuf := newStdio("")
	cfg := baseConfig()

	if err := cfg.Run(stdio, Options{Keep: true}, []string{a, b}); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if errBuf.Len() != 0 {
		t.Errorf("stderr = %q, want empty", errBuf.String())
	}
}

// --- existsErr helper directly ---------------------------------------------

func TestExistsErr_DefaultAndOverride(t *testing.T) {
	t.Parallel()
	cfg := baseConfig()
	if got := cfg.existsErr("out.gz").Error(); !strings.Contains(got, "out.gz already exists; use -f to overwrite") {
		t.Errorf("default existsErr = %q, want the standard message", got)
	}
	cfg.ExistsErr = func(out string) error { return fmt.Errorf("override %s", out) }
	if got := cfg.existsErr("out.gz").Error(); got != "override out.gz" {
		t.Errorf("overridden existsErr = %q, want %q", got, "override out.gz")
	}
}
