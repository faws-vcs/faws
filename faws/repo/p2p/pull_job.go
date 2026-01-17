package p2p

import (
	"math/big"
	"runtime"
	"sync"
	"sync/atomic"

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

func calc_range(id int, num_workers int) (min, max cas.ContentID) {
	if id >= num_workers {
		panic(id)
	}

	// domain = 2^160-1
	var domain big.Int
	domain.Exp(big.NewInt(2), big.NewInt(cas.ContentIDSize*8), nil)
	domain.Sub(&domain, big.NewInt(1))

	// fraction = domain / num_workers
	var fraction big.Int
	fraction.Div(&domain, big.NewInt(int64(num_workers)))

	// lower = fraction * id
	var lower big.Int
	lower.Mul(&fraction, big.NewInt(int64(id)))

	// upper = max(domain, fraction * id+1)
	var upper big.Int
	upper.Mul(&fraction, big.NewInt(int64(id+1)))
	var limit big.Int
	limit.Add(&upper, &fraction)
	if limit.Cmp(&domain) > 0 {
		upper = domain
	}

	copy(min[:], lower.Bytes())
	copy(max[:], upper.Bytes())

	return
}

func (pull_job *pull_job) customer_worker(task_channel <-chan cas.ContentID) (err error) {
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
		// select a random object
		var ok bool
		object_hash, ok = <-task_channel
		if !ok {
			return
		}

		// first, check that we have the object.
		_, err = subscription.repository.StatObject(object_hash)
		if err == nil {
			// if we have the object already, dispatch it to our object receivers
			object_prefix, object_data, err = pull_job.subscription.repository.LoadObject(object_hash)
			if err == nil {
				subscription.dispatch_object(false, object_hash, object_prefix, object_data)
				continue
			}
		}
		err = nil

		if err = subscription.place_object_order(object_hash); err != nil {
			return
		}
	}
}

func (pull_job *pull_job) spawn_customer(task_channel <-chan cas.ContentID, init_channel chan<- struct{}, error_channel chan<- error) {
	init_channel <- struct{}{}
	close(init_channel)
	error_channel <- pull_job.customer_worker(task_channel)
	close(error_channel)
}

func (pull_job *pull_job) spawn() {
	var pulling_objects event.NotifyParams
	pulling_objects.Stage = event.StagePullObjects
	pull_job.subscription.agent.options.notify(event.NotifyBeginStage, &pulling_objects)

	task_channel := make(chan cas.ContentID, 1024)

	// Spawn customers: workers that continuously pick through the heap of available objects
	// and repeatedly try to negotiate requests for those objects with other peers
	num_customers := max(runtime.NumCPU()/2, 2)
	// num_customers := 80
	customer_error_channels := make([]chan error, num_customers)
	for i := 0; i < num_customers; i++ {
		customer_init_channel := make(chan struct{})
		customer_error_channels[i] = make(chan error)
		go pull_job.spawn_customer(task_channel, customer_init_channel, customer_error_channels[i])
		<-customer_init_channel
	}

	// Pick random items from the wishlist repeatedly and send them to customers
	for {
		object, err := pull_job.subscription.object_wishlist.Pick()
		if err != nil {
			break
		}

		task_channel <- object
	}
	// once there are no more items, all customers are stopped.
	close(task_channel)

	pulling_objects.Success = true

	// wait for customers to stop running
	for i := 0; i < num_customers; i++ {
		if err := <-customer_error_channels[i]; err != nil {
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
