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
	builtinEcho     builtin = "echo"
	builtinExit     builtin = "exit"
	builtinType     builtin = "type"
	builtinPwd      builtin = "pwd"
	builtinCd       builtin = "cd"
	builtinComplete builtin = "complete"
	builtinJobs     builtin = "jobs"
)

var builtins = []builtin{builtinEcho, builtinExit, builtinType, builtinPwd, builtinCd, builtinComplete, builtinJobs}

type Command struct {
	command string
	args    []string
	// flag for the background execution – (&) operator
	background bool
	stdin      io.Reader
	stdout     io.Writer
	stderr     io.Writer
}

func newCommand(args []string) *Command {
	if len(args) == 0 {
		return nil
	}

	return &Command{
		command: args[0],
		args:    args[1:],
		stdin:   os.Stdin,
		stdout:  os.Stdout,
		stderr:  os.Stderr,
	}
}

func (c *Command) String() string {
	if len(c.args) == 0 {
		return c.command
	}

	return fmt.Sprintf("%s %s", c.command, strings.Join(c.args, " "))
}

// handle runs the command on its own: redirects are applied, the command runs,
// and any file opened for a redirect is closed again.
func (c *Command) handle() {
	closeRedirects := c.applyRedirects()
	defer closeRedirects()

	c.run()
}

// applyRedirects consumes a trailing redirect operator, if any, and points
// stdout/stderr at the target file. The returned func closes that file, so it
// must be called once the command is done writing.
func (c *Command) applyRedirects() func() {
	if len(c.args) < 2 {
		return func() {}
	}

	var fileToClose *os.File

	switch c.args[len(c.args)-2] {
	case ">", "1>":
		c.args = handleRedirect(c.args, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, &c.stdout, &fileToClose)
	case "2>":
		c.args = handleRedirect(c.args, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, &c.stderr, &fileToClose)
	case ">>", "1>>":
		c.args = handleRedirect(c.args, os.O_APPEND|os.O_CREATE|os.O_WRONLY, &c.stdout, &fileToClose)
	case "2>>":
		c.args = handleRedirect(c.args, os.O_APPEND|os.O_CREATE|os.O_WRONLY, &c.stderr, &fileToClose)
	}

	if fileToClose == nil {
		return func() {}
	}

	return func() {
		fileToClose.Close()
	}
}

// run dispatches to a builtin or an external binary.
// Redirects must already be applied.
func (c *Command) run() {
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
	case builtinComplete:
		c.completeCMD()
	case builtinType:
		c.typeCMD()
	case builtinJobs:
		c.jobsCMD()
	}
}

// newExecCmd builds the *exec.Cmd for an external command, wired to whatever
// stdin/stdout/stderr the command currently holds (a file, a pipe, or the shell's own streams).
func (c *Command) newExecCmd() *exec.Cmd {
	cmd := exec.Command(c.command, c.args...)
	cmd.Stdin = c.stdin
	cmd.Stdout = c.stdout
	cmd.Stderr = c.stderr

	return cmd
}

func (c *Command) execCMD() error {
	cmd := c.newExecCmd()

	if !c.background {
		return cmd.Run()
	}

	err := cmd.Start()
	if err != nil {
		return err
	}

	job := addJob(c.String(), cmd)
	go func(jobId int) {
		cmd.Wait()
		jobMap[jobId].done = true
	}(job.id)

	fmt.Fprintf(c.stdout, "[%d] %d\n", job.id, job.cmd.Process.Pid)

	return nil
}

func (c *Command) jobsCMD() {
	listJobs(c.stdout, false)
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

var completionRegistry = make(map[string]string)

func (c *Command) completeCMD() {
	if len(c.args) < 2 {
		return
	}

	if c.args[0] == "-C" && len(c.args) >= 3 {
		scriptPath := c.args[1]
		cmdName := c.args[2]
		completionRegistry[cmdName] = scriptPath
	} else if c.args[0] == "-p" {
		cmdName := c.args[1]

		if scriptPath, ok := completionRegistry[cmdName]; ok {
			fmt.Fprintf(c.stdout, "%s -C '%s' %s\n", builtinComplete, scriptPath, cmdName)
		} else {
			fmt.Fprintf(c.stderr, "%s: %s: no completion specification\n", builtinComplete, cmdName)
		}
	} else if c.args[0] == "-r" {
		delete(completionRegistry, c.args[1])
	}
}
