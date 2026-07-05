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
	builtinType builtin = "type"
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
		} else if strings.HasPrefix(command, builtinType+" ") {
			typeCMD(strings.TrimPrefix(command, builtinType+" "))
		} else {
			// Print the error message
			fmt.Println(command + ": command not found")
		}
	}
}

func typeCMD(command string) {
	switch command {
	case builtinEcho, builtinExit, builtinType:
		fmt.Println(command + "is a shell builtin")
	default:
		fmt.Println(command + ": not found")
	}
}
