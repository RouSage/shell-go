package shell

import (
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
	"sync"
)

type History struct {
	entries      []string
	lastAppended int
	mu           sync.Mutex
}

func NewHistory() *History {
	entries := make([]string, 0)

	if historyFile := os.Getenv("HISTFILE"); historyFile != "" {
		fileEntries, err := read(historyFile)
		if err != nil {
			panic(err)
		}

		entries = fileEntries
	}

	return &History{
		entries:      entries,
		lastAppended: 0,
	}
}

func (h *History) Add(entry string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.entries = append(h.entries, entry)
}

func (h *History) Len() int {
	h.mu.Lock()
	defer h.mu.Unlock()

	return len(h.entries)
}

// Print writes entries from start onwards, numbered from 1 as bash does. The
// snapshot is taken under the lock and formatted without it, because w is a
// pipe for `history | head` and blocking on a full pipe buffer while holding
// mu would stall every other stage of the line.
func (h *History) Print(w io.Writer, start int) {
	h.mu.Lock()
	entries := slices.Clone(h.entries[min(start, len(h.entries)):])
	h.mu.Unlock()

	for i, entry := range entries {
		fmt.Fprintf(w, "%4d  %s\n", start+i+1, entry)
	}
}

func (h *History) Read(path string) error {
	entries, err := read(path)
	if err != nil {
		return err
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	h.entries = append(h.entries, entries...)

	return nil
}

func (h *History) Write(path string) error {
	h.mu.Lock()
	data := entriesToString(h.entries)
	h.mu.Unlock()

	return os.WriteFile(path, []byte(data), 0644)
}

// Append writes the entries added since the last Append. Unlike Write, the lock
// is held across the file write: taking entries[lastAppended:], writing it and
// advancing lastAppended is one read-modify-write, and splitting it would let
// two concurrent `history -a` stages duplicate or drop lines.
func (h *History) Append(path string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	data := entriesToString(h.entries[h.lastAppended:])

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.WriteString(data); err != nil {
		return err
	}

	h.lastAppended = len(h.entries)

	return nil
}

func read(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var entries []string
	for line := range strings.Lines(string(data)) {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		entries = append(entries, line)
	}

	return entries, nil
}

func entriesToString(entries []string) string {
	if len(entries) == 0 {
		return ""
	}

	return strings.Join(entries, "\n") + "\n"
}
