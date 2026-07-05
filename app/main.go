package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
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
		if command == "exit" {
			break
		}

		// Print the error message
		fmt.Printf("%s: command not found\n", command)
	}
}
