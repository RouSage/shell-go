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
