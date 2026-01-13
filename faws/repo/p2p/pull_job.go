package p2p

import (
	"errors"
	"io"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/faws-vcs/console"
	"github.com/faws-vcs/faws/faws/repo/cas"
	"github.com/faws-vcs/faws/faws/repo/event"
)

type pull_job struct {
	subscription *subscription
	cancel       atomic.Bool
	done         sync.WaitGroup
}

func (pull_job *pull_job) init(subscription *subscription, initial_objects []cas.ContentID) {
	pull_job.subscription = subscription
	pull_job.done.Add(1)

	for _, object := range initial_objects {
		pull_job.subscription.object_wishlist.Push(object)
	}

	var notify_queue_count event.NotifyParams
	notify_queue_count.Count = int64(pull_job.subscription.object_wishlist.Len())
	pull_job.subscription.agent.options.notify(event.NotifyPullQueueCount, &notify_queue_count)

	go pull_job.spawn()
}

func (pull_job *pull_job) worker() (err error) {
	subscription := pull_job.subscription

	var (
		object_hash   cas.ContentID
		object_prefix cas.Prefix
		object_data   []byte
	)

	// in practice, this continously re-broadcasts peernet.WantObject
	// sometimes, on the same object repeatedly.
	// this may be understood if peers are continuously joining the swarm
	for {
		if pull_job.cancel.Load() {
			return
		}

		// select a random object
		object_hash, err = subscription.object_wishlist.Pick()
		if err != nil {
			if errors.Is(err, io.EOF) {
				err = nil
			}
			return
		}

		load_scale := float64(1) / float64(subscription.object_wishlist.AvailableLen())
		duration := time.Duration(float64(1500)*load_scale) * time.Millisecond
		time.Sleep(duration)

		// time.Sleep(time.Millisecond * 100 * time.Duration(rand.Intn(5)))

		// first, check that we have the object.
		_, err = subscription.repository.StatObject(object_hash)
		if err == nil {
			subscription.lock_object(object_hash)
			if subscription.object_wishlist.IsCompleted(object_hash) {
				subscription.unlock_object(object_hash)
				continue
			}
			// if we have the object already, process it
			object_prefix, object_data, err = pull_job.subscription.repository.LoadObject(object_hash)
			if err == nil {
				// treat it like we received it from a peer
				subscription.receive_object(object_hash, object_prefix, object_data)
				subscription.unlock_object(object_hash)
				continue
			}
			subscription.unlock_object(object_hash)
		}
		err = nil

		if err = subscription.place_object_order(object_hash); err != nil {
			return
		}
	}
}

func (pull_job *pull_job) start_worker(error_channel chan<- error) {
	error_channel <- pull_job.worker()
}

func (pull_job *pull_job) spawn() {
	var pulling_objects event.NotifyParams
	pulling_objects.Stage = event.StagePullObjects
	pull_job.subscription.agent.options.notify(event.NotifyBeginStage, &pulling_objects)

	num_workers := max(runtime.NumCPU()/2, 2)
	error_channels := make([]chan error, num_workers)

	for i := 0; i < num_workers; i++ {
		error_channels[i] = make(chan error)
		go pull_job.start_worker(error_channels[i])
	}

	pulling_objects.Success = true

	// wait for workers to notice that the queue of tasks is complete
	for i := 0; i < num_workers; i++ {
		if err := <-error_channels[i]; err != nil {
			console.Println(err)
			pulling_objects.Success = false
		}
	}

	// cause Wait() to return
	pull_job.done.Done()

	pull_job.subscription.agent.options.notify(event.NotifyCompleteStage, &pulling_objects)
}

func (pull_job *pull_job) Wait() {
	pull_job.done.Wait()
}

func (pull_job *pull_job) Cancel() {
	pull_job.cancel.Store(true)
}
