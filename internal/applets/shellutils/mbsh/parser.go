package mbsh

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// tokKind distinguishes a word from a shell operator.
type tokKind int

const (
	tokWord tokKind = iota
	tokOp           // ; | < > >>
)

// token is one word or operator produced by the tokenizer. For a word it also
// carries the split name/value when the word is an unquoted NAME=value
// assignment.
type token struct {
	kind      tokKind
	value     string // expanded word, or the operator string
	assignKey string // non-empty when the word is an unquoted NAME=... assignment
	assignVal string // expanded value part of an assignment
}

// isOperatorByte reports whether c begins an unquoted shell operator.
func isOperatorByte(c byte) bool {
	return c == ';' || c == '|' || c == '<' || c == '>'
}

// tokenize splits a command line into words the way a POSIX-ish shell does:
// honoring single quotes (literal), double quotes (with $ expansion),
// backslash escapes, $VAR / ${VAR} / $? expansion, and leading ~ expansion.
// lastStatus supplies $?. Expansion does not re-split on whitespace, so an
// expanded value stays a single word.
func tokenize(input string, lastStatus int) ([]token, error) {
	var toks []token
	i, n := 0, len(input)
	for i < n {
		for i < n && (input[i] == ' ' || input[i] == '\t') {
			i++
		}
		if i >= n {
			break
		}
		if isOperatorByte(input[i]) {
			op, next := parseOperator(input, i)
			toks = append(toks, token{kind: tokOp, value: op})
			i = next
			continue
		}
		tok, next, err := parseWord(input, i, lastStatus)
		if err != nil {
			return nil, err
		}
		toks = append(toks, tok)
		i = next
	}
	return toks, nil
}

// parseOperator reads the operator at input[i] (">>" or one of ; | < >) and
// returns it with the index just past it.
func parseOperator(input string, i int) (string, int) {
	if input[i] == '>' && i+1 < len(input) && input[i+1] == '>' {
		return ">>", i + 2
	}
	return string(input[i]), i + 1
}

// parseWord scans one word starting at start and returns it with the index just
// past it.
func parseWord(input string, start, lastStatus int) (token, int, error) {
	var sb strings.Builder
	var nameBuf strings.Builder
	nameScan := true   // still scanning a possible leading assignment NAME
	tildeOK := true     // a leading '~' is eligible for home expansion
	assignKey := ""
	assignValStart := -1

	i, n := start, len(input)
	for i < n {
		c := input[i]
		if c == ' ' || c == '\t' || isOperatorByte(c) {
			break
		}

		switch c {
		case '\'':
			nameScan, tildeOK = false, false
			i++
			for i < n && input[i] != '\'' {
				sb.WriteByte(input[i])
				i++
			}
			if i >= n {
				return token{}, 0, errors.New("unterminated single quote")
			}
			i++ // closing '
		case '"':
			nameScan, tildeOK = false, false
			i++
			for i < n && input[i] != '"' {
				switch {
				case input[i] == '\\' && i+1 < n && isDquoteEscape(input[i+1]):
					sb.WriteByte(input[i+1])
					i += 2
				case input[i] == '$':
					val, ni, err := expandVar(input, i, lastStatus)
					if err != nil {
						return token{}, 0, err
					}
					sb.WriteString(val)
					i = ni
				default:
					sb.WriteByte(input[i])
					i++
				}
			}
			if i >= n {
				return token{}, 0, errors.New("unterminated double quote")
			}
			i++ // closing "
		case '\\':
			nameScan, tildeOK = false, false
			if i+1 < n {
				sb.WriteByte(input[i+1])
				i += 2
			} else {
				i++ // a trailing backslash is dropped
			}
		case '$':
			nameScan, tildeOK = false, false
			val, ni, err := expandVar(input, i, lastStatus)
			if err != nil {
				return token{}, 0, err
			}
			sb.WriteString(val)
			i = ni
		case '~':
			if tildeOK && (i+1 >= n || input[i+1] == '/' || input[i+1] == ' ' || input[i+1] == '\t') {
				sb.WriteString(os.Getenv("HOME"))
			} else {
				sb.WriteByte('~')
			}
			nameScan, tildeOK = false, false
			i++
		case '=':
			if nameScan && nameBuf.Len() > 0 {
				assignKey = nameBuf.String()
			}
			sb.WriteByte('=')
			i++
			nameScan, tildeOK = false, false
			if assignKey != "" {
				assignValStart = sb.Len()
			}
		default:
			if nameScan {
				if isNameChar(c, nameBuf.Len() == 0) {
					nameBuf.WriteByte(c)
				} else {
					nameScan = false
				}
			}
			sb.WriteByte(c)
			tildeOK = false
			i++
		}
	}

	value := sb.String()
	tok := token{value: value}
	if assignKey != "" && assignValStart >= 0 {
		tok.assignKey = assignKey
		tok.assignVal = value[assignValStart:]
	}
	return tok, i, nil
}

// expandVar expands the variable reference beginning at input[i] (which is '$')
// and returns its value and the index just past the reference.
func expandVar(input string, i, lastStatus int) (string, int, error) {
	i++ // skip '$'
	n := len(input)
	if i >= n {
		return "$", i, nil // a lone trailing '$' is literal
	}
	switch {
	case input[i] == '?':
		return strconv.Itoa(lastStatus), i + 1, nil
	case input[i] == '{':
		i++ // skip '{'
		start := i
		for i < n && input[i] != '}' {
			i++
		}
		if i >= n {
			return "", 0, errors.New("bad substitution: missing '}'")
		}
		name := input[start:i]
		i++ // skip '}'
		if name == "?" {
			return strconv.Itoa(lastStatus), i, nil
		}
		if !isValidName(name) {
			return "", 0, fmt.Errorf("bad substitution: ${%s}", name)
		}
		return os.Getenv(name), i, nil
	case isNameStart(input[i]):
		start := i
		for i < n && isNameContinue(input[i]) {
			i++
		}
		return os.Getenv(input[start:i]), i, nil
	default:
		return "$", i, nil // '$' followed by a non-name char is literal
	}
}

func isDquoteEscape(c byte) bool {
	// Inside double quotes a backslash is literal except before these.
	return c == '$' || c == '`' || c == '"' || c == '\\'
}

func isNameStart(c byte) bool {
	return c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func isNameContinue(c byte) bool {
	return isNameStart(c) || (c >= '0' && c <= '9')
}

// isNameChar reports whether c can be part of an assignment name; first is true
// for the first character (which may not be a digit).
func isNameChar(c byte, first bool) bool {
	if first {
		return isNameStart(c)
	}
	return isNameContinue(c)
}

func isValidName(s string) bool {
	if s == "" || !isNameStart(s[0]) {
		return false
	}
	for i := 1; i < len(s); i++ {
		if !isNameContinue(s[i]) {
			return false
		}
	}
	return true
}
