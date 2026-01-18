package p2p

import (
	"bytes"
	"runtime"
	"sync"
	"time"

	"github.com/faws-vcs/faws/faws/identity"
	"github.com/faws-vcs/faws/faws/repo/cas"
	"github.com/faws-vcs/faws/faws/repo/p2p/tracker"
	"github.com/faws-vcs/faws/faws/repo/queue"
)

type subscription struct {
	agent *Agent

	// a peer is tracked here once it connects from the peernet. Pretty simple.
	guard_peers sync.RWMutex
	peers       map[identity.ID]*peer

	// here the job is placed
	guard_job sync.Mutex
	job       subscription_job

	// stores the most recent version of the manifest
	guard_manifest sync.RWMutex
	manifest_bytes []byte
	manifest_info  tracker.ManifestInfo
	manifest_time  time.Time

	// the topic of the subscription. its UUID has to match the repository
	topic      tracker.Topic
	repository Repository

	object_server_channels       []chan object_request
	object_server_error_channels []chan error

	// all the objects we wanted
	object_wishlist                queue.TaskHeap[cas.ContentID]
	object_request_limiter         *time.Ticker
	object_receiver_channels       []chan named_object
	object_receiver_error_channels []chan error
	shutdown_channel               chan struct{}
}

func (subscription *subscription) init(agent *Agent, topic tracker.Topic, repository Repository) (err error) {
	subscription.agent = agent
	subscription.topic = topic
	subscription.repository = repository

	subscription.object_wishlist.Init()
	subscription.object_request_limiter = time.NewTicker(time.Second / time.Duration(agent.options.requests_per_second))
	subscription.peers = make(map[identity.ID]*peer)

	// Spawn servers: workers that receive incoming requests and process them
	num_servers := max(runtime.NumCPU()/2, 2)
	subscription.object_server_error_channels = make([]chan error, num_servers)
	subscription.object_server_channels = make([]chan object_request, num_servers)
	for i := 0; i < num_servers; i++ {
		object_server_error_channel := make(chan error)
		object_server_channel := make(chan object_request, 512)
		subscription.object_server_channels[i] = object_server_channel
		go subscription.spawn_object_server(object_server_error_channel, object_server_channel)
	}

	// Spawn receivers: workers that receive objects as they come in
	num_receivers := max(runtime.NumCPU()/2, 2)
	subscription.object_receiver_error_channels = make([]chan error, num_receivers)
	subscription.object_receiver_channels = make([]chan named_object, num_receivers)
	for i := 0; i < num_receivers; i++ {
		object_receiver_error_channel := make(chan error)
		object_receiver_channel := make(chan named_object, 1)
		subscription.object_receiver_channels[i] = object_receiver_channel
		go subscription.spawn_object_receiver(object_receiver_error_channel, object_receiver_channel)
	}
	subscription.shutdown_channel = make(chan struct{})

	go func() {
		<-subscription.shutdown_channel
		// stop the job
		subscription.guard_job.Lock()
		if subscription.job != nil {
			subscription.job.Cancel()
		}
		subscription.guard_job.Unlock()
		subscription.object_request_limiter.Stop()
		// stop the server workers
		for _, object_server_channel := range subscription.object_server_channels {
			close(object_server_channel)
		}
		// receive errors from stopped workers
		for _, object_server_error_channel := range subscription.object_server_error_channels {
			<-object_server_error_channel
		}

		// stop the receiver workers
		for _, object_receiver_channel := range subscription.object_receiver_channels {
			close(object_receiver_channel)
		}
		// receive errors from stopped workers
		for _, object_receiver_error_channel := range subscription.object_receiver_error_channels {
			<-object_receiver_error_channel
		}

	}()

	// subscribe to other peers
	subscription.agent.peernet_client.Subscribe(topic)

	return
}

func (subscription *subscription) shutdown() {
	subscription.shutdown_channel <- struct{}{}
	close(subscription.shutdown_channel)
}

// refreshes the manifest and queues
func (subscription *subscription) update_manifest() (manifest_changed bool, err error) {
	subscription.guard_manifest.Lock()
	defer subscription.guard_manifest.Unlock()
	// Download the manifest
	tracker_client := subscription.agent.peernet_client.Tracker()

	var (
		manifest_bytes []byte
		manifest       tracker.Manifest
		manifest_info  tracker.ManifestInfo
	)
	manifest_bytes, err = tracker_client.FetchManifest(subscription.topic.Hash().String())
	if err != nil {
		return
	}

	err = tracker.DecodeManifest(manifest_bytes, &manifest)
	if err != nil {
		return
	}
	// if the manifest doesn't even purport to belong to the correct identity, something is wrong with the server
	if manifest.Publisher != subscription.topic.Publisher {
		err = ErrEvilServer
		return
	}

	// again, if the server isn't checking that the manifest was signed by the identity
	// something is tampering with the server
	if !identity.Verify(subscription.topic.Publisher, &manifest.Signature, manifest.Info) {
		err = ErrEvilServer
		return
	}

	err = tracker.DecodeManifestInfo(manifest.Info, subscription.topic, &manifest_info)
	if err != nil {
		return
	}

	manifest_changed = true

	if subscription.manifest_bytes != nil {
		// ignore this manifest if it is older than the one we currently have
		// (indicates something fucked up happening, but we can manage)
		if manifest.Time().Before(subscription.manifest_time) {
			manifest_changed = false
			return
		}
		if bytes.Equal(subscription.manifest_bytes, manifest_bytes) {
			manifest_changed = false
		}
	}

	subscription.manifest_bytes = manifest_bytes
	subscription.manifest_time = manifest.Time()
	subscription.manifest_info = manifest_info

	return
}
