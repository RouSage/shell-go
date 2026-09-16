package shell

import (
	"os"
	"os/exec"
	"strings"
)

func lookPath(command string) (string, error) {
	return exec.LookPath(command)
}

func execOutput(command string, arg ...string) ([]byte, error) {
	cmd := exec.Command(command, arg...)
	return cmd.CombinedOutput()
}

func closeAll(files []*os.File) {
	for _, file := range files {
		file.Close()
	}
}

func loadHistory(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var history []string
	for line := range strings.Lines(string(data)) {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		history = append(history, line)
	}

	return history, nil
}

func saveHistory(path string, history []string) error {
	data := historyToString(history)
	if err := os.WriteFile(path, []byte(data), 0644); err != nil {
		return err
	}

	return nil
}

func appendHistory(path string, history []string) error {
	data := historyToString(history)

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.WriteString(data); err != nil {
		return err
	}

	return nil
}

func historyToString(history []string) string {
	return strings.Join(history, "\n") + "\n"
}
