// mimixbox/internal/lib/io_inject_test.go
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
package mb

import (
	"bytes"
	"os/user"
	"strings"
	"testing"
)

func TestQuestionFrom(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"y is yes", "y\n", true},
		{"yes is yes", "yes\n", true},
		{"uppercase YES is yes", "YES\n", true},
		{"n is no", "n\n", false},
		{"no is no", "no\n", false},
		{"reprompt on blank then yes", "\ny\n", true},
		{"reprompt on invalid then no", "maybe\nn\n", false},
		{"eof returns false", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var out bytes.Buffer
			got := QuestionFrom(strings.NewReader(tt.input), &out, "Continue?")
			if got != tt.want {
				t.Errorf("QuestionFrom(%q) = %v, want %v", tt.input, got, tt.want)
			}
			if !strings.Contains(out.String(), "Continue? [Y/n]") {
				t.Errorf("prompt = %q, want it to contain the question", out.String())
			}
		})
	}
}

func TestShowVersionTo(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	ShowVersionTo(&out, "mimixbox", "1.2.3")
	want := "mimixbox version 1.2.3 (under Apache License version 2.0)\n"
	if got := out.String(); got != want {
		t.Errorf("ShowVersionTo() = %q, want %q", got, want)
	}
}

func TestPrintStrWithNumberLineTo(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	PrintStrWithNumberLineTo(&out, 7, "%6d  %s", "hello")
	want := "     7  hello"
	if got := out.String(); got != want {
		t.Errorf("PrintStrWithNumberLineTo() = %q, want %q", got, want)
	}
}

func TestPrintStrListWithNumberLineTo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		lines      []string
		countEmpty bool
		want       string
	}{
		{
			name:       "blank line keeps its own line when not counted",
			lines:      []string{"a\n", "\n", "b\n"},
			countEmpty: false,
			want:       "     1  a\n\n     2  b\n",
		},
		{
			name:       "blank line is numbered when counted",
			lines:      []string{"a\n", "\n", "b\n"},
			countEmpty: true,
			want:       "     1  a\n     2  \n     3  b\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var out bytes.Buffer
			PrintStrListWithNumberLineTo(&out, tt.lines, tt.countEmpty)
			if got := out.String(); got != tt.want {
				t.Errorf("PrintStrListWithNumberLineTo() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDumpTo(t *testing.T) {
	t.Parallel()

	var plain bytes.Buffer
	DumpTo(&plain, []string{"x\n", "y\n"}, false)
	if got, want := plain.String(), "x\ny\n"; got != want {
		t.Errorf("DumpTo(withNumber=false) = %q, want %q", got, want)
	}

	var numbered bytes.Buffer
	DumpTo(&numbered, []string{"x\n"}, true)
	if got, want := numbered.String(), "     1  x\n"; got != want {
		t.Errorf("DumpTo(withNumber=true) = %q, want %q", got, want)
	}
}

func TestDumpGroupsTo(t *testing.T) {
	t.Parallel()

	groups := []user.Group{
		{Name: "wheel", Gid: "10"},
		{Name: "staff", Gid: "20"},
	}

	var byName bytes.Buffer
	DumpGroupsTo(&byName, groups, true)
	if got, want := byName.String(), "wheel staff\n"; got != want {
		t.Errorf("DumpGroupsTo(showName=true) = %q, want %q", got, want)
	}

	var byGid bytes.Buffer
	DumpGroupsTo(&byGid, groups, false)
	if got, want := byGid.String(), "10 20\n"; got != want {
		t.Errorf("DumpGroupsTo(showName=false) = %q, want %q", got, want)
	}
}

func TestParrotFrom(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	ParrotFrom(strings.NewReader("alpha\nbeta\n"), &out, false)
	if got, want := out.String(), "alpha\nbeta\n"; got != want {
		t.Errorf("ParrotFrom(withNl=false) = %q, want %q", got, want)
	}
}
