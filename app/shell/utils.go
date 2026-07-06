package shell

import "os/exec"

func lookPath(command string) (string, error) {
	return exec.LookPath(command)
}
