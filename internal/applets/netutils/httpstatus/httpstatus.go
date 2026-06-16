// Package httpstatus implements the http-status-code applet: explain HTTP
// status codes (meaning and RFC reference) from the command line. It is a
// clean-room port of the maintainer's archived nao1215/http-status-code.
package httpstatus

import (
	"context"
	"fmt"
	"sort"
	"strconv"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the http-status-code applet.
type Command struct{}

// New returns an http-status-code command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "http-status-code" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Explain HTTP status codes and their RFC references" }

// status describes one HTTP status code.
type status struct {
	meaning string
	ref     string
}

// Run executes http-status-code.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "(search CODE... | list)", stdio.Err).WithHelp(command.Help{
		Description: "Explain HTTP status codes and their defining references. The list " +
			"subcommand prints every known code, while search CODE... prints the meaning " +
			"and reference for each requested code.",
		Examples: []command.Example{
			{Command: "http-status-code list", Explain: "List every known status code."},
			{Command: "http-status-code search 404", Explain: "Explain the 404 status code."},
			{Command: "http-status-code search 200 301 500", Explain: "Explain several codes at once."},
		},
		ExitStatus: "0  every requested code was explained.\n1  a code was invalid, unknown, or no subcommand was given.",
	})
	fs.SetInterspersed(false)

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	rest := fs.Args()
	if len(rest) == 0 {
		return command.Failuref("expected a subcommand: search or list")
	}

	switch rest[0] {
	case "list":
		return c.list(stdio)
	case "search":
		return c.search(stdio, rest[1:])
	default:
		return command.Failuref("unknown subcommand %q", rest[0])
	}
}

// list prints every known status code in ascending order.
func (c *Command) list(stdio command.IO) error {
	codes := make([]int, 0, len(table))
	for code := range table {
		codes = append(codes, code)
	}
	sort.Ints(codes)
	for _, code := range codes {
		if _, err := fmt.Fprintln(stdio.Out, format(code, table[code])); err != nil {
			return command.Failure(err)
		}
	}
	return nil
}

// search prints the explanation for each requested code.
func (c *Command) search(stdio command.IO, codes []string) error {
	if len(codes) == 0 {
		return command.Failuref("search requires at least one CODE")
	}
	var firstErr error
	for _, raw := range codes {
		code, err := strconv.Atoi(raw)
		if err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "http-status-code: invalid code %q\n", raw)
			if firstErr == nil {
				firstErr = command.SilentFailure()
			}
			continue
		}
		s, ok := table[code]
		if !ok {
			_, _ = fmt.Fprintf(stdio.Err, "http-status-code: unknown status code %d\n", code)
			if firstErr == nil {
				firstErr = command.SilentFailure()
			}
			continue
		}
		if _, err := fmt.Fprintln(stdio.Out, format(code, s)); err != nil {
			return command.Failure(err)
		}
	}
	return firstErr
}

// format renders one status line: "CODE Meaning (ref.=RFC..., ...)".
func format(code int, s status) string {
	return fmt.Sprintf("%d %s (ref.=%s)", code, s.meaning, s.ref)
}

// table maps each status code to its meaning and defining reference.
var table = map[int]status{
	100: {"Continue", "RFC9110, Section 15.2.1"},
	101: {"Switching Protocols", "RFC9110, Section 15.2.2"},
	103: {"Early Hints", "RFC8297"},
	200: {"OK", "RFC9110, Section 15.3.1"},
	201: {"Created", "RFC9110, Section 15.3.2"},
	202: {"Accepted", "RFC9110, Section 15.3.3"},
	203: {"Non-Authoritative Information", "RFC9110, Section 15.3.4"},
	204: {"No Content", "RFC9110, Section 15.3.5"},
	205: {"Reset Content", "RFC9110, Section 15.3.6"},
	206: {"Partial Content", "RFC9110, Section 15.3.7"},
	300: {"Multiple Choices", "RFC9110, Section 15.4.1"},
	301: {"Moved Permanently", "RFC9110, Section 15.4.2"},
	302: {"Found", "RFC9110, Section 15.4.3"},
	303: {"See Other", "RFC9110, Section 15.4.4"},
	304: {"Not Modified", "RFC9110, Section 15.4.5"},
	307: {"Temporary Redirect", "RFC9110, Section 15.4.8"},
	308: {"Permanent Redirect", "RFC9110, Section 15.4.9"},
	400: {"Bad Request", "RFC9110, Section 15.5.1"},
	401: {"Unauthorized", "RFC9110, Section 15.5.2"},
	402: {"Payment Required", "RFC9110, Section 15.5.3"},
	403: {"Forbidden", "RFC9110, Section 15.5.4"},
	404: {"Not Found", "RFC9110, Section 15.5.5"},
	405: {"Method Not Allowed", "RFC9110, Section 15.5.6"},
	406: {"Not Acceptable", "RFC9110, Section 15.5.7"},
	407: {"Proxy Authentication Required", "RFC9110, Section 15.5.8"},
	408: {"Request Timeout", "RFC9110, Section 15.5.9"},
	409: {"Conflict", "RFC9110, Section 15.5.10"},
	410: {"Gone", "RFC9110, Section 15.5.11"},
	411: {"Length Required", "RFC9110, Section 15.5.12"},
	412: {"Precondition Failed", "RFC9110, Section 15.5.13"},
	413: {"Content Too Large", "RFC9110, Section 15.5.14"},
	414: {"URI Too Long", "RFC9110, Section 15.5.15"},
	415: {"Unsupported Media Type", "RFC9110, Section 15.5.16"},
	416: {"Range Not Satisfiable", "RFC9110, Section 15.5.17"},
	417: {"Expectation Failed", "RFC9110, Section 15.5.18"},
	418: {"I'm a teapot", "RFC9110, Section 15.5.19"},
	421: {"Misdirected Request", "RFC9110, Section 15.5.20"},
	422: {"Unprocessable Content", "RFC9110, Section 15.5.21"},
	426: {"Upgrade Required", "RFC9110, Section 15.5.22"},
	428: {"Precondition Required", "RFC6585"},
	429: {"Too Many Requests", "RFC6585"},
	431: {"Request Header Fields Too Large", "RFC6585"},
	451: {"Unavailable For Legal Reasons", "RFC7725"},
	500: {"Internal Server Error", "RFC9110, Section 15.6.1"},
	501: {"Not Implemented", "RFC9110, Section 15.6.2"},
	502: {"Bad Gateway", "RFC9110, Section 15.6.3"},
	503: {"Service Unavailable", "RFC9110, Section 15.6.4"},
	504: {"Gateway Timeout", "RFC9110, Section 15.6.5"},
	505: {"HTTP Version Not Supported", "RFC9110, Section 15.6.6"},
	511: {"Network Authentication Required", "RFC6585"},
}
