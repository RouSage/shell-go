package shell

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/chzyer/readline"
)

type Completer struct {
	readline.AutoCompleter
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
		fmt.Fprintf(os.Stderr, "%c", 0x07)
	}

	return newLine, length
}
