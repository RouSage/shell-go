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
	builtinPwd  builtin = "pwd"
)

var builtins = []builtin{builtinEcho, builtinExit, builtinType, builtinPwd}

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
		command = strings.TrimSpace(command)

		args := strings.Fields(command)
		if len(args) == 0 {
			os.Exit(0)
		}

		if slices.Contains(builtins, args[0]) {
			builtinCMD(args[0], args[1:]...)
		} else if _, err := lookPath(args[0]); err == nil {
			execCMD(args[0], args[1:]...)
		} else {
			// Print the error message
			fmt.Println(args[0] + ": command not found")
		}
	}
}

func builtinCMD(command builtin, args ...string) {
	switch command {
	case builtinExit:
		os.Exit(0)
	case builtinEcho:
		fmt.Println(strings.Join(args, " "))
	case builtinPwd:
		pwdCMD()
	case builtinType:
		typeCMD(args[0])
	}
}

func typeCMD(command string) {
	if slices.Contains(builtins, command) {
		fmt.Printf("%s is a shell builtin\n", command)
	} else if path, err := lookPath(command); err == nil {
		fmt.Printf("%s is %s\n", command, path)
	} else {
		fmt.Printf("%s: not found\n", command)
	}
}

func pwdCMD() {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	fmt.Println(dir)
}

func execCMD(command string, args ...string) {
	cmd := exec.Command(command, args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	err := cmd.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

func lookPath(command string) (string, error) {
	return exec.LookPath(command)
}
