package shell

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/chzyer/readline"
)

type Completer struct {
	readline.AutoCompleter
	belledFor string
	belled    bool
}

func NewCompleter(completers ...readline.PrefixCompleterInterface) *Completer {
	paths := filepath.SplitList(os.Getenv("PATH"))
	for _, path := range paths {
		entries, err := os.ReadDir(path)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			info, err := entry.Info()
			if err != nil {
				continue
			}
			if !info.IsDir() && info.Mode().Perm()&0111 != 0 {
				completers = append(completers, readline.PcItem(info.Name()))
			}
		}
	}

	completer := readline.NewPrefixCompleter(completers...)

	return &Completer{AutoCompleter: completer}
}

func (c *Completer) Do(line []rune, pos int) ([][]rune, int) {
	newLine, length := c.AutoCompleter.Do(line, pos)

	// Ring the bell on no matches
	if len(newLine) == 0 {
		c.reset()
		bell()
		return nil, 0
	}
	// One match, let readline handle it (autocomplete)
	if len(newLine) == 1 {
		c.reset()
		return newLine, length
	}

	// Multiple matches: bell on first <TAB>
	// list matches on the second <TAB>
	if !(c.belled && c.belledFor == string(line)) {
		c.belled = true
		c.belledFor = string(line)
		bell()
		return nil, 0
	}

	c.reset()

	prefix := string(line[:pos])
	names := make([]string, len(newLine))
	for i, suffix := range newLine {
		names[i] = prefix + string(suffix)
	}
	slices.Sort(names)

	fmt.Fprintf(os.Stdout, "\n%s\n$ %s", strings.Join(names, "  "), prefix)

	return nil, 0
}

func (c *Completer) reset() {
	c.belled = false
	c.belledFor = ""
}

func bell() {
	fmt.Fprintf(os.Stderr, "%c", 0x07)
}
