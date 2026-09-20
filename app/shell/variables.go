package shell

import (
	"strings"
	"sync"
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

func (v *Variables) Parse(args []string) {
	v.mu.Lock()
	defer v.mu.Unlock()

	for _, arg := range args {
		key, value, found := strings.Cut(arg, "=")
		if !found {
			continue
		}

		if key != "" && value != "" {
			v.Set(key, value)
		}
	}
}
