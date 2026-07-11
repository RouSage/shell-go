package shell

import "strings"

func parseArgs(command string) []string {
	var args []string
	var current strings.Builder
	hasToken := false
	inSingleQuote := false
	inDoubleQuote := false
	isEscaped := false

	for _, r := range command {
		switch {
		case inSingleQuote:
			if r == '\'' {
				inSingleQuote = false
			} else {
				current.WriteRune(r)
			}
		case inDoubleQuote:
			if r == '"' {
				inDoubleQuote = false
			} else {
				current.WriteRune(r)
			}
		case isEscaped:
			current.WriteRune(r)
			isEscaped = false
		case r == '\\':
			isEscaped = true
		case r == '\'':
			if inDoubleQuote {
				current.WriteRune(r)
			} else {
				inSingleQuote = true
				hasToken = true
			}
		case r == '"':
			inDoubleQuote = true
			hasToken = true
		case r == ' ' || r == '\t':
			if hasToken {
				args = append(args, current.String())
				current.Reset()
				hasToken = false
			}
		default:
			current.WriteRune(r)
			hasToken = true
		}
	}

	if hasToken {
		args = append(args, current.String())
	}

	return args
}
