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
				// Skip builtins so that this shell handles them separately from the system
				if !slices.Contains(builtins, info.Name()) {
					completers = append(completers, readline.PcItem(info.Name()))
				}
			}
		}
	}

	fileCompletions := readline.PcItemDynamic(listFiles("./"))
	for _, completer := range completers {
		completer.SetChildren([]readline.PrefixCompleterInterface{fileCompletions})
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

	// On subsequent <TAB>s, do a partial completion with the longest common prefix
	prefix := string(line[:pos])
	names := make([]string, len(newLine))
	for i, suffix := range newLine {
		names[i] = prefix + string(suffix)
	}

	commonPrefix := longestCommonPrefix(names)

	if len(commonPrefix) > len(prefix) {
		diff := len(commonPrefix) - len(prefix)
		return [][]rune{commonPrefix[len(commonPrefix)-diff:]}, 1
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

	// List matches in alphabetical order
	slices.Sort(names)
	fmt.Fprintf(os.Stdout, "\n%s\n$ %s", strings.Join(names, "  "), string(prefix))

	return nil, 0
}

func (c *Completer) reset() {
	c.belled = false
	c.belledFor = ""
}

func bell() {
	fmt.Fprintf(os.Stderr, "%c", 0x07)
}

func longestCommonPrefix(strs []string) []rune {
	commonPrefix := []rune(strs[0])
	for _, match := range strs {
		runes := []rune(match)

		for i := range commonPrefix {
			if len(runes) <= i || runes[i] != commonPrefix[i] {
				commonPrefix = commonPrefix[:i]
				break
			}
		}
	}

	return commonPrefix
}

func listFiles(path string) readline.DynamicCompleteFunc {
	return func(line string) []string {
		files := []string{}

		entries, err := os.ReadDir(path)
		if err != nil {
			return files
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				files = append(files, entry.Name())
			}
		}

		return files
	}
}
