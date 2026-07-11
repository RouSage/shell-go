package shell

import (
	"fmt"
	"io"
	"os"
	"os/exec"
)

func lookPath(command string) (string, error) {
	return exec.LookPath(command)
}

func handleRedirect(args []string, flag int, writer *io.Writer, file **os.File) []string {
	outputFile, err := os.OpenFile(args[len(args)-1], flag, 0644)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	*writer = outputFile
	*file = outputFile

	return args[:len(args)-2]
}
