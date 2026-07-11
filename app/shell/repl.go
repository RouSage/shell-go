package shell

import (
	"bufio"
	"fmt"
	"os"
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

		cmd := NewCommand(command)
		if cmd == nil {
			continue
		}

		cmd.handle()
	}
}
