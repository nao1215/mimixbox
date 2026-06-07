// Package mktemp implements the mktemp applet: create a uniquely-named
// temporary file or directory and print its path.
package mktemp

import (
	"context"
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// defaultTemplate is used when no TEMPLATE operand is given, matching GNU
// mktemp.
const defaultTemplate = "tmp.XXXXXXXXXX"

// randChars are the characters used to fill the X placeholders, matching the
// alphabet GNU mktemp draws from.
const randChars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// Command is the mktemp applet.
type Command struct{}

// New returns a mktemp command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "mktemp" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Create a temporary file or directory" }

// Run executes mktemp.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... [TEMPLATE]", stdio.Err)
	directory := fs.BoolP("directory", "d", false, "create a directory, not a file")
	dryRun := fs.BoolP("dry-run", "u", false, "do not create anything; merely print a name (unsafe)")
	quiet := fs.BoolP("quiet", "q", false, "suppress diagnostics about file/dir-creation failure")
	tmpdir := fs.StringP("tmpdir", "p", "", "interpret TEMPLATE relative to DIR; default is $TMPDIR or /tmp")
	tFlag := fs.BoolP("legacy", "t", false, "interpret TEMPLATE as a single file name component, relative to a temporary directory")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	operands := fs.Args()
	if len(operands) > 1 {
		return failf(stdio, *quiet, "too many templates")
	}

	template := defaultTemplate
	if len(operands) == 1 {
		template = operands[0]
	}

	// Whether the user explicitly asked for a tmpdir-relative template.
	tmpdirGiven := fs.Changed("tmpdir")
	dir := strings.TrimSpace(*tmpdir)
	if dir == "" {
		dir = tempDir()
	}

	path, err := resolve(template, dir, tmpdirGiven, *tFlag)
	if err != nil {
		return failf(stdio, *quiet, "%v", err)
	}

	result, err := generate(path, *directory, *dryRun)
	if err != nil {
		return failf(stdio, *quiet, "%v", err)
	}

	_, _ = fmt.Fprintln(stdio.Out, result)
	return nil
}

// tempDir returns the directory temporary files are created in: $TMPDIR when
// set, otherwise /tmp (matching GNU mktemp rather than os.TempDir, which is
// effectively the same but documented here for clarity).
func tempDir() string {
	if d := os.Getenv("TMPDIR"); d != "" {
		return d
	}
	return "/tmp"
}

// resolve turns the user's TEMPLATE plus the relevant options into the full
// path whose X's will be replaced. It also enforces that the template carries
// at least three trailing X's.
func resolve(template, dir string, tmpdirGiven, tFlag bool) (string, error) {
	switch {
	case tmpdirGiven || tFlag:
		// TEMPLATE is a single component relative to dir; it must not contain
		// a path separator (GNU rejects that with -p/-t).
		if strings.ContainsRune(template, os.PathSeparator) {
			return "", fmt.Errorf("invalid template, %q, contains directory separator", template)
		}
		return filepath.Join(dir, template), nil
	default:
		// Plain form: TEMPLATE is used as-is (relative to the cwd unless it is
		// already absolute).
		return template, nil
	}
}

// generate replaces the trailing run of X's in path with random characters,
// retrying until it finds a name that does not yet exist. With dryRun it only
// computes a name; otherwise it creates the file or directory atomically.
func generate(path string, directory, dryRun bool) (string, error) {
	prefix, suffix, xs, err := splitTemplate(path)
	if err != nil {
		return "", err
	}

	if dryRun {
		name, derr := randomName(prefix, suffix, xs)
		if derr != nil {
			return "", derr
		}
		return name, nil
	}

	// Retry a bounded number of times to avoid an unbounded loop if something
	// is wrong with the directory; GNU mktemp likewise gives up eventually.
	const maxAttempts = 1000
	for i := 0; i < maxAttempts; i++ {
		name, derr := randomName(prefix, suffix, xs)
		if derr != nil {
			return "", derr
		}
		if directory {
			if derr := os.Mkdir(name, 0o700); derr != nil {
				if os.IsExist(derr) {
					continue
				}
				return "", fmt.Errorf("failed to create directory %q: %w", name, derr)
			}
			return name, nil
		}
		f, derr := os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0o600)
		if derr != nil {
			if os.IsExist(derr) {
				continue
			}
			return "", fmt.Errorf("failed to create file %q: %w", name, derr)
		}
		_ = f.Close()
		return name, nil
	}
	return "", fmt.Errorf("failed to create file via template %q", path)
}

// splitTemplate locates the trailing run of X's in path, returning the text
// before it, the text after it, and the number of X's. GNU mktemp requires the
// X's to be at the end of the final path component and at least three of them.
func splitTemplate(path string) (prefix, suffix string, xs int, err error) {
	base := filepath.Base(path)
	dir := path[:len(path)-len(base)]

	// Find the run of X's that ends the base name.
	end := len(base)
	for end > 0 && base[end-1] == 'X' {
		end--
	}
	xs = len(base) - end
	if xs < 3 {
		return "", "", 0, fmt.Errorf("too few X's in template %q", path)
	}
	prefix = dir + base[:end]
	suffix = ""
	return prefix, suffix, xs, nil
}

// randomName fills the X placeholders with cryptographically random characters
// and returns the full candidate name.
func randomName(prefix, suffix string, xs int) (string, error) {
	buf := make([]byte, xs)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("failed to read random bytes: %w", err)
	}
	var b strings.Builder
	b.WriteString(prefix)
	for _, v := range buf {
		b.WriteByte(randChars[int(v)%len(randChars)])
	}
	b.WriteString(suffix)
	return b.String(), nil
}

// failf prints "mktemp: <message>" to stderr unless quiet, then returns a
// silent failure so the runner exits non-zero without re-printing.
func failf(stdio command.IO, quiet bool, format string, a ...any) error {
	if !quiet {
		_, _ = fmt.Fprintf(stdio.Err, "mktemp: "+format+"\n", a...)
	}
	return command.SilentFailure()
}
