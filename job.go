package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

const (
	// MinExecTimeMs is the minimal execution runtime for a job in ms
	MinExecTimeMs = 10
	// MaxExecTimeMs is the maximal execution runtime for a job in ms
	MaxExecTimeMs = 100
)

// Job represents a Kube Job or CronJob resource.
type Job struct {
	ID                string        `yaml:"id"`
	Name              string        `yaml:"name"`
	Starttime         int64         `yaml:"-"`
	Endtime           int64         `yaml:"-"`
	Exectime          time.Duration `yaml:"-"`
	Status            string        `yaml:"-"`
	Dependents        []string      `yaml:"deps"`
	CompletedUpstream chan bool     `yaml:"-"`
	Every             int           `yaml:"every"`
}

// New creates a new job.
func newjob(id, name string, numupstream int) Job {
	j := Job{
		ID:                id,
		Name:              name,
		Status:            "scheduled",
		CompletedUpstream: make(chan bool, numupstream),
	}
	return j
}

// AddDep adds one or more dependent jobs by ID.
func (j *Job) adddep(depj ...string) {
	for _, d := range depj {
		j.Dependents = append(j.Dependents, d)
	}
}

// Launch launches a job, making sure it's only executed
// when all upstream jobs have completed.
func (j *Job) launch(dj DependentJobs, wg *sync.WaitGroup) {
	j.wait4upstream()
	fmt.Printf(j.render("Launched"))
	j.execute(dj, wg)
	// fmt.Printf("%s notifying my dependents: %v\n", j.Name, j.Dependents)
	for _, did := range j.Dependents {
		d := dj.Lookup(did)
		if dj.TimeToRun(did) {
			go d.launch(dj, wg)
		} else {
			wg.Done()
			dj.callseq <- d
		}
	}
}

func (j Job) execute(dj DependentJobs, wg *sync.WaitGroup) {
	defer wg.Done()
	et := time.Duration(1000000 * (MinExecTimeMs + rand.Intn(MaxExecTimeMs-MinExecTimeMs)))
	j.Exectime = et
	j.Starttime = time.Now().UnixNano()
	time.Sleep(et)
	j.Endtime = time.Now().UnixNano()
	j.Status = "completed"
	fmt.Printf(j.render("Executed"))
	dj.callseq <- j
	for _, did := range j.Dependents {
		d := dj.Lookup(did)
		d.CompletedUpstream <- true
	}
}

func (j Job) wait4upstream() {
	upstreamcount := cap(j.CompletedUpstream)
	if upstreamcount == 0 {
		return
	}
	i := 0
	for {
		select {
		case <-j.CompletedUpstream:
			i++
		}
		if upstreamcount == i {
			break
		}
	}
}

func (j Job) render(msg string) string {
	now := time.Now().UnixNano()
	return fmt.Sprintf("%v| %s: %#v\n", now, msg, j)
}

func (j *Job) periodic(every int) {
	j.Every = every
}

// GoString return a canonical string represenation of a Job
func (j Job) GoString() string {
	return fmt.Sprintf("<ID: %v, Every: %d, Status: %v, Exectime: %v, Deps: %v>", j.ID, j.Every+1, j.Status, j.Exectime, j.Dependents)
}
