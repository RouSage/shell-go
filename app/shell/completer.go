package shell

import (
	"fmt"
	"os"

	"github.com/chzyer/readline"
)

type Completer struct {
	readline.AutoCompleter
}

func NewCompleter(completer readline.AutoCompleter) *Completer {
	return &Completer{AutoCompleter: completer}
}

func (c *Completer) Do(line []rune, pos int) ([][]rune, int) {
	newLine, length := c.AutoCompleter.Do(line, pos)
	if len(newLine) == 0 {
		fmt.Fprintf(os.Stderr, "%c", 0x07)
	}

	return newLine, length
}
