// Package patch implements the patch applet: apply a unified diff (as produced
// by "diff -u") to the target files. It reads the patch from a file given with
// -i or from standard input, supports -pN path stripping, -R reverse
// application and --dry-run.
package patch

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the patch applet.
type Command struct{}

// New returns a patch command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "patch" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Apply a diff file to an original" }

// Run executes patch.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... [ORIGFILE]", stdio.Err).WithHelp(command.Help{
		Description: "Apply a unified diff (as produced by \"diff -u\") to the target files. The patch is " +
			"read from the file given with -i, or from standard input when -i is omitted.",
		Examples: []command.Example{
			{Command: "patch -i changes.diff", Explain: "Apply the unified diff in changes.diff."},
			{Command: "patch -p1 -i changes.diff", Explain: "Strip one leading path component from file names."},
			{Command: "patch -R -i changes.diff", Explain: "Reverse a previously applied patch."},
		},
		ExitStatus: "0  all patches applied successfully.\n1  a patch could not be applied or the input was invalid.",
	})
	strip := fs.IntP("strip", "p", 0, "strip NUM leading components from file names")
	input := fs.StringP("input", "i", "", "read patch from FILE instead of stdin")
	reverse := fs.BoolP("reverse", "R", false, "assume patches were created with old and new swapped")
	dryRun := fs.Bool("dry-run", false, "print the results without changing any files")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	var patchText string
	if *input != "" {
		data, err := os.ReadFile(*input) //nolint:gosec // user-named patch file
		if err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "patch: %s\n", command.FileError(*input, err))
			return command.SilentFailure()
		}
		patchText = string(data)
	} else {
		var b strings.Builder
		sc := bufio.NewScanner(stdio.In)
		sc.Buffer(make([]byte, 0, 64*1024), command.MaxLineSize)
		for sc.Scan() {
			b.WriteString(sc.Text())
			b.WriteByte('\n')
		}
		patchText = b.String()
	}

	patches, err := parseUnified(patchText)
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "patch: %v\n", err)
		return command.SilentFailure()
	}
	if len(patches) == 0 {
		_, _ = fmt.Fprintln(stdio.Err, "patch: no valid patches in input")
		return command.SilentFailure()
	}

	// An explicit ORIGFILE operand overrides the file name in the patch (used
	// for single-file patches).
	override := ""
	if rest := fs.Args(); len(rest) > 0 {
		override = rest[0]
	}

	var failed bool
	for i := range patches {
		fp := &patches[i]
		if *reverse {
			reverseHunks(fp)
		}
		target := override
		if target == "" {
			target = pickName(fp, *strip, *reverse)
		}
		if err := applyFile(stdio, target, fp, *dryRun); err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "patch: %v\n", err)
			failed = true
			continue
		}
		if !*dryRun {
			_, _ = fmt.Fprintf(stdio.Out, "patching file %s\n", target)
		}
	}
	if failed {
		return command.SilentFailure()
	}
	return nil
}

// pickName chooses the target file name from the patch headers, applying -p
// stripping. For reverse patches the old name is the target.
func pickName(fp *filePatch, strip int, reverse bool) string {
	name := fp.newName
	if reverse {
		name = fp.oldName
	}
	return stripComponents(name, strip)
}

// stripComponents removes the first n path components from a patch file name.
func stripComponents(name string, n int) string {
	name = strings.TrimSpace(name)
	// Drop a trailing tab-separated timestamp if present.
	if i := strings.IndexByte(name, '\t'); i >= 0 {
		name = name[:i]
	}
	for i := 0; i < n; i++ {
		if idx := strings.IndexByte(name, '/'); idx >= 0 {
			name = name[idx+1:]
		}
	}
	return name
}

// applyFile applies a file patch to its target on disk.
func applyFile(stdio command.IO, target string, fp *filePatch, dryRun bool) error {
	data, err := os.ReadFile(target) //nolint:gosec // user-named target file
	if err != nil {
		return fmt.Errorf("cannot open %s", command.FileError(target, err))
	}
	orig := splitLinesKeep(string(data))

	result, err := applyHunks(orig, fp.hunks)
	if err != nil {
		return fmt.Errorf("%s: %v", target, err)
	}

	out := strings.Join(result, "\n")
	if len(result) > 0 {
		out += "\n"
	}
	if dryRun {
		_, _ = fmt.Fprintf(stdio.Out, "checking file %s\n", target)
		return nil
	}
	return os.WriteFile(target, []byte(out), 0o644) //nolint:gosec // preserve simple default mode
}

// splitLinesKeep splits text into lines, dropping a single trailing newline.
func splitLinesKeep(s string) []string {
	if s == "" {
		return nil
	}
	s = strings.TrimSuffix(s, "\n")
	return strings.Split(s, "\n")
}
