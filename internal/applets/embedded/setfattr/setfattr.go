// Package setfattr implements the setfattr applet: set or remove the extended
// attributes (xattrs) of files.
package setfattr

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the setfattr applet.
type Command struct{}

// New returns a setfattr command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "setfattr" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Set extended attributes of files" }

// xattrBackend writes extended attributes. Tests inject a fake.
var xattrBackend Backend = osBackend{}

// Backend abstracts the host xattr syscalls so command planning can be unit
// tested hermetically.
type Backend interface {
	// Set stores value under name on path. follow controls symlink handling.
	Set(path, name string, value []byte, follow bool) error
	// Remove deletes the named attribute from path.
	Remove(path, name string, follow bool) error
}

// action is one planned mutation: set a value, or remove an attribute.
type action struct {
	remove bool
	name   string
	value  []byte
}

// Run executes setfattr.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-h] {-n name -v value | -x name} FILE...", stdio.Err).WithHelp(command.Help{
		Description: "Set or remove an extended attribute on each FILE. -n NAME with -v VALUE sets the " +
			"attribute (an empty -v stores a zero-length value); -x NAME removes it. The value's encoding " +
			"is inferred from its prefix: 0x... is hex, 0s... is base64, anything else is literal text. " +
			"-h operates on a symbolic link itself instead of its target. This command MODIFIES files; the " +
			"removal form (-x) is destructive and cannot be undone.",
		Examples: []command.Example{
			{Command: "setfattr -n user.demo -v hello file.txt", Explain: "Set user.demo to the text 'hello'."},
			{Command: "setfattr -n user.bin -v 0xdeadbeef file.txt", Explain: "Set a binary value from hex."},
			{Command: "setfattr -x user.demo file.txt", Explain: "Remove the user.demo attribute (destructive)."},
		},
		ExitStatus: "0  every file was updated.\n1  a file could not be updated or the arguments were invalid.",
		Notes: []string{
			"Most filesystems only allow unprivileged writes to the 'user.*' namespace.",
			"On a filesystem mounted without xattr support, writes fail with a documented error.",
		},
	})
	name := fs.StringP("name", "n", "", "name of the attribute to set")
	value := fs.StringP("value", "v", "", "value for the -n attribute (prefix 0x/0s for hex/base64)")
	remove := fs.StringP("remove", "x", "", "name of the attribute to remove (destructive)")
	noDeref := fs.BoolP("no-dereference", "h", false, "act on a symlink itself, not its target")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	act, err := plan(*name, *value, *remove, fs.Changed("value"))
	if err != nil {
		return command.Failuref("%v", err)
	}
	files := fs.Args()
	if len(files) == 0 {
		return command.Failuref("at least one file operand is required")
	}
	follow := !*noDeref

	failed := false
	for _, file := range files {
		var aerr error
		if act.remove {
			aerr = xattrBackend.Remove(file, act.name, follow)
		} else {
			aerr = xattrBackend.Set(file, act.name, act.value, follow)
		}
		if aerr != nil {
			_, _ = fmt.Fprintf(stdio.Err, "setfattr: %s\n", command.FileError(file, aerr))
			failed = true
		}
	}
	if failed {
		return command.SilentFailure()
	}
	return nil
}

// plan validates the mutually exclusive option set and decodes the value,
// returning the single action to apply to every file. valueGiven distinguishes
// an explicit empty -v from a missing one.
func plan(name, value, remove string, valueGiven bool) (action, error) {
	switch {
	case remove != "" && name != "":
		return action{}, fmt.Errorf("-x and -n are mutually exclusive")
	case remove != "":
		return action{remove: true, name: remove}, nil
	case name == "":
		return action{}, fmt.Errorf("one of -n NAME or -x NAME is required")
	}
	if !valueGiven {
		return action{}, fmt.Errorf("-n requires a value; pass -v (use -v '' for an empty value)")
	}
	decoded, err := decodeValue(value)
	if err != nil {
		return action{}, err
	}
	return action{name: name, value: decoded}, nil
}

// decodeValue interprets the setfattr value encoding prefixes: 0x.. hex,
// 0s.. base64, otherwise the literal bytes of the string.
func decodeValue(v string) ([]byte, error) {
	switch {
	case strings.HasPrefix(v, "0x") || strings.HasPrefix(v, "0X"):
		b, err := hex.DecodeString(v[2:])
		if err != nil {
			return nil, fmt.Errorf("invalid hex value: %v", err)
		}
		return b, nil
	case strings.HasPrefix(v, "0s") || strings.HasPrefix(v, "0S"):
		b, err := base64.StdEncoding.DecodeString(v[2:])
		if err != nil {
			return nil, fmt.Errorf("invalid base64 value: %v", err)
		}
		return b, nil
	default:
		return []byte(v), nil
	}
}
