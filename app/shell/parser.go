package shell

import (
	"slices"
	"strings"
)

type Token struct {
	value  string
	quoted bool
	op     bool
}

func newToken(value string, quoted bool) Token {
	return Token{
		value:  value,
		quoted: quoted,
		op:     !quoted && (value == "&" || slices.Contains(redirectOps, value)),
	}
}

func parseSegments(line string) [][]Token {
	var args []Token
	var current strings.Builder
	var cmds = make([][]Token, 0)

	var (
		hasToken      = false
		literal       = false
		inSingleQuote = false
		inDoubleQuote = false
		isEscaped     = false
	)

	flush := func() {
		if !hasToken {
			return
		}

		args = append(args, newToken(current.String(), literal))
		current.Reset()
		hasToken, literal = false, false
	}

	for _, r := range line {
		switch {
		case inSingleQuote:
			if r == '\'' {
				inSingleQuote = false
			} else {
				current.WriteRune(r)
			}
		case inDoubleQuote:
			if isEscaped {
				if r == '"' || r == '\\' || r == '$' || r == '`' || r == '\n' {
					current.WriteRune(r)
				} else {
					current.WriteRune('\\')
					current.WriteRune(r)
				}
				isEscaped = false
				hasToken, literal = true, true
			} else if r == '\\' {
				isEscaped = true
			} else if r == '"' {
				inDoubleQuote = false
			} else {
				current.WriteRune(r)
			}
		case isEscaped:
			current.WriteRune(r)
			isEscaped = false
			hasToken, literal = true, true
		case r == '\\':
			isEscaped = true
		case r == '\'':
			if inDoubleQuote {
				current.WriteRune(r)
			} else {
				inSingleQuote = true
				hasToken, literal = true, true
			}
		case r == '"':
			inDoubleQuote = true
			hasToken, literal = true, true
		case r == '|':
			flush()

			if len(args) > 0 {
				cmds = append(cmds, args)
			}
			args = nil
		case r == ' ' || r == '\t':
			flush()
		default:
			current.WriteRune(r)
			hasToken = true
		}
	}

	flush()

	if len(args) > 0 {
		cmds = append(cmds, args)
	}

	return cmds
}
