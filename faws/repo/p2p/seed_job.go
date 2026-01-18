package p2p

import (
	"sync"

	"github.com/faws-vcs/faws/faws/repo/event"
)

type seed_job struct {
	wait_group sync.WaitGroup
}

func (seed_job *seed_job) init(subscription *subscription) {
	var serving_objects event.NotifyParams
	serving_objects.Stage = event.StageServeObjects
	subscription.agent.options.notify(event.NotifyBeginStage, &serving_objects)
	seed_job.wait_group.Add(1)
}

func (seed_job *seed_job) Wait() {
	seed_job.wait_group.Wait()
}

func (seed_job *seed_job) Cancel() {
	seed_job.wait_group.Done()
}
