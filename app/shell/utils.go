package shell

import (
	"os"
	"os/exec"
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
