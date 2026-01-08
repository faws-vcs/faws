package main

import (
	"fmt"
	"os"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/faws-vcs/faws/faws/identity"
	"github.com/faws-vcs/faws/faws/repo/p2p/peernet"
	"github.com/faws-vcs/faws/faws/repo/p2p/tracker"
)

func main() {
	var (
		state peernet.ChannelState = peernet.ChannelDisconnected
	)

	var peernet_client peernet.Client
	if err := peernet_client.Init(peernet.WithTrackerURL("http://localhost:48775")); err != nil {
		panic(err)
	}

	peernet_client.OnChannelUpdate(func(topic tracker.Topic, peer identity.ID, channel_state peernet.ChannelState) {
		fmt.Println(topic, peer, channel_state)
		state = channel_state
	})
	peernet_client.OnMessage(func(topic tracker.Topic, peer identity.ID, message_id peernet.MessageID, message []byte) {
		fmt.Println(topic, peer, message_id, spew.Sdump(message))
	})

	var topic tracker.Topic
	var err error
	if err = tracker.ParseTopicURI(os.Args[1], &topic); err != nil {
		fmt.Println(err)
		return
	}

	go func() {
		for {
			if state == peernet.ChannelActive {
				peernet_client.Broadcast(topic, peernet.Chat, []byte("hello world!"))
			}
			time.Sleep(5 * time.Second)
		}
	}()

	peernet_client.Subscribe(topic)
	c := make(chan bool)
	<-c
}
