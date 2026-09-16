package shell

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
)

// Pipeline is one input line: a sequence of commands joined by `|`. A line
// without any `|` is simply a pipeline of length one, so there is a single code
// path for both cases.
type Pipeline struct {
	sh *Shell

	cmds       []*Command
	background bool
	stdout     io.Writer
	stderr     io.Writer
	wg         sync.WaitGroup
}

func NewPipeline(shell *Shell, input string) *Pipeline {
	segments := parseSegments(input)
	if len(segments) == 0 {
		return nil
	}

	p := &Pipeline{
		sh:     shell,
		cmds:   make([]*Command, 0, len(segments)),
		stdout: os.Stdout,
		stderr: os.Stderr,
	}

	// A trailing `&` backgrounds the whole line, so strip it before the last
	// segment is turned into a Command.
	last := segments[len(segments)-1]
	if len(last) > 0 && last[len(last)-1].op && last[len(last)-1].value == "&" {
		p.background = true
		segments[len(segments)-1] = last[:len(last)-1]
	}

	for _, args := range segments {
		cmd := newCommand(p.sh, args)
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
		p.sh.jobs.List(p.stdout, true)
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
	pipeEnds, writeEnds := p.wireStages()
	if pipeEnds == nil {
		return
	}
	// apply redirects for each command
	// redirects are applied before the command is started, because it overwrites stdout
	redirectClosers := make([]func(), 0, len(p.cmds))
	for _, cmd := range p.cmds {
		redirectClosers = append(redirectClosers, cmd.applyRedirects())
	}

	execCmds := p.startStages(writeEnds)
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

		p.wg.Wait()

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

	job := p.sh.jobs.AddJob(p.String(), last)
	go func(jobId int) {
		wait()
		p.sh.jobs.MarkJobDone(jobId)
	}(job.id)
	fmt.Fprintln(p.stdout, job.String())
}

// wireStages sets up the pipeline by connecting commands with pipes and returns the pipe ends
func (p *Pipeline) wireStages() ([]*os.File, []*os.File) {
	// every pipe has a read and write end, store them in a slice
	// so that they can be closed after the pipeline is done
	pipeEnds := make([]*os.File, 0, 2*(len(p.cmds)-1))
	// every pipe has a write end, or nil for the last command
	// captured here, before applyRedirects can replace cmd.stdout with a file
	writeEnds := make([]*os.File, len(p.cmds))

	// set stdin/stdout for each command
	// first command reads from stdin, last command writes to stdout
	for i := range len(p.cmds) - 1 {
		pr, pw, err := os.Pipe()
		if err != nil {
			fmt.Fprintf(p.stderr, "pipeline failed: %s\n", err)
			closeAll(pipeEnds)
			return nil, nil
		}

		p.cmds[i].stdout = pw
		p.cmds[i+1].stdin = pr
		writeEnds[i] = pw

		pipeEnds = append(pipeEnds, pr)
		if !p.cmds[i].isBuiltin() {
			pipeEnds = append(pipeEnds, pw)
		}
	}

	return pipeEnds, writeEnds
}

// startStages starts each command in a pipeline concurrently
// starting them sequentially can lead to deadlocks, when output reaches the pipe buffer,
// so Start() instead of Run()
func (p *Pipeline) startStages(writeEnds []*os.File) []*exec.Cmd {
	execCmds := make([]*exec.Cmd, len(p.cmds))

	for i, cmd := range p.cmds {
		if cmd.isBuiltin() {
			p.wg.Add(1)
			go func(cmd *Command, w *os.File) {
				defer p.wg.Done()
				if w != nil {
					defer w.Close()
				}

				cmd.runBuiltinStage()
			}(cmd, writeEnds[i])
			continue
		}

		if _, err := lookPath(cmd.command); err != nil {
			fmt.Fprintf(p.stderr, "%s: command not found\n", cmd.command)
			continue
		}

		execCmd := cmd.newExecCmd()
		if err := execCmd.Start(); err != nil {
			fmt.Fprintf(p.stderr, "cannot start \"%s\" command: %s\n", cmd.String(), err)
			continue
		}

		execCmds[i] = execCmd
	}

	return execCmds
}
