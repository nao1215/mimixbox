// Package fmt implements the fmt applet: reflow paragraphs of text so each
// output line fits within a target width.
package fmt

import (
	"bufio"
	"context"
	stdfmt "fmt"
	"io"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the fmt applet.
type Command struct{}

// New returns a fmt command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "fmt" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Simple optimal text formatter" }

// Run executes fmt.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... [FILE]...", stdio.Err)
	width := fs.IntP("width", "w", 75, "maximum line width")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	if *width <= 0 {
		_, _ = stdfmt.Fprintf(stdio.Err, "fmt: invalid width: %d\n", *width)
		return command.SilentFailure()
	}

	files := fs.Args()
	if len(files) == 0 {
		files = []string{"-"}
	}

	var firstErr error
	for _, name := range files {
		if err := c.fmtFile(stdio, name, *width); err != nil {
			_, _ = stdfmt.Fprintf(stdio.Err, "fmt: %s\n", command.FileError(name, err))
			if firstErr == nil {
				firstErr = command.SilentFailure()
			}
		}
	}
	return firstErr
}

// fmtFile reflows the text read from name. Blank lines separate paragraphs and
// are preserved; each paragraph's words are greedily packed into width columns.
func (c *Command) fmtFile(stdio command.IO, name string, width int) error {
	r, err := command.Open(stdio, name)
	if err != nil {
		return err
	}
	defer func() { _ = r.Close() }()

	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var para []string
	flush := func() error {
		if len(para) == 0 {
			return nil
		}
		text := reflow(strings.Join(para, " "), width)
		para = para[:0]
		_, werr := io.WriteString(stdio.Out, text)
		return werr
	}

	for sc.Scan() {
		line := sc.Text()
		if strings.TrimSpace(line) == "" {
			if err := flush(); err != nil {
				return err
			}
			if _, err := io.WriteString(stdio.Out, "\n"); err != nil {
				return err
			}
			continue
		}
		para = append(para, strings.Fields(line)...)
	}
	if err := sc.Err(); err != nil {
		return err
	}
	return flush()
}

// reflow greedily packs words into lines no wider than width, returning the
// joined lines each terminated by a newline.
func reflow(text string, width int) string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return ""
	}
	var b strings.Builder
	lineLen := 0
	for i, w := range words {
		switch {
		case i == 0:
			b.WriteString(w)
			lineLen = len(w)
		case lineLen+1+len(w) > width:
			b.WriteByte('\n')
			b.WriteString(w)
			lineLen = len(w)
		default:
			b.WriteByte(' ')
			b.WriteString(w)
			lineLen += 1 + len(w)
		}
	}
	b.WriteByte('\n')
	return b.String()
}
