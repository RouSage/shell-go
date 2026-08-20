package shell

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type Completer struct {
	commands  []string
	belledFor string
	belled    bool
}

// NewCompleter collects the names this shell can run: builtins, plus every
// executable on PATH. Builtins come first so that a builtin shadows a PATH entry of
// the same name to execute it instead of the PATH entry.
func NewCompleter() *Completer {
	seen := make(map[string]struct{})
	commands := make([]string, 0, len(builtins))

	add := func(name string) {
		if _, ok := seen[name]; ok {
			return
		}
		seen[name] = struct{}{}
		commands = append(commands, name)
	}

	for _, builtin := range builtins {
		add(builtin)
	}

	for _, path := range filepath.SplitList(os.Getenv("PATH")) {
		entries, err := os.ReadDir(path)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			info, err := entry.Info()
			if err != nil || info.IsDir() || info.Mode().Perm()&0111 == 0 {
				continue
			}

			add(info.Name())
		}
	}

	slices.Sort(commands)

	return &Completer{commands: commands}
}

// Do completes the word under the cursor: the first word on the line against the
// known commands, every later word against the filesystem. Each word is completed on
// its own, so "cat app/ f<TAB>" matches "f" in the current directory rather than
// somewhere under app/.
func (c *Completer) Do(line []rune, pos int) ([][]rune, int) {
	full := string(line[:pos])

	word := full
	var matches []string

	// If there's a space, complete the argument
	// Otherwise, complete the command (first word)
	if spaceIdx := strings.LastIndex(full, " "); spaceIdx > 1 {
		// If there's a completer for the command, use it; otherwise, complete the argument
		parts := strings.Fields(full)
		command := parts[0]
		if completer, ok := completionRegistry[command]; ok {
			partialWord := ""
			if len(parts) > 1 {
				partialWord = parts[len(parts)-1]
			}
			previousWord := ""
			if len(parts) > 2 {
				previousWord = parts[len(parts)-2]
			}

			output, err := execOutput(completer, command, partialWord, previousWord)
			if err != nil {
				return nil, 0
			}
			// If the output is empty, do not change the input and ring the bell
			if len(output) == 0 {
				c.reset()
				bell()
				return nil, 0
			}
			normalizedOutput := strings.TrimSpace(string(output))

			if len(partialWord) > 1 {
				return [][]rune{completion(normalizedOutput, len(partialWord))}, 0
			}
			return [][]rune{completion(normalizedOutput, 0)}, 0
		} else {
			word = parts[len(parts)-1]
			matches = matchFilesAndDirs(word)
		}
	} else {
		matches = matchPrefix(c.commands, word)
	}

	wordLen := len([]rune(word))

	// Ring the bell on no matches
	if len(matches) == 0 {
		c.reset()
		bell()
		return nil, 0
	}

	// One match, complete it outright
	if len(matches) == 1 {
		c.reset()
		return [][]rune{completion(matches[0], wordLen)}, wordLen
	}

	// On subsequent <TAB>s, do a partial completion with the longest common prefix
	commonPrefix := longestCommonPrefix(matches)
	if len(commonPrefix) > wordLen {
		return [][]rune{commonPrefix[wordLen:]}, wordLen
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
	slices.Sort(matches)
	fmt.Fprintf(os.Stdout, "\n%s\n$ %s", strings.Join(matches, "  "), full)

	return nil, 0
}

func (c *Completer) reset() {
	c.belled = false
	c.belledFor = ""
}

func bell() {
	fmt.Fprintf(os.Stderr, "%c", 0x07)
}

// completion is the part of name still left to type. Directories are left open so
// that the next <TAB> can descend into them; anything else closes the word with a
// space.
func completion(name string, wordLen int) []rune {
	if !strings.HasSuffix(name, "/") {
		name += " "
	}

	return []rune(name)[wordLen:]
}

func matchPrefix(candidates []string, prefix string) []string {
	results := []string{}

	for _, candidate := range candidates {
		if strings.HasPrefix(candidate, prefix) {
			results = append(results, candidate)
		}
	}

	return results
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

// matchFilesAndDirs matches files and directories against word, a partially typed
// path, and spells each match with the directory exactly as the user typed it.
// Matches are compared against the whole word, so, e.g., "app/main.go" is what has
// to come back for "app/ma".
func matchFilesAndDirs(word string) []string {
	dir, dirPrefix, namePrefix := "./", "", word
	// if there's a slash, we need to match files in that directory
	if slashIdx := strings.LastIndex(word, "/"); slashIdx >= 0 {
		dir = word[:slashIdx+1]
		dirPrefix = dir
		namePrefix = word[slashIdx+1:]
	}

	results := []string{}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return results
	}

	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasPrefix(name, namePrefix) {
			continue
		}

		if entry.IsDir() {
			name += "/"
		}

		results = append(results, dirPrefix+name)
	}

	return results
}
