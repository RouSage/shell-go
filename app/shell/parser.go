package shell

import "strings"

func parseSegments(line string) [][]string {
	var args []string
	var current strings.Builder
	var cmds = make([][]string, 0)

	var (
		hasToken      = false
		inSingleQuote = false
		inDoubleQuote = false
		isEscaped     = false
	)

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
		case r == '|':
			if hasToken {
				args = append(args, current.String())
				current.Reset()
				hasToken = false
			}

			if len(args) > 0 {
				cmds = append(cmds, args)
			}
			args = nil
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
	if len(args) > 0 {
		cmds = append(cmds, args)
	}

	return cmds
}
