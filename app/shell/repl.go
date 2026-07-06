package shell

import (
	"bufio"
	"fmt"
	"os"
	"slices"
	"strings"
)

func REPL() {
	reader := bufio.NewReader(os.Stdin)
	for {
		// Wait for user input
		fmt.Print("$ ")

		// Read the command
		command, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading input:", err)
			os.Exit(1)
		}
		command = strings.TrimSpace(command)

		args := parseArgs(command)
		if len(args) == 0 {
			continue
		}

		handleCommand(args[0], args[1:])
	}
}

func handleCommand(command string, args []string) {
	if slices.Contains(builtins, command) {
		builtinCMD(command, args...)
	} else if _, err := lookPath(command); err == nil {
		execCMD(command, args...)
	} else {
		// Print the error message
		fmt.Printf("%s: command not found", command)
	}
}
