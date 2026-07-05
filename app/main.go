package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type builtin = string

const (
	builtinExit builtin = "exit"
	builtinEcho builtin = "echo"
)

func main() {
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
		// Trim whitespace from the command
		command = strings.TrimSpace(command)

		// Builtin "exit" command
		if command == builtinExit {
			break
		} else if strings.HasPrefix(command, builtinEcho+" ") {
			fmt.Println(strings.TrimPrefix(command, builtinEcho+" "))
		} else {
			// Print the error message
			fmt.Println(command + ": command not found")
		}
	}
}
