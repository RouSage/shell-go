package shell

import "os/exec"

// Background jobs
type Job struct {
	id   int
	name string
	cmd  *exec.Cmd
	done bool
}

var jobMap = make(map[int]*Job)
