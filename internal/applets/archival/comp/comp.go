// Package comp is the shared compressor/decompressor frontend used by the
// single-stream archival applets (gzip/gunzip, bzip2/bunzip2, xz/lzma/...,
// lzop/..., compress/uncompress). They all follow the same file-handling model:
// read standard input to standard output when given no FILE (or "-"); otherwise
// process each FILE in place, adding or stripping a suffix, honoring -c (write
// to stdout, keep input), -k (keep input), -f (overwrite output) and, for the
// codecs that support it, -t (test integrity). This package factors that model
// out so each applet only supplies its codec and a small Config describing the
// few points where the applets legitimately differ.
package comp

import (
	"fmt"
	"io"
	"os"

	"github.com/nao1215/mimixbox/internal/command"
)

// Options captures the per-invocation flags shared by every compressor applet.
// Each applet parses its own flag set (so its help text and accepted spellings
// stay exactly as before) and then fills this struct.
type Options struct {
	// Decompress reverses the codec instead of compressing.
	Decompress bool
	// Stdout writes the result to standard output and keeps every input file
	// (the -c flag).
	Stdout bool
	// Keep leaves the input file in place after a successful in-place rewrite
	// (the -k flag); by default the input is removed.
	Keep bool
	// Force overwrites an existing output file (the -f flag); without it an
	// existing output is an error.
	Force bool
	// Test only verifies that each input decodes, writing nothing (the -t
	// flag). It is meaningful only when Test is configured on the frontend.
	Test bool
}

// Config wires an applet's codec and naming rules into the shared frontend.
// Only Transform and OutputName are required.
type Config struct {
	// Name is the applet name, used as the prefix of error messages.
	Name string

	// Transform copies r to w, compressing (decompress=false) or decompressing
	// (decompress=true) along the way.
	Transform func(r io.Reader, w io.Writer, decompress bool) error

	// Test verifies that r is a valid stream, writing nothing. It is nil for
	// codecs without a -t mode; when nil the frontend never calls it.
	Test func(r io.Reader) error

	// OutputName derives the output filename for an in-place rewrite of name:
	// the suffixed name when compressing, or the de-suffixed name when
	// decompressing. It returns an error (e.g. "unknown suffix") to skip a file.
	OutputName func(name string, decompress bool) (string, error)

	// RemoveOutputOnError, when true, deletes a partially written output file if
	// the transform fails, so a failed run leaves no truncated artifact behind.
	RemoveOutputOnError bool

	// ExistsErr formats the error returned when the output already exists and
	// -f was not given. When nil a default "<out> already exists; use -f to
	// overwrite" message is used.
	ExistsErr func(out string) error

	// WrapFileErr formats a per-file error before it is printed. When nil the
	// error is printed verbatim. gzip uses this to apply command.FileError.
	WrapFileErr func(name string, err error) error
}

// existsErr returns the configured (or default) "already exists" error.
func (cfg *Config) existsErr(out string) error {
	if cfg.ExistsErr != nil {
		return cfg.ExistsErr(out)
	}
	return fmt.Errorf("%s already exists; use -f to overwrite", out)
}

// Run is the shared entry point. files is fs.Args(); with none (or a single
// "-") it streams standard input to standard output, otherwise it processes
// each file. It prints failures on stderr and returns a silent failure that
// only sets the exit code, matching the per-applet behavior it replaces.
func (cfg *Config) Run(stdio command.IO, opts Options, files []string) error {
	if len(files) == 0 || (len(files) == 1 && files[0] == "-") {
		return cfg.runStream(stdio, opts)
	}

	var failed bool
	for _, name := range files {
		// A "-" operand among real files streams standard input to standard
		// output, just like the no-operand case.
		if name == "-" {
			if err := cfg.runStream(stdio, opts); err != nil {
				failed = true
			}
			continue
		}
		if err := cfg.processFile(stdio, name, opts); err != nil {
			if cfg.WrapFileErr != nil {
				err = cfg.WrapFileErr(name, err)
			}
			_, _ = fmt.Fprintf(stdio.Err, "%s: %v\n", cfg.Name, err)
			failed = true
		}
	}
	if failed {
		return command.SilentFailure()
	}
	return nil
}

// runStream handles the stdin/stdout (no FILE) case.
func (cfg *Config) runStream(stdio command.IO, opts Options) error {
	if opts.Test && cfg.Test != nil {
		if err := cfg.Test(stdio.In); err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "%s: %v\n", cfg.Name, err)
			return command.SilentFailure()
		}
		return nil
	}
	if err := cfg.Transform(stdio.In, stdio.Out, opts.Decompress); err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "%s: %v\n", cfg.Name, err)
		return command.SilentFailure()
	}
	return nil
}

// processFile compresses, decompresses or tests one named file: to stdout with
// -c, otherwise in place with the suffix added or stripped.
func (cfg *Config) processFile(stdio command.IO, name string, opts Options) error {
	if opts.Test && cfg.Test != nil {
		in, err := os.Open(name) //nolint:gosec // user-named file
		if err != nil {
			return err
		}
		defer func() { _ = in.Close() }()
		if err := cfg.Test(in); err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
		return nil
	}

	if opts.Stdout {
		in, err := os.Open(name) //nolint:gosec // user-named file
		if err != nil {
			return err
		}
		defer func() { _ = in.Close() }()
		return cfg.Transform(in, stdio.Out, opts.Decompress)
	}

	out, err := cfg.OutputName(name, opts.Decompress)
	if err != nil {
		return err
	}
	if !opts.Force {
		if _, statErr := os.Stat(out); statErr == nil {
			return cfg.existsErr(out)
		}
	}

	in, err := os.Open(name) //nolint:gosec // user-named file
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()

	w, err := os.Create(out) //nolint:gosec // user-named file
	if err != nil {
		return err
	}
	if err := cfg.Transform(in, w, opts.Decompress); err != nil {
		_ = w.Close()
		if cfg.RemoveOutputOnError {
			_ = os.Remove(out) // don't leave a partial output file behind on failure
		}
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	if !opts.Keep {
		return os.Remove(name)
	}
	return nil
}
