package p2p

import (
	"fmt"
	"sync"

	"github.com/faws-vcs/faws/faws/identity"
	"github.com/faws-vcs/faws/faws/repo/event"
	"github.com/faws-vcs/faws/faws/repo/p2p/peernet"
	"github.com/faws-vcs/faws/faws/repo/p2p/tracker"
)

// Agent an agent is responsible for transferring local and remote objects into a repository over
// a peer-to-peer network
type Agent struct {
	options             agent_options
	peernet_client      peernet.Client
	guard_subscriptions sync.Mutex
	subscriptions       map[tracker.Topic]*subscription
}

func (agent *Agent) Init(options ...Option) (err error) {
	agent.options.requests_per_second = 128
	agent.options.tracker_url = tracker.DefaultURL
	agent.options.notify = func(n event.Notification, params *event.NotifyParams) {}

	for _, option := range options {
		option(&agent.options)
	}
	if !agent.options.set_peer_identity {
		agent.options.peer_identity, err = identity.New()
		if err != nil {
			return
		}
	}

	agent.subscriptions = make(map[tracker.Topic]*subscription)

	if err = agent.peernet_client.Init(
		peernet.WithIdentity(agent.options.peer_identity),
		peernet.WithTrackerURL(agent.options.tracker_url),
		peernet.WithForceTURN(agent.options.use_turn),
	); err != nil {
		return
	}

	agent.set_peernet_handlers()
	return
}

func (agent *Agent) Shutdown() {
	agent.guard_subscriptions.Lock()
	var topics []tracker.Topic
	for topic := range agent.subscriptions {
		topics = append(topics, topic)
	}
	agent.guard_subscriptions.Unlock()

	for _, topic := range topics {
		agent.Unsubscribe(topic)
	}

	agent.peernet_client.Shutdown()
}

func (agent *Agent) get_subscription(topic tracker.Topic) (subscription *subscription, is_subscribed bool) {
	agent.guard_subscriptions.Lock()
	subscription, is_subscribed = agent.subscriptions[topic]
	agent.guard_subscriptions.Unlock()
	return
}

// Subscribe: connect a Topic to our repository.
// This is necessary to allow objects to flow between our repository and other peers.
// Important: the agent must be shutdown before the repo is closed.
func (agent *Agent) Subscribe(topic tracker.Topic, repo Repository) (err error) {
	agent.guard_subscriptions.Lock()
	subscription_, already_subscribed := agent.subscriptions[topic]
	if already_subscribed {
		err = fmt.Errorf("%w: %s", ErrAlreadySubscribed, topic)
	} else {
		subscription_ = new(subscription)
		err = subscription_.init(agent, topic, repo)
		if err == nil {
			agent.subscriptions[topic] = subscription_
		}
	}
	agent.guard_subscriptions.Unlock()
	return
}

// func (agent *Agent) Mirror()

func (agent *Agent) Unsubscribe(topic tracker.Topic) (err error) {
	agent.guard_subscriptions.Lock()
	subscription_, subscribed := agent.subscriptions[topic]
	if subscribed {
		subscription_.shutdown()
		delete(agent.subscriptions, topic)
	} else {
		err = fmt.Errorf("%w: %s", ErrNotSubscribed, topic)
	}
	agent.guard_subscriptions.Unlock()
	return
}
