package shell

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// Pipeline is one input line: a sequence of commands joined by `|`. A line
// without any `|` is simply a pipeline of length one, so there is a single code
// path for both cases.
type Pipeline struct {
	cmds       []*Command
	background bool
	stdout     io.Writer
	stderr     io.Writer
}

func NewPipeline(input string) *Pipeline {
	segments := parseSegments(input)
	if len(segments) == 0 {
		return nil
	}

	p := &Pipeline{
		cmds:   make([]*Command, 0, len(segments)),
		stdout: os.Stdout,
		stderr: os.Stderr,
	}

	// A trailing `&` backgrounds the whole line, so strip it before the last
	// segment is turned into a Command.
	last := segments[len(segments)-1]
	if len(last) > 0 && last[len(last)-1] == "&" {
		p.background = true
		segments[len(segments)-1] = last[:len(last)-1]
	}

	for _, args := range segments {
		cmd := newCommand(args)
		if cmd == nil {
			return nil
		}

		p.cmds = append(p.cmds, cmd)
	}

	if len(p.cmds) == 0 {
		return nil
	}

	return p
}

func (p *Pipeline) Run() {
	if len(p.cmds) == 1 {
		c := p.cmds[0]
		c.background = p.background
		c.handle()
	} else {
		p.runStages()
	}

	// Report background jobs that finished while the line was running
	if len(p.cmds) != 1 || p.cmds[0].command != builtinJobs {
		listJobs(p.stdout, true)
	}
}

func (p *Pipeline) String() string {

	segments := make([]string, 0, len(p.cmds))
	for _, cmd := range p.cmds {
		segments = append(segments, cmd.String())
	}

	return strings.Join(segments, " | ")
}

// runStages wires the commands together with pipes and runs them concurrently.
func (p *Pipeline) runStages() {
	// every pipe has a read and write end, store them in a slice
	// so that they can be closed after the pipeline is done
	pipeEnds := make([]*os.File, 0, 2*(len(p.cmds)-1))

	// set stdin/stdout for each command
	// first command reads from stdin, last command writes to stdout
	for i := range len(p.cmds) - 1 {
		pr, pw, err := os.Pipe()
		if err != nil {
			fmt.Fprintf(p.stderr, "pipeline failed: %s\n", err)
			closeAll(pipeEnds)
			return
		}

		p.cmds[i].stdout = pw
		p.cmds[i+1].stdin = pr
		pipeEnds = append(pipeEnds, pr, pw)
	}

	// apply redirects for each command
	// redirects are applied before the command is started, because it overwrites stdout
	redirectClosers := make([]func(), 0, len(p.cmds))
	for _, cmd := range p.cmds {
		redirectClosers = append(redirectClosers, cmd.applyRedirects())
	}

	// start each command concurrently
	// starting them sequentially can lead to deadlocks, when output reaches the pipe buffer,
	// so Start() instead of Run()
	execCmds := make([]*exec.Cmd, len(p.cmds))
	for i, cmd := range p.cmds {
		execCmd := cmd.newExecCmd()
		err := execCmd.Start()
		if err != nil {
			fmt.Fprintf(p.stderr, "cannot start \"%s\" command: %s\n", cmd.String(), err)
			execCmds[i] = nil
			continue
		}

		execCmds[i] = execCmd
	}

	// close all pipe ends, children will have a copy of them, so we can close them here
	closeAll(pipeEnds)

	// wait for all commands to complete
	wait := func() {
		for _, execCmd := range execCmds {
			if execCmd == nil {
				continue
			}

			execCmd.Wait()
		}

		// close all redirects after the pipeline is done
		for _, closer := range redirectClosers {
			closer()
		}
	}

	if !p.background {
		wait()
		return
	}

	// background execution registers one job for the whole pipeline
	// and reports a pid of the last command if it's not nil
	last := execCmds[len(execCmds)-1]
	if last == nil {
		go wait()
		return
	}

	job := addJob(p.String(), last)
	go func(jobId int) {
		wait()
		jobMap[jobId].done = true
	}(job.id)
	fmt.Fprintln(p.stdout, job.String())
}
