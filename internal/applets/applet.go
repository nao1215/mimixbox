// mimixbox/internal/applets/applet.go
//
// # Copyright 2021 Naohiro CHIKAMATSU
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package applets

//go:generate go run github.com/nao1215/mimixbox/cmd/genapplets

import (
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

type Applet struct {
	// Cmd is the applet itself. The top-level dispatcher runs it through
	// internal/command.Execute with an injected command.IO, so an applet can be
	// dispatched entirely in memory without mutating os.Args or touching the
	// process streams.
	Cmd  command.Command
	Desc string
}

// Applets is the applet table, populated by the generated init in
// applet_registry_gen.go (run `make generate` to refresh it).
var Applets map[string]Applet

// reg builds an Applet entry for a command. The command's own Synopsis becomes
// the listed description, so the two never drift apart.
func reg(c command.Command) Applet {
	return Applet{Cmd: c, Desc: c.Synopsis()}
}

// register adds c to the applet table under its own Name(). The generated init
// calls this once per applet constructor, so a key can never drift from the
// command it dispatches to.
func register(c command.Command) {
	Applets[c.Name()] = reg(c)
}

func HasApplet(target string) bool {
	_, ok := Applets[target]
	return ok
}

// ListAppletsTo writes the "name - description" table to w.
func ListAppletsTo(w io.Writer) {
	format := "%" + strconv.Itoa(longestAppletLength()) + "s - %s\n"
	for _, key := range SortApplet() {
		_, _ = fmt.Fprintf(w, format, key, Applets[key].Desc)
	}
}

// ShowAppletsBySpaceSeparatedTo writes the space-separated, wrapped applet
// names to w.
func ShowAppletsBySpaceSeparatedTo(w io.Writer) {
	const wrap = 60
	var b strings.Builder
	lineLen := 0
	for i, key := range SortApplet() {
		if i > 0 {
			if lineLen+1+len(key) > wrap {
				b.WriteByte('\n')
				lineLen = 0
			} else {
				b.WriteByte(' ')
				lineLen++
			}
		}
		b.WriteString(key)
		lineLen += len(key)
	}
	b.WriteByte('\n')
	_, _ = fmt.Fprint(w, b.String())
}

func SortApplet() []string {
	var keys []string
	for applet := range Applets {
		keys = append(keys, applet)
	}
	sort.Strings(keys)
	return keys
}

func longestAppletLength() int {
	max := 0
	for _, key := range SortApplet() {
		if max < len(key) {
			max = len(key)
		}
	}
	return max
}
