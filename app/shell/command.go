package shell

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
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
	builtinCd   builtin = "cd"
)

var builtins = []builtin{builtinEcho, builtinExit, builtinType, builtinPwd, builtinCd}
var redirectOperators = []string{">", "1>", "2>"}

type Command struct {
	command string
	args    []string
	stdout  io.Writer
	stderr  io.Writer
}

func NewCommand(command string) *Command {
	args := parseArgs(command)
	if len(args) == 0 {
		return nil
	}

	return &Command{
		command: args[0],
		args:    args[1:],
		stdout:  os.Stdout,
		stderr:  os.Stderr,
	}
}

func (c *Command) handle() {
	if len(c.args) > 2 {
		var fileToClose *os.File

		arg := c.args[len(c.args)-2]
		switch arg {
		case ">", "1>":
			c.args = handleRedirect(c.args, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, &c.stdout, &fileToClose)
		case "2>":
			c.args = handleRedirect(c.args, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, &c.stderr, &fileToClose)
		case ">>", "1>>":
			c.args = handleRedirect(c.args, os.O_APPEND|os.O_CREATE|os.O_WRONLY, &c.stdout, &fileToClose)
		case "2>>":
			c.args = handleRedirect(c.args, os.O_APPEND|os.O_CREATE|os.O_WRONLY, &c.stderr, &fileToClose)
		}

		if fileToClose != nil {
			defer fileToClose.Close()
		}
	}

	if slices.Contains(builtins, c.command) {
		c.builtinCMD()
	} else if _, err := lookPath(c.command); err == nil {
		c.execCMD()
	} else {
		// Print the error message
		fmt.Fprintf(c.stdout, "%s: command not found\n", c.command)
	}
}

func (c *Command) builtinCMD() {
	switch c.command {
	case builtinExit:
		os.Exit(0)
	case builtinEcho:
		fmt.Fprintln(c.stdout, strings.Join(c.args, " "))
	case builtinPwd:
		c.pwdCMD()
	case builtinCd:
		c.cdCMD()
	case builtinType:
		c.typeCMD()
	}
}

func (c *Command) execCMD() {
	cmd := exec.Command(c.command, c.args...)
	cmd.Stderr = c.stderr
	cmd.Stdout = c.stdout

	cmd.Run()
}

func (c *Command) typeCMD() {
	command := c.args[0]

	if slices.Contains(builtins, command) {
		fmt.Fprintf(c.stdout, "%s is a shell builtin\n", command)
	} else if path, err := lookPath(command); err == nil {
		fmt.Fprintf(c.stdout, "%s is %s\n", command, path)
	} else {
		fmt.Fprintf(c.stdout, "%s: not found\n", command)
	}
}

func (c *Command) pwdCMD() {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(c.stderr, err)
	}

	fmt.Fprintln(c.stdout, dir)
}

func (c *Command) cdCMD() {
	dir := c.args[0]

	if dir == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintln(c.stderr, err)
			return
		}
		dir = home
	}

	_, err := os.Stat(dir)
	if errors.Is(err, fs.ErrNotExist) {
		fmt.Printf("%s: %s: No such file or directory\n", builtinCd, dir)
		return
	}

	err = os.Chdir(dir)
	if err != nil {
		fmt.Fprintln(c.stderr, err)
	}
}
