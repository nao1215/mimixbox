package printf

import (
	"fmt"
	"os"
	"strconv"

	"github.com/jessevdk/go-flags"
	mb "github.com/nao1215/mimixbox/internal/lib"
)

const cmdName string = "printf"
const version = "1.0.0"

var osExit = os.Exit

type options struct {
	Version bool `short:"v" long:"version" description:"Show printf command version"`
}

func Run() (int, error) {
	var output []any
	var opts options
	var args []string
	var err error
	var outputstr string
	if args, err = parseArgs(&opts); err != nil {
		return mb.ExitFailure, nil
	}

	if len(args) == 0 {
		output = make([]any, 0)
	} else {
		for i := 1; i < len(args); i++ {
			num, err := strconv.Atoi(args[i])
			if err == nil {
				output = append(output, num)
			} else {
				output = append(output, args[i])
			}
		}
	}
	outputstr = fmt.Sprintf(args[0], output...)
	fmt.Fprintln(os.Stdout, outputstr)
	return mb.ExitSuccess, nil
}

func parseArgs(opts *options) ([]string, error) {
	p := initParser(opts)

	args, err := p.Parse()
	if err != nil {
		return nil, err
	}

	if opts.Version {
		mb.ShowVersion(cmdName, version)
		osExit(mb.ExitSuccess)
	}

	return args, nil
}

func initParser(opts *options) *flags.Parser {
	parser := flags.NewParser(opts, flags.Default)
	parser.Name = cmdName
	parser.Usage = "[OPTIONS] STRING"

	return parser
}
