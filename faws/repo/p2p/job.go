package p2p

import (
	"fmt"

	"github.com/faws-vcs/faws/faws/repo/cas"
	"github.com/faws-vcs/faws/faws/repo/event"
	"github.com/faws-vcs/faws/faws/repo/p2p/tracker"
)

type Job interface {
	// Blocks until the job is complete or Cancel is called
	Wait()
	Cancel()
}

type subscription_job interface {
	Job
}

func (subscription *subscription) set_current_job(subscription_job subscription_job) {
	subscription.guard_job.Lock()
	if subscription.job != nil {
		subscription.job.Cancel()
		subscription.job = nil
	}
	subscription.job = subscription_job
	subscription.guard_job.Unlock()
}

// Pull: start a job to retrieve specific objects from the network
func (agent *Agent) Pull(topic tracker.Topic, objects ...cas.ContentID) (job Job, err error) {
	subscription, is_subscribed := agent.get_subscription(topic)
	if !is_subscribed {
		err = fmt.Errorf("%w: %s", ErrNotSubscribed, topic)
		return
	}

	pull_job_ := new(pull_job)
	pull_job_.init(subscription, objects)

	subscription.set_current_job(pull_job_)
	job = pull_job_

	return
}

// Clone: download all objects attached to the manifest once and then finishes
func (agent *Agent) Clone(topic tracker.Topic) (job Job, err error) {
	subscription, is_subscribed := agent.get_subscription(topic)
	if !is_subscribed {
		err = fmt.Errorf("%w: %s", ErrNotSubscribed, topic)
		return
	}

	// download the manifest
	_, err = subscription.update_manifest()
	if err != nil {
		return
	}

	var pull_tags_stage event.NotifyParams
	pull_tags_stage.Stage = event.StagePullTags
	subscription.agent.options.notify(event.NotifyBeginStage, &pull_tags_stage)

	var tags_in_queue event.NotifyParams
	tags_in_queue.Count = int64(len(subscription.manifest_info.Tags))
	subscription.agent.options.notify(event.NotifyTagQueueCount, &tags_in_queue)

	var objects []cas.ContentID
	for _, tag := range subscription.manifest_info.Tags {
		previous_tag_commit_, _ := subscription.repository.ReadTag(tag.Name)

		objects = append(objects, tag.CommitHash)
		if err = subscription.repository.WriteTag(tag.Name, tag.CommitHash); err != nil {
			return
		}

		var notify_pull_tag event.NotifyParams
		notify_pull_tag.Name1 = tag.Name
		notify_pull_tag.Object1 = previous_tag_commit_
		notify_pull_tag.Object2 = tag.CommitHash
		subscription.agent.options.notify(event.NotifyPullTag, &notify_pull_tag)
	}

	pull_tags_stage.Success = true
	subscription.agent.options.notify(event.NotifyCompleteStage, &pull_tags_stage)

	pull_job_ := new(pull_job)
	pull_job_.init(subscription, objects)

	subscription.set_current_job(pull_job_)
	return
}

// Mirror: start a job

// Seed: basically does nothing except provide a Job that blocks until the agent shuts down or is canceled
func (agent *Agent) Seed(topic tracker.Topic) (job Job, err error) {
	subscription, is_subscribed := agent.get_subscription(topic)
	if !is_subscribed {
		err = fmt.Errorf("%w: %s", ErrNotSubscribed, topic)
		return
	}

	seed_job_ := new(seed_job)
	seed_job_.init()

	subscription.set_current_job(seed_job_)
	job = seed_job_

	return
}
