package p2p

import "sync"

type seed_job struct {
	wait_group sync.WaitGroup
}

func (seed_job *seed_job) init() {
	seed_job.wait_group.Add(1)
}

func (seed_job *seed_job) Wait() {
	seed_job.wait_group.Wait()
}

func (seed_job *seed_job) Cancel() {
	seed_job.wait_group.Done()
}
