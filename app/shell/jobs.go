package shell

import (
	"fmt"
	"io"
	"os/exec"
	"sort"
)

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

func listJobs(w io.Writer, doneOnly bool) {
	jobsLen := len(jobMap)

	keys := make([]int, 0, jobsLen)
	for _, job := range jobMap {
		keys = append(keys, job.id)
	}
	sort.Ints(keys)

	for idx, key := range keys {
		job := jobMap[key]
		marker := " "
		switch idx + 1 {
		case jobsLen:
			marker = "+"
		case jobsLen - 1:
			marker = "-"
		}

		status := "Running"
		trailing := " &"
		if job.done {
			status = "Done"
			trailing = ""
		}

		if doneOnly {
			if job.done {
				fmt.Fprintf(w, "[%d]%s  %-24s%s%s\n", job.id, marker, status, job.name, trailing)
			}
		} else {
			fmt.Fprintf(w, "[%d]%s  %-24s%s%s\n", job.id, marker, status, job.name, trailing)
		}

		if job.done {
			delete(jobMap, job.id)
		}
	}
}
