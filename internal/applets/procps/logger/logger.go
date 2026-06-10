// Package logger implements the logger applet: write a message to the system
// log via syslog.
package logger

import (
	"context"
	"fmt"
	"io"
	"log/syslog"
	"os/user"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the logger applet.
type Command struct{}

// New returns a logger command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "logger" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Write a message to the system log" }

// logFunc is indirected so logging is testable without a running syslogd.
var logFunc = func(p syslog.Priority, tag, msg string) error {
	w, err := syslog.New(p, tag)
	if err != nil {
		return err
	}
	defer func() { _ = w.Close() }()
	_, err = io.WriteString(w, msg)
	return err
}

var facilities = map[string]syslog.Priority{
	"kern": syslog.LOG_KERN, "user": syslog.LOG_USER, "mail": syslog.LOG_MAIL,
	"daemon": syslog.LOG_DAEMON, "auth": syslog.LOG_AUTH, "syslog": syslog.LOG_SYSLOG,
	"lpr": syslog.LOG_LPR, "news": syslog.LOG_NEWS, "uucp": syslog.LOG_UUCP,
	"cron": syslog.LOG_CRON, "authpriv": syslog.LOG_AUTHPRIV, "ftp": syslog.LOG_FTP,
	"local0": syslog.LOG_LOCAL0, "local1": syslog.LOG_LOCAL1, "local2": syslog.LOG_LOCAL2,
	"local3": syslog.LOG_LOCAL3, "local4": syslog.LOG_LOCAL4, "local5": syslog.LOG_LOCAL5,
	"local6": syslog.LOG_LOCAL6, "local7": syslog.LOG_LOCAL7,
}

var levels = map[string]syslog.Priority{
	"emerg": syslog.LOG_EMERG, "alert": syslog.LOG_ALERT, "crit": syslog.LOG_CRIT,
	"err": syslog.LOG_ERR, "error": syslog.LOG_ERR, "warning": syslog.LOG_WARNING,
	"warn": syslog.LOG_WARNING, "notice": syslog.LOG_NOTICE, "info": syslog.LOG_INFO,
	"debug": syslog.LOG_DEBUG,
}

// Run executes logger.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-p PRIORITY] [-t TAG] [-s] MESSAGE...", stdio.Err).WithHelp(command.Help{
		Description: "Write MESSAGE (or standard input) to the system log. -p sets the priority as " +
			"facility.level (default user.notice), -t sets the tag (default the user name), and -s " +
			"also echoes the message to standard error.",
		Examples: []command.Example{
			{Command: "logger -t myapp 'started'", Explain: "Log a tagged message."},
			{Command: "logger -p auth.warning 'bad login'", Explain: "Log with a priority."},
		},
		ExitStatus: "0  the message was logged.\n1  the priority was invalid or syslog was unreachable.",
	})
	prio := fs.StringP("priority", "p", "user.notice", "facility.level for the message")
	tag := fs.StringP("tag", "t", "", "mark the message with this tag")
	toStderr := fs.BoolP("stderr", "s", false, "also write the message to standard error")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	priority, err := parsePriority(*prio)
	if err != nil {
		return command.Failuref("%v", err)
	}

	var message string
	if rest := fs.Args(); len(rest) > 0 {
		message = strings.Join(rest, " ")
	} else {
		data, _ := io.ReadAll(stdio.In)
		message = strings.TrimRight(string(data), "\n")
	}

	t := *tag
	if t == "" {
		t = defaultTag()
	}

	if err := logFunc(priority, t, message); err != nil {
		return command.Failuref("%v", err)
	}
	if *toStderr {
		_, _ = fmt.Fprintf(stdio.Err, "%s: %s\n", t, message)
	}
	return nil
}

// parsePriority converts a "facility.level" spec to a syslog priority.
func parsePriority(spec string) (syslog.Priority, error) {
	facName, levName, hasDot := strings.Cut(spec, ".")
	if !hasDot {
		facName, levName = "user", spec
	}
	fac, ok := facilities[strings.ToLower(facName)]
	if !ok {
		return 0, fmt.Errorf("unknown facility: %q", facName)
	}
	lev, ok := levels[strings.ToLower(levName)]
	if !ok {
		return 0, fmt.Errorf("unknown level: %q", levName)
	}
	return fac | lev, nil
}

func defaultTag() string {
	if u, err := user.Current(); err == nil {
		return u.Username
	}
	return "logger"
}
