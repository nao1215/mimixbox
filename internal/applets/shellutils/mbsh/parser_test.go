package mbsh

import (
	"strings"
	"testing"
)

// splitToks separates leading assignment tokens from the command arguments, the
// same way execInput does.
func splitToks(toks []token) (env, args []string) {
	i := 0
	for i < len(toks) && toks[i].assignKey != "" {
		env = append(env, toks[i].assignKey+"="+toks[i].assignVal)
		i++
	}
	for ; i < len(toks); i++ {
		args = append(args, toks[i].value)
	}
	return env, args
}

func TestTokenize(t *testing.T) {
	t.Setenv("FOO", "bar")
	t.Setenv("HOME", "/home/test")

	tests := []struct {
		name       string
		input      string
		lastStatus int
		wantEnv    []string
		wantArgs   []string
	}{
		{name: "plain words", input: "echo hello world", wantArgs: []string{"echo", "hello", "world"}},
		{name: "double quotes keep spaces", input: `echo "a b"`, wantArgs: []string{"echo", "a b"}},
		{name: "single quotes are literal", input: `echo 'a $FOO'`, wantArgs: []string{"echo", "a $FOO"}},
		{name: "dollar var expands", input: "echo $FOO", wantArgs: []string{"echo", "bar"}},
		{name: "braced var expands", input: "echo ${FOO}x", wantArgs: []string{"echo", "barx"}},
		{name: "var in double quotes expands", input: `echo "$FOO!"`, wantArgs: []string{"echo", "bar!"}},
		{name: "last status", input: "echo $?", lastStatus: 7, wantArgs: []string{"echo", "7"}},
		{name: "backslash escapes space", input: `echo a\ b`, wantArgs: []string{"echo", "a b"}},
		{name: "escaped dollar in double quotes", input: `echo "\$FOO"`, wantArgs: []string{"echo", "$FOO"}},
		{name: "tilde expands", input: "echo ~", wantArgs: []string{"echo", "/home/test"}},
		{name: "tilde slash expands", input: "echo ~/x", wantArgs: []string{"echo", "/home/test/x"}},
		{name: "undefined var is empty", input: "echo [${UNSET_VAR_XYZ}]", wantArgs: []string{"echo", "[]"}},
		{name: "single assignment prefix", input: "FOO=baz cmd", wantEnv: []string{"FOO=baz"}, wantArgs: []string{"cmd"}},
		{name: "multiple assignment prefixes", input: "A=1 B=2 cmd x", wantEnv: []string{"A=1", "B=2"}, wantArgs: []string{"cmd", "x"}},
		{name: "assignment value expands", input: "P=$FOO cmd", wantEnv: []string{"P=bar"}, wantArgs: []string{"cmd"}},
		{name: "equals after command is an argument", input: "echo A=1", wantArgs: []string{"echo", "A=1"}},
		{name: "only assignments", input: "FOO=bar", wantEnv: []string{"FOO=bar"}},
		{name: "empty input", input: "   "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			toks, err := tokenize(tt.input, tt.lastStatus)
			if err != nil {
				t.Fatalf("tokenize(%q) error = %v", tt.input, err)
			}
			env, args := splitToks(toks)
			if strings.Join(env, " ") != strings.Join(tt.wantEnv, " ") {
				t.Errorf("env = %v, want %v", env, tt.wantEnv)
			}
			if strings.Join(args, "\x00") != strings.Join(tt.wantArgs, "\x00") {
				t.Errorf("args = %q, want %q", args, tt.wantArgs)
			}
		})
	}
}

func TestTokenizeErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{name: "unterminated double quote", input: `echo "a b`},
		{name: "unterminated single quote", input: `echo 'a b`},
		{name: "missing closing brace", input: "echo ${FOO"},
		{name: "bad name in braces", input: "echo ${1bad}"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := tokenize(tt.input, 0); err == nil {
				t.Errorf("tokenize(%q) should have failed", tt.input)
			}
		})
	}
}
