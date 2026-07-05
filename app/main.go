package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"
)

type builtin = string

const (
	builtinEcho builtin = "echo"
	builtinExit builtin = "exit"
	builtinType builtin = "type"
)

var builtins = []builtin{builtinEcho, builtinExit, builtinType}

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
	if slices.Contains(builtins, command) {
		fmt.Println(command + " is a shell builtin")
	} else if path, err := exec.LookPath(command); err == nil {
		fmt.Printf("%s is %s\n", command, path)
	} else {
		fmt.Println(command + ": not found")

	}
}
