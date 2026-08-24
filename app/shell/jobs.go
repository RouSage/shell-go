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

func addJob(name string, cmd *exec.Cmd) *Job {
	job := &Job{
		id:   len(jobMap) + 1,
		name: name,
		cmd:  cmd,
		done: false,
	}
	jobMap[job.id] = job

	return job
}
