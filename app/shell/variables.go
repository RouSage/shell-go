package shell

import (
	"strings"
	"sync"
	"unicode"
)

type Variables struct {
	values map[string]string
	mu     sync.Mutex
}

func NewVariables() *Variables {
	return &Variables{
		values: make(map[string]string),
	}
}

func (v *Variables) Set(key, value string) {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.values[key] = value
}

func (v *Variables) Get(key string) (string, bool) {
	value, ok := v.values[key]
	if !ok {
		return "", ok
	}

	return value, ok
}

func (v *Variables) Parse(args []string) (string, bool) {
	for _, arg := range args {
		identifier, value, found := strings.Cut(arg, "=")
		if !found {
			continue
		}

		if identifier != "" && value != "" {
			if isValid := validateIdentifier(identifier); !isValid {
				return arg, false
			}

			v.Set(identifier, value)
		}
	}

	return "", true
}

func validateIdentifier(name string) bool {
	for i, ch := range name {
		// Must start with a letter or an underscore
		if i == 0 {
			if ch != '_' && !unicode.IsLetter(ch) {
				return false
			}

			continue
		}

		if ch != '_' && !unicode.IsLetter(ch) && !unicode.IsDigit(ch) {
			return false
		}
	}

	return true
}
