package main

import (
	"fmt"
	"os"

	"github.com/faws-vcs/faws/faws/identity"
	"github.com/faws-vcs/faws/faws/p2p/tracker"
)

func main() {
	id, err := identity.New()
	if err != nil {
		panic(err)
	}
	fmt.Println("my id is", id.ID())
	var client tracker.Client
	if err = client.Init("http://localhost:48775", &id); err != nil {
		return
	}

	client.OnPeer(func(topic tracker.Topic, peer identity.ID) {
		fmt.Println("peer", peer, topic.String())
		client.Signal(topic, peer, tracker.Chat, []byte("o hello!"))
	})

	client.OnSignal(func(topic tracker.Topic, peer identity.ID, signal tracker.Signal, message []byte) {
		fmt.Println("signal", peer, signal, topic.String())
		switch signal {
		case tracker.Chat:
			fmt.Println("msg: ", string(message))
		}
	})

	var topic tracker.Topic
	if err = tracker.ParseTopicURI(os.Args[1], &topic); err != nil {
		fmt.Println(err)
		return
	}
	client.Subscribe(topic)

	c := make(chan bool)
	<-c
}
