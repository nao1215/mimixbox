// Package speaker implements the speaker applet: read text aloud using whatever
// text-to-speech engine is installed. It is a clean-room port of
// nao1215/speaker; because MimixBox builds with CGO disabled it shells out to an
// installed engine (spd-say, espeak or say) rather than linking an audio
// library, and degrades gracefully when none is present.
package speaker

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the speaker applet.
type Command struct{}

// New returns a speaker command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "speaker" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Read text aloud using an installed TTS engine" }

// engine describes a text-to-speech program and how to invoke it with text.
type engine struct {
	name string
	args func(text, lang string) []string
}

// engines is the ordered list of TTS programs speaker will try.
var engines = []engine{
	{"spd-say", func(text, lang string) []string {
		args := []string{"-w"}
		if lang != "" {
			args = append(args, "-l", lang)
		}
		return append(args, text)
	}},
	{"espeak", func(text, lang string) []string {
		args := []string{}
		if lang != "" {
			args = append(args, "-v", lang)
		}
		return append(args, text)
	}},
	{"say", func(text, lang string) []string { return []string{text} }},
}

// lookPath resolves a program to its absolute path; tests replace it.
var lookPath = exec.LookPath

// runEngine runs the chosen engine with its arguments; tests replace it.
var runEngine = func(name string, args []string) error {
	return exec.Command(name, args...).Run() //nolint:gosec // invoking a known TTS engine
}

// Run executes speaker.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... TEXT...", stdio.Err)
	lang := fs.StringP("language", "l", "", "language/voice to use (engine-specific)")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	text := strings.Join(fs.Args(), " ")
	if text == "" {
		return command.Failuref("no text to speak")
	}

	eng, ok := selectEngine()
	if !ok {
		return command.Failuref("no text-to-speech engine found (install spd-say, espeak, or say)")
	}

	if err := runEngine(eng.name, eng.args(text, *lang)); err != nil {
		return command.Failuref("%s failed: %v", eng.name, err)
	}
	_, _ = fmt.Fprintf(stdio.Out, "spoke %d characters via %s\n", len(text), eng.name)
	return nil
}

// selectEngine returns the first installed TTS engine.
func selectEngine() (engine, bool) {
	for _, e := range engines {
		if _, err := lookPath(e.name); err == nil {
			return e, true
		}
	}
	return engine{}, false
}
