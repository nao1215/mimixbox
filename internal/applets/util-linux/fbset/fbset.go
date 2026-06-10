// Package fbset implements the fbset applet: show the current framebuffer video
// mode. Changing the mode is not done by this slice.
package fbset

import (
	"context"
	"encoding/binary"
	"fmt"
	"os"
	"unsafe"

	"github.com/nao1215/mimixbox/internal/command"
	"golang.org/x/sys/unix"
)

// Command is the fbset applet.
type Command struct{}

// New returns a fbset command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "fbset" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Show the framebuffer video mode" }

// fbioGetVScreenInfo is FBIOGET_VSCREENINFO, not exported by this x/sys version.
const fbioGetVScreenInfo = 0x4600

// varScreenInfo holds the fields of struct fb_var_screeninfo this applet reports.
type varScreenInfo struct {
	xres, yres, bpp uint32
}

// readVarFn is indirected so the formatting can be tested without a framebuffer.
var readVarFn = func(device string) (varScreenInfo, error) {
	f, err := os.Open(device) //nolint:gosec // user-named framebuffer device
	if err != nil {
		return varScreenInfo{}, err
	}
	defer func() { _ = f.Close() }()

	var buf [160]byte // struct fb_var_screeninfo is 160 bytes
	_, _, errno := unix.Syscall(unix.SYS_IOCTL, f.Fd(), fbioGetVScreenInfo, uintptr(unsafe.Pointer(&buf[0])))
	if errno != 0 {
		return varScreenInfo{}, errno
	}
	return varScreenInfo{
		xres: binary.LittleEndian.Uint32(buf[0:]),
		yres: binary.LittleEndian.Uint32(buf[4:]),
		bpp:  binary.LittleEndian.Uint32(buf[24:]), // bits_per_pixel
	}, nil
}

// Run executes fbset.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-fb DEVICE]", stdio.Err).WithHelp(command.Help{
		Description: "Show the current video mode of the framebuffer DEVICE (default /dev/fb0): its " +
			"resolution and color depth, in the usual mode/geometry/endmode block. Changing the mode " +
			"is not done by this build.",
		Examples: []command.Example{
			{Command: "fbset", Explain: "Show the /dev/fb0 video mode."},
			{Command: "fbset -fb /dev/fb1", Explain: "Show another framebuffer's mode."},
		},
		ExitStatus: "0  the mode was shown.\n1  the framebuffer could not be read.",
	})
	device := fs.String("fb", "/dev/fb0", "framebuffer device to query")

	// fbset uses the historical single-dash "-fb"; rewrite it to the long form so
	// pflag does not read it as a "-f -b" shorthand cluster.
	for i, a := range args {
		if a == "-fb" {
			args[i] = "--fb"
		}
	}

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	info, err := readVarFn(*device)
	if err != nil {
		return command.Failuref("%s: %v", *device, err)
	}

	_, _ = fmt.Fprintf(stdio.Out, "mode \"%dx%d\"\n", info.xres, info.yres)
	_, _ = fmt.Fprintf(stdio.Out, "    geometry %d %d %d %d %d\n", info.xres, info.yres, info.xres, info.yres, info.bpp)
	_, _ = fmt.Fprintln(stdio.Out, "endmode")
	return nil
}
