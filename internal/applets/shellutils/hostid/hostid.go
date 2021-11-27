//
// mimixbox/internal/applets/shellutils/hostid/hostid.go
//
// Copyright 2021 Naohiro CHIKAMATSU
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package hostid

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/jessevdk/go-flags"
	mb "github.com/nao1215/mimixbox/internal/lib"
)

const cmdName string = "hostid"
const version = "0.9.1"

var osExit = os.Exit

type options struct {
	Version bool `short:"v" long:"version" description:"Show hostid command version"`
}

// Exit code
const (
	ExitSuccess int = iota // 0
	ExitFailuer
)

func Run() (int, error) {
	var opts options
	var err error

	if _, err = parseArgs(&opts); err != nil {
		return ExitSuccess, nil
	}
	return hostid()
}

func hostid() (int, error) {
	ip4, err := mb.Ip4()
	if err != nil {
		return ExitFailuer, err
	}

	//TODO: The output doesn't match the Coreutils version of hostid command.
	// First, the IP address should be calculated from the hostname.
	// Next, the process of converting the IP address to hexadecimal does not match.
	// I don't know the cause, so I'll deal with it later.
	for _, ip := range ip4 {
		ipList := strings.Split(ip, ".")
		fmt.Fprintf(os.Stdout, "%02x%02x%02x%02x\n",
			atoi(ipList[1]), atoi(ipList[0]), atoi(ipList[3]), atoi(ipList[2]))
	}

	return ExitSuccess, nil
}

func atoi(decimal string) int {
	i, _ := strconv.Atoi(decimal)
	return i
}

func parseArgs(opts *options) ([]string, error) {
	p := initParser(opts)

	args, err := p.Parse()
	if err != nil {
		return nil, err
	}

	if opts.Version {
		mb.ShowVersion(cmdName, version)
		osExit(ExitSuccess)
	}

	return args, nil
}

func initParser(opts *options) *flags.Parser {
	parser := flags.NewParser(opts, flags.Default)
	parser.Name = cmdName
	parser.Usage = "[OPTIONS]"

	return parser
}
