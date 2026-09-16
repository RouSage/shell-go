package shell

import (
	"fmt"
	"io"
	"os"
	"strings"
)

type History struct {
	entries      []string
	lastAppended int
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
	h.entries = append(h.entries, entry)
}

func (h *History) Len() int {
	return len(h.entries)
}

func (h *History) Print(w io.Writer, start int) {
	for i := start; i < len(h.entries); i++ {
		fmt.Fprintf(w, "%4d  %s\n", i+1, h.entries[i])
	}
}

func (h *History) Read(path string) error {
	entries, err := read(path)
	if err != nil {
		return err
	}

	h.entries = append(h.entries, entries...)

	return nil
}

func (h *History) Write(path string) error {
	return h.write(path)
}

func (h *History) Append(path string) error {
	return h.appendHistory(path)
}

func (h *History) write(path string) error {
	data := entriesToString(h.entries)
	if err := os.WriteFile(path, []byte(data), 0644); err != nil {
		return err
	}

	return nil
}

func (h *History) appendHistory(path string) error {
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
	return strings.Join(entries, "\n") + "\n"
}
