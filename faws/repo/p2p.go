package repo

import (
	"github.com/faws-vcs/console"
	"github.com/faws-vcs/faws/faws/identity"
	"github.com/faws-vcs/faws/faws/repo/cas"
	"github.com/faws-vcs/faws/faws/repo/event"
	"github.com/faws-vcs/faws/faws/repo/p2p"
	"github.com/faws-vcs/faws/faws/repo/p2p/tracker"
)

func (repo *Repository) clone_p2p() (err error) {
	var topic tracker.Topic
	err = tracker.ParseTopicURI(repo.config.Origin, &topic)
	if err != nil {
		return
	}

	var agent p2p.Agent
	if err = agent.Init(
		p2p.WithNotify(repo.notify),
		p2p.WithTrackerURL(repo.tracker_url),
	); err != nil {
		return
	}

	if err = agent.Subscribe(topic, repo); err != nil {
		return
	}

	var job p2p.Job
	job, err = agent.Clone(topic)
	if err != nil {
		return
	}
	job.Wait()

	agent.Shutdown()

	return
}

func (repo *Repository) pull_p2p(ref ...string) (err error) {
	objects := make([]cas.ContentID, len(ref))
	for i := range ref {
		objects[i], err = repo.ParseRef(ref[i])
		if err != nil {
			return
		}
	}

	var topic tracker.Topic
	err = tracker.ParseTopicURI(repo.config.Origin, &topic)
	if err != nil {
		return
	}

	var agent p2p.Agent
	if err = agent.Init(
		p2p.WithNotify(repo.notify),
		p2p.WithTrackerURL(repo.tracker_url),
	); err != nil {
		return
	}

	if err = agent.Subscribe(topic, repo); err != nil {
		return
	}

	var job p2p.Job
	job, err = agent.Pull(topic, objects...)
	if err != nil {
		return
	}
	job.Wait()

	agent.Shutdown()

	return
}

func (repo *Repository) fetch_manifest_info(manifest_info *tracker.ManifestInfo) (err error) {
	var topic tracker.Topic
	err = tracker.ParseTopicURI(repo.config.Origin, &topic)
	if err != nil {
		return
	}

	var tracker_client tracker.Client
	err = tracker_client.Init(repo.tracker_url, nil)
	if err != nil {
		return
	}

	var manifest_bytes []byte
	manifest_bytes, err = tracker_client.FetchManifest(topic.Hash().String())
	if err != nil {
		return
	}

	var manifest tracker.Manifest

	err = tracker.DecodeManifest(manifest_bytes, &manifest)
	if err != nil {
		return
	}
	if manifest.Publisher != topic.Publisher {
		err = p2p.ErrEvilServer
		return
	}

	if !identity.Verify(topic.Publisher, &manifest.Signature, manifest.Info) {
		err = p2p.ErrEvilServer
		return
	}

	err = tracker.DecodeManifestInfo(manifest.Info, topic, manifest_info)
	return
}

func (repo *Repository) pull_tags_p2p() (err error) {
	var manifest_info tracker.ManifestInfo
	if err = repo.fetch_manifest_info(&manifest_info); err != nil {
		return
	}

	// begin to pull tags
	var notify_pull_tags event.NotifyParams
	notify_pull_tags.Stage = event.StagePullTags
	repo.notify(event.NotifyBeginStage, &notify_pull_tags)
	// notify success if err == nil
	defer func() {
		if err == nil {
			notify_pull_tags.Success = true
		}
		repo.notify(event.NotifyCompleteStage, &notify_pull_tags)
	}()

	// notify tag count
	var tags_in_queue event.NotifyParams
	tags_in_queue.Count = int64(len(manifest_info.Tags))
	repo.notify(event.NotifyTagQueueCount, &tags_in_queue)

	for _, tag := range manifest_info.Tags {
		previous_tag_commit, _ := repo.read_tag(tag.Name)

		if err = repo.WriteTag(tag.Name, tag.CommitHash); err != nil {
			return
		}

		var notify_pull_tag event.NotifyParams
		notify_pull_tag.Name1 = tag.Name
		notify_pull_tag.Object1 = previous_tag_commit
		notify_pull_tag.Object2 = tag.CommitHash
		repo.notify(event.NotifyPullTag, &notify_pull_tag)
	}

	return
}

func (repo *Repository) pull_tag_p2p(tags ...string) (err error) {
	var manifest_info tracker.ManifestInfo
	if err = repo.fetch_manifest_info(&manifest_info); err != nil {
		return
	}

	tag_lookup := make(map[string]cas.ContentID)
	for _, tag := range manifest_info.Tags {
		tag_lookup[tag.Name] = tag.CommitHash
	}

	// begin to pull tags
	var notify_pull_tags event.NotifyParams
	notify_pull_tags.Stage = event.StagePullTags
	repo.notify(event.NotifyBeginStage, &notify_pull_tags)
	// notify success if err == nil
	defer func() {
		if err == nil {
			notify_pull_tags.Success = true
		}
		repo.notify(event.NotifyCompleteStage, &notify_pull_tags)
	}()

	// notify tag count
	var tags_in_queue event.NotifyParams
	tags_in_queue.Count = int64(len(tags))
	repo.notify(event.NotifyTagQueueCount, &tags_in_queue)

	for _, tag := range tags {
		var current_tag_commit cas.ContentID
		current_tag_commit, _ = repo.read_tag(tag)

		remote_tag_commit, tag_found := tag_lookup[tag]
		if !tag_found {
			err = ErrRefNotFound
			return
		}

		if err = repo.WriteTag(tag, remote_tag_commit); err != nil {
			return
		}

		var notify_pull_tag event.NotifyParams
		notify_pull_tag.Name1 = tag
		notify_pull_tag.Object1 = current_tag_commit
		notify_pull_tag.Object2 = remote_tag_commit
		repo.notify(event.NotifyPullTag, &notify_pull_tag)
	}

	return
}

func (repo *Repository) Seed(topic tracker.Topic, peer_identity identity.Pair) (err error) {
	if peer_identity == identity.Nil {
		peer_identity, err = identity.New()
		if err != nil {
			return
		}
	}

	if topic.Repository != repo.config.UUID {
		err = ErrTopicRepositoryMismatch
		return
	}

	var agent p2p.Agent
	if err = agent.Init(
		p2p.WithNotify(repo.notify),
		p2p.WithTrackerURL(repo.tracker_url),
		p2p.WithIdentity(peer_identity),
	); err != nil {
		return
	}

	if err = agent.Subscribe(topic, repo); err != nil {
		return
	}

	var job p2p.Job
	job, err = agent.Seed(topic)
	if err != nil {
		return
	}
	job.Wait()
	console.Println("done seeding i guess")

	agent.Shutdown()
	return
}
