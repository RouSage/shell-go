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
	pci       readline.PrefixCompleterInterface
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

	completer := readline.NewPrefixCompleter(completers...)

	return &Completer{pci: completer}
}

func (c *Completer) Do(line []rune, pos int) ([][]rune, int) {
	full := string(line[:pos])

	// if there's a space, we complete the argument
	if spaceIdx := strings.LastIndex(full, " "); spaceIdx >= 0 {
		arg := full[spaceIdx+1:]
		dir := "./"
		argPrefix := ""
		filePrefix := arg

		// if there's a slash, we need to match files in that directory
		if slashIdx := strings.LastIndex(arg, "/"); slashIdx >= 0 {
			dir = arg[:slashIdx+1]
			argPrefix = dir
			filePrefix = arg[slashIdx+1:]
		}

		fileCompletions := readline.PcItemDynamic(listFilesAndDirs(dir, argPrefix, filePrefix))
		for _, completer := range c.pci.GetChildren() {
			completer.SetChildren([]readline.PrefixCompleterInterface{fileCompletions})
		}
	}

	newLine, length := c.pci.Do(line, pos)
	for i, line := range newLine {
		// if the line ends with "/ ", strip the trailing space
		// so that the it's possibl to continue with the nested dir completions
		if s := string(line); strings.HasSuffix(s, "/ ") {
			newLine[i] = []rune(s[:len(s)-1])
		}
	}

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

	// On subsequent <TAB>s, do a partial completion with the longest common full
	names := make([]string, len(newLine))
	for i, suffix := range newLine {
		names[i] = full + string(suffix)
	}

	commonPrefix := longestCommonPrefix(names)

	if len(commonPrefix) > len(full) {
		diff := len(commonPrefix) - len(full)
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
	fmt.Fprintf(os.Stdout, "\n%s\n$ %s", strings.Join(names, "  "), string(full))

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

// listFilesAndDirs matches files in path, and spells each match with argPrefix, the
// directory exactly as the user typed it. Completions are matched against the
// whole argument, so, e.g., "app/main.go" is what has to come back for "app/ma".
func listFilesAndDirs(path, argPrefix, prefix string) readline.DynamicCompleteFunc {
	return func(_ string) []string {
		results := []string{}

		entries, err := os.ReadDir(path)
		if err != nil {
			return results
		}

		for _, entry := range entries {
			name := entry.Name()

			if entry.IsDir() {
				results = append(results, argPrefix+name+"/")
				continue
			}

			if strings.HasPrefix(name, prefix) {
				results = append(results, argPrefix+name)
			}
		}

		return results
	}
}
