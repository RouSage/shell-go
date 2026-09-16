package shell

import (
	"fmt"
	"io"
	"os/exec"
	"sort"
	"sync"
)

// Background jobs
type Job struct {
	id   int
	name string
	cmd  *exec.Cmd
	done bool
}

func (job *Job) String() string {
	return fmt.Sprintf("[%d] %d", job.id, job.cmd.Process.Pid)
}

type Jobs struct {
	jobMap map[int]*Job
	nextID int
	mu     sync.Mutex
}

func NewJobs() *Jobs {
	return &Jobs{
		jobMap: make(map[int]*Job),
	}
}

func (j *Jobs) AddJob(name string, cmd *exec.Cmd) *Job {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.addJob(name, cmd)
}

func (j *Jobs) List(w io.Writer, doneOnly bool) {
	for _, line := range j.reap(doneOnly) {
		fmt.Fprint(w, line)
	}
}

func (j *Jobs) MarkJobDone(jobId int) {
	j.mu.Lock()
	defer j.mu.Unlock()
	if job, ok := j.jobMap[jobId]; ok {
		job.done = true
	}
}

func (j *Jobs) addJob(name string, cmd *exec.Cmd) *Job {
	j.nextID++
	job := &Job{
		id:   j.nextID,
		name: name,
		cmd:  cmd,
		done: false,
	}
	j.jobMap[job.id] = job

	return job
}

// reap formats the jobs to report and removes the finished ones. Formatting
// happens under the lock, writing does not: w is a pipe for `jobs | head`, and
// blocking on a full pipe buffer while holding j.mu would stall every
// background goroutine trying to mark itself done.
func (j *Jobs) reap(doneOnly bool) []string {
	j.mu.Lock()
	defer j.mu.Unlock()

	jobsLen := len(j.jobMap)

	keys := make([]int, 0, jobsLen)
	for _, job := range j.jobMap {
		keys = append(keys, job.id)
	}
	sort.Ints(keys)

	lines := make([]string, 0, jobsLen)
	for idx, key := range keys {
		job := j.jobMap[key]
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

		if !doneOnly || job.done {
			lines = append(lines, fmt.Sprintf("[%d]%s  %-24s%s%s\n", job.id, marker, status, job.name, trailing))
		}

		if job.done {
			delete(j.jobMap, job.id)
		}
	}

	return lines
}
