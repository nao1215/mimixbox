// Package banner implements the banner applet: print its argument as large
// ASCII-art letters built from '#' characters.
package banner

import (
	"context"
	"io"
	"strings"
	"unicode"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the banner applet.
type Command struct{}

// New returns a banner command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "banner" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Print a string as large ASCII-art letters" }

// glyphHeight is the number of rows every glyph occupies.
const glyphHeight = 5

// Run executes banner.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "MESSAGE...", stdio.Err).WithHelp(command.Help{
		Description: "Print MESSAGE as large ASCII-art letters built from '#' characters.\n" +
			"Multiple operands are joined with spaces, lowercase letters are\n" +
			"upper-cased, and unknown characters are rendered as blank glyphs.",
		Examples: []command.Example{
			{Command: "banner HI", Explain: "print HI as ASCII-art letters"},
			{Command: "banner hello world", Explain: "join the words and print them large"},
		},
		ExitStatus: "0  success.\n1  no message operand was given.",
	})

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	message := strings.Join(fs.Args(), " ")
	if message == "" {
		return command.Failuref("missing message operand")
	}

	if _, err := io.WriteString(stdio.Out, Render(message)); err != nil {
		return command.Failure(err)
	}
	return nil
}

// Render returns the banner art for message: glyphHeight rows of '#'-art, one
// space column between glyphs. Lowercase letters are upper-cased; unknown runes
// are rendered as blank glyphs.
func Render(message string) string {
	rows := make([]strings.Builder, glyphHeight)
	for i, r := range []rune(strings.ToUpper(message)) {
		glyph, ok := font[r]
		if !ok {
			glyph = blank
		}
		for row := 0; row < glyphHeight; row++ {
			if i > 0 {
				rows[row].WriteByte(' ')
			}
			rows[row].WriteString(glyph[row])
		}
	}

	var b strings.Builder
	for row := 0; row < glyphHeight; row++ {
		b.WriteString(strings.TrimRight(rows[row].String(), " "))
		b.WriteByte('\n')
	}
	return b.String()
}

// blank is the glyph used for runes with no defined art.
var blank = [glyphHeight]string{"     ", "     ", "     ", "     ", "     "}

// upperLetters reports whether r is a rune the font defines (for tests).
func defined(r rune) bool {
	_, ok := font[unicode.ToUpper(r)]
	return ok
}

// font maps each supported rune to its 5x5 '#'-art glyph.
var font = map[rune][glyphHeight]string{
	' ': {"     ", "     ", "     ", "     ", "     "},
	'A': {" ### ", "#   #", "#####", "#   #", "#   #"},
	'B': {"#### ", "#   #", "#### ", "#   #", "#### "},
	'C': {" ####", "#    ", "#    ", "#    ", " ####"},
	'D': {"#### ", "#   #", "#   #", "#   #", "#### "},
	'E': {"#####", "#    ", "#### ", "#    ", "#####"},
	'F': {"#####", "#    ", "#### ", "#    ", "#    "},
	'G': {" ####", "#    ", "#  ##", "#   #", " ####"},
	'H': {"#   #", "#   #", "#####", "#   #", "#   #"},
	'I': {"#####", "  #  ", "  #  ", "  #  ", "#####"},
	'J': {"#####", "   # ", "   # ", "#  # ", " ##  "},
	'K': {"#   #", "#  # ", "###  ", "#  # ", "#   #"},
	'L': {"#    ", "#    ", "#    ", "#    ", "#####"},
	'M': {"#   #", "## ##", "# # #", "#   #", "#   #"},
	'N': {"#   #", "##  #", "# # #", "#  ##", "#   #"},
	'O': {" ### ", "#   #", "#   #", "#   #", " ### "},
	'P': {"#### ", "#   #", "#### ", "#    ", "#    "},
	'Q': {" ### ", "#   #", "# # #", "#  # ", " ## #"},
	'R': {"#### ", "#   #", "#### ", "#  # ", "#   #"},
	'S': {" ####", "#    ", " ### ", "    #", "#### "},
	'T': {"#####", "  #  ", "  #  ", "  #  ", "  #  "},
	'U': {"#   #", "#   #", "#   #", "#   #", " ### "},
	'V': {"#   #", "#   #", "#   #", " # # ", "  #  "},
	'W': {"#   #", "#   #", "# # #", "## ##", "#   #"},
	'X': {"#   #", " # # ", "  #  ", " # # ", "#   #"},
	'Y': {"#   #", " # # ", "  #  ", "  #  ", "  #  "},
	'Z': {"#####", "   # ", "  #  ", " #   ", "#####"},
	'0': {" ### ", "#  ##", "# # #", "##  #", " ### "},
	'1': {"  #  ", " ##  ", "  #  ", "  #  ", "#####"},
	'2': {" ### ", "#   #", "  ## ", " #   ", "#####"},
	'3': {"#### ", "    #", " ### ", "    #", "#### "},
	'4': {"#  # ", "#  # ", "#####", "   # ", "   # "},
	'5': {"#####", "#    ", "#### ", "    #", "#### "},
	'6': {" ### ", "#    ", "#### ", "#   #", " ### "},
	'7': {"#####", "   # ", "  #  ", " #   ", " #   "},
	'8': {" ### ", "#   #", " ### ", "#   #", " ### "},
	'9': {" ### ", "#   #", " ####", "    #", " ### "},
}
