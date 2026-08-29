package shell

import (
	"fmt"
	"io"
	"os"
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

// runStages wires the commands together with pipes and runs them concurrently.
// TODO: steps 3-4 of PIPE.md — os.Pipe() wiring, start-all/close-all/wait-all.
func (p *Pipeline) runStages() {
	fmt.Fprintln(p.stderr, "pipelines are not implemented yet")
}
