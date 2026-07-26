package shell

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

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

	if len(newLine) == 0 {
		c.reset()
		bell()
		return nil, 0
	}
	if len(newLine) == 1 {
		c.reset()
		return newLine, length
	}

	sort.Slice(newLine, func(i, j int) bool {
		return string(newLine[i]) < string(newLine[j])
	})

	if !(c.belled && c.belledFor == string(line)) {
		c.belled = true
		c.belledFor = string(line)
		bell()
		return nil, 0
	}

	c.reset()
	return newLine, length
}

func (c *Completer) reset() {
	c.belled = false
	c.belledFor = ""
}

func bell() {
	fmt.Fprintf(os.Stderr, "%c", 0x07)
}
