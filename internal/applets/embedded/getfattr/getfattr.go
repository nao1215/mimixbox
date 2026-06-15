// Package getfattr implements the getfattr applet: display the extended
// attributes (xattrs) of files.
package getfattr

import (
	"context"
	"encoding/base64"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the getfattr applet.
type Command struct{}

// New returns a getfattr command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "getfattr" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Get extended attributes of files" }

// xattrBackend reads extended attributes. It is a package variable so tests can
// inject a fake without touching a real filesystem.
var xattrBackend Backend = osBackend{}

// Backend abstracts the host xattr syscalls so the parsing and formatting logic
// can be unit tested hermetically.
type Backend interface {
	// List returns the names of the extended attributes on path. follow
	// controls whether a symlink is dereferenced (true) or operated on
	// directly (false).
	List(path string, follow bool) ([]string, error)
	// Get returns the value of the named extended attribute on path.
	Get(path, name string, follow bool) ([]byte, error)
}

// Run executes getfattr.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-h] [-d] [-n name] [-e en] [-m pattern] FILE...", stdio.Err).WithHelp(command.Help{
		Description: "Print the extended attributes of each FILE. By default the names of all attributes in " +
			"the 'user' namespace are listed; with -d their values are dumped too. -n NAME prints a single " +
			"named attribute. -e selects the value encoding: text (default), hex, or base64. -h operates on " +
			"a symbolic link itself instead of the file it points to. This is a read-only command; it never " +
			"modifies any attribute.",
		Examples: []command.Example{
			{Command: "getfattr file.txt", Explain: "List the user.* attribute names of file.txt."},
			{Command: "getfattr -d file.txt", Explain: "Dump every user.* attribute name and value."},
			{Command: "getfattr -n user.demo file.txt", Explain: "Print one named attribute."},
			{Command: "getfattr -d -e hex file.txt", Explain: "Dump values hex-encoded."},
		},
		ExitStatus: "0  all files were read.\n1  a file could not be read or an option was invalid.",
		Notes: []string{
			"Attributes whose namespace is not 'user' are hidden unless -m matches them or -n names them.",
			"On a filesystem mounted without xattr support, reads fail with a documented error.",
		},
	})
	dump := fs.BoolP("dump", "d", false, "dump the values of all matched attributes")
	name := fs.StringP("name", "n", "", "print the value of the single attribute NAME")
	encoding := fs.StringP("encoding", "e", "text", "encode values as text, hex, or base64")
	match := fs.StringP("match", "m", "", "only attributes whose name contains this substring")
	noDeref := fs.BoolP("no-dereference", "h", false, "act on a symlink itself, not its target")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	files := fs.Args()
	if len(files) == 0 {
		return command.Failuref("at least one file operand is required")
	}
	enc, err := parseEncoding(*encoding)
	if err != nil {
		return command.Failuref("%v", err)
	}
	follow := !*noDeref

	failed := false
	for _, file := range files {
		if err := c.dumpFile(stdio, file, opts{dump: *dump || *name != "", name: *name, match: *match, enc: enc, follow: follow}); err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "getfattr: %s\n", command.FileError(file, err))
			failed = true
		}
	}
	if failed {
		return command.SilentFailure()
	}
	return nil
}

// opts collects the per-file dump options.
type opts struct {
	dump   bool
	name   string
	match  string
	enc    encoding
	follow bool
}

// dumpFile writes the attributes of one file in the getfattr text format.
func (c *Command) dumpFile(stdio command.IO, file string, o opts) error {
	var names []string
	if o.name != "" {
		names = []string{o.name}
	} else {
		all, err := xattrBackend.List(file, o.follow)
		if err != nil {
			return err
		}
		names = filterNames(all, o.match)
		sort.Strings(names)
	}
	if len(names) == 0 {
		return nil
	}

	var b strings.Builder
	fmt.Fprintf(&b, "# file: %s\n", file)
	for _, n := range names {
		if !o.dump {
			fmt.Fprintf(&b, "%s\n", n)
			continue
		}
		val, err := xattrBackend.Get(file, n, o.follow)
		if err != nil {
			return fmt.Errorf("%s: %w", n, err)
		}
		fmt.Fprintf(&b, "%s=%s\n", n, o.enc.encode(val))
	}
	b.WriteByte('\n')
	_, _ = fmt.Fprint(stdio.Out, b.String())
	return nil
}

// filterNames keeps only attributes that match the request: when match is
// empty, only the user namespace is shown (getfattr's default); otherwise any
// attribute whose name contains the substring.
func filterNames(all []string, match string) []string {
	out := make([]string, 0, len(all))
	for _, n := range all {
		if match != "" {
			if strings.Contains(n, match) {
				out = append(out, n)
			}
			continue
		}
		if strings.HasPrefix(n, "user.") {
			out = append(out, n)
		}
	}
	return out
}

// encoding selects how an attribute value is rendered.
type encoding int

const (
	encText encoding = iota
	encHex
	encBase64
)

// parseEncoding maps the -e flag to an encoding.
func parseEncoding(s string) (encoding, error) {
	switch s {
	case "text", "t":
		return encText, nil
	case "hex", "x":
		return encHex, nil
	case "base64", "b":
		return encBase64, nil
	default:
		return encText, fmt.Errorf("unknown encoding: %q (want text, hex, or base64)", s)
	}
}

// encode renders a raw attribute value in the chosen encoding, matching
// getfattr's output: text values are double-quoted, hex is "0x...", base64 is
// "0s...".
func (e encoding) encode(val []byte) string {
	switch e {
	case encHex:
		var b strings.Builder
		b.WriteString("0x")
		for _, c := range val {
			fmt.Fprintf(&b, "%02x", c)
		}
		return b.String()
	case encBase64:
		return "0s" + base64.StdEncoding.EncodeToString(val)
	default:
		return strconv.Quote(string(val))
	}
}
