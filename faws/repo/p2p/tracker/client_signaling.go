package tracker

import (
	"bytes"
	"fmt"
	"math/big"
	"net/url"
	"time"

	"github.com/faws-vcs/faws/faws/identity"
)

type (
	ClientSignalHandlerFunc func(topic Topic, peer identity.ID, signal Signal, message []byte)
	ClientPeerHandlerFunc   func(topic Topic, peer identity.ID)
)

func ignore_peer(topic Topic, peer identity.ID) {
}

func ignore_signal(topic Topic, peer identity.ID, signal Signal, message []byte) {

}

// type SignalingClient struct {
// 	client *Client
// }

// func (signaling_client *SignalingClient) Init(client *Client, topic )

type subscription struct {
	topic *Topic
}

// Subscribe starts the client receiving peer notifications and signals from the given topic channel
func (client *Client) Subscribe(topic Topic) {
	client.pending_subscriptions <- topic
}

func (client *Client) Unsubscribe(topic Topic) {
	client.pending_unsubscriptions <- topic
}

// OnSignal set a handler function to receive
func (client *Client) OnSignal(signal_handler ClientSignalHandlerFunc) {
	client.signal_handler = signal_handler
}

func (client *Client) OnPeer(peer_handler ClientPeerHandlerFunc) {
	client.peer_handler = peer_handler
}

// Transmit an signal to a peer subscribed to the given topic
func (client *Client) Signal(topic Topic, peer identity.ID, signal Signal, message []byte) (err error) {
	if client.closed.Load() {
		err = ErrClientIsShutdown
		return
	}

	topic_hash := topic.Hash()

	var (
		signal_buffer bytes.Buffer
		signature     identity.Signature
	)
	signal_buffer.WriteByte(3)
	signal_buffer.WriteByte(byte(signal))
	signal_buffer.Write(message)

	identity.Sign(&client.peer_identity, signal_buffer.Bytes(), &signature)

	encrypt(topic.Key(), signal_buffer.Bytes()[1:], signal_buffer.Bytes()[1:])

	var argument_buffer bytes.Buffer
	argument_buffer.Write(topic_hash[:])
	argument_buffer.Write(peer[:])
	argument_buffer.Write(signature[:])
	argument_buffer.Write(signal_buffer.Bytes()[1:])

	client.pending_commands <- command_message{sp_signal, argument_buffer.Bytes()}
	return
}

func (client *Client) login() (err error) {
	id := client.peer_identity.ID()
	client.connection.command(sp_authenticate, id[:])

	var (
		command signaling_protocol_command
		data    []byte
	)

loop:
	for {
		command, data, err = client.connection.receive_command()
		if err != nil {
			return
		}

		switch command {
		case sp_authenticate:
			break loop
		case sp_challenge:
			var challenge_buffer bytes.Buffer
			challenge_buffer.WriteByte(2)
			challenge_buffer.Write(data)

			var signature identity.Signature
			identity.Sign(&client.peer_identity, challenge_buffer.Bytes(), &signature)

			client.connection.command(sp_challenge, signature[:])
		case sp_kick:
			err = fmt.Errorf("faws/p2p/tracker: kicked by server: %s", data)
			return
		}
		if command != sp_challenge {
			err = ErrBadLogin
			return
		}
	}

	return
}

func (client *Client) subscribe(topic Topic) {
	client.guard_subscriptions.Lock()
	topic_hash := topic.Hash()
	client.subscriptions[topic_hash] = topic
	client.guard_subscriptions.Unlock()
	client.connection.command(sp_subscribe, topic_hash[:])
}

func (client *Client) unsubscribe(topic Topic) {
	client.guard_subscriptions.Lock()
	topic_hash := topic.Hash()
	client.subscriptions[topic_hash] = topic
	client.guard_subscriptions.Unlock()
	client.connection.command(sp_subscribe, topic_hash[:])
}

func (client *Client) handle_signal(data []byte) {
	if len(data) < TopicHashSize+identity.IDSize+identity.SignatureSize+1 {
		return
	}

	// server prefix         peer signal
	// [topic] [source peer] [ [signature] [encrypted message] ]
	var topic_channel TopicHash
	copy(topic_channel[:], data[:TopicHashSize])
	data = data[TopicHashSize:]

	var source_peer identity.ID
	copy(source_peer[:], data[:identity.IDSize])
	data = data[identity.IDSize:]

	var peer_signal_signature identity.Signature
	copy(peer_signal_signature[:], data[:identity.SignatureSize])
	data = data[identity.SignatureSize:]

	client.guard_subscriptions.Lock()
	topic, topic_found := client.subscriptions[topic_channel]
	client.guard_subscriptions.Unlock()

	if !topic_found {
		// console.Println("bad topic receive")
		return
	}

	decrypt(topic.Key(), data, data)
	if !identity.Verify(source_peer, &peer_signal_signature, append([]byte{3}, data...)) {
		// console.Println("dropped invalid message")
		return
	}

	client.signal_handler(topic, source_peer, Signal(data[0]), data[1:])
}

func (client *Client) handle_peer(data []byte) {
	var (
		topic_channel TopicHash
		peer          identity.ID
	)
	if len(data) < identity.IDSize+TopicHashSize {
		return
	}
	copy(topic_channel[:], data[:TopicHashSize])
	data = data[TopicHashSize:]

	copy(peer[:], data)
	data = data[identity.IDSize:]

	client.guard_subscriptions.Lock()
	topic, topic_found := client.subscriptions[topic_channel]
	client.guard_subscriptions.Unlock()

	if !topic_found {
		return
	}

	client.peer_handler(topic, peer)
}

func (client *Client) handle_command(command signaling_protocol_command, data []byte) (err error) {
	switch command {
	case sp_signal:
		client.handle_signal(data)
		return
	case sp_peer:
		client.handle_peer(data)
		return
	}

	return
}

func (client *Client) signaling_url() string {
	http_url := client.base_url + "/tracker/v1/signaling"
	u, err := url.Parse(http_url)
	if err != nil {
		panic(err)
	}
	switch u.Scheme {
	case "http":
		u.Scheme = "ws"
	case "https":
		u.Scheme = "wss"
	}
	return u.String()
}

func (client *Client) manage_signaling_connection(incoming_message chan<- []byte) {
	var err error

	cooldown_duration := func(miss_count int64) time.Duration {
		var n big.Int
		var limit big.Int
		limit.SetInt64(int64(time.Minute))
		n.Exp(big.NewInt(2), big.NewInt(miss_count+20), nil)
		if n.Cmp(&limit) > 0 {
			return time.Minute
		}
		return time.Duration(n.Int64())
	}

establish_connection:
	for {
		if client.closed.Load() {
			return
		}

		miss_count := int64(0)

		// establish the connection
		for {
			signaling_url := client.signaling_url()
			// console.Println("dial!")

			err = client.connection.dial(signaling_url)
			if err == nil {
				break
			}
			// console.Println(miss_count, "failed to connect to", signaling_url, err)
			if client.closed.Load() {
				return
			}

			miss_count++

			cooldown := cooldown_duration(miss_count)
			// console.Println("sleeping for", cooldown)
			time.Sleep(cooldown)
		}

		// login
		if err = client.login(); err != nil {
			continue establish_connection
		}

		// resume subscriptions
		client.guard_subscriptions.Lock()
		for subscription := range client.subscriptions {
			client.connection.command(sp_subscribe, subscription[:])
		}
		client.guard_subscriptions.Unlock()

		// read messages from the connection
		var (
			command signaling_protocol_command
			data    []byte
		)
		for {
			command, data, err = client.connection.receive_command()
			if err != nil {
				client.connection.Close()
				continue establish_connection
			}
			// console.Println("receive", command)

			if err = client.handle_command(command, data); err != nil {
				client.connection.Close()
				continue establish_connection
			}
		}
	}
}

func (client *Client) manage_signaling() {
	incoming_messages := make(chan []byte)

	go client.manage_signaling_connection(incoming_messages)

process:
	for {
		select {
		case topic_hash := <-client.pending_subscriptions:
			client.subscribe(topic_hash)
		case topic_hash := <-client.pending_unsubscriptions:
			client.unsubscribe(topic_hash)
		case command := <-client.pending_commands:
			client.connection.command(command.command, command.argument)
		case <-client.shutdown:
			client.connection.Close()
			break process
		}
	}

	client.closed.Store(true)
	close(client.pending_subscriptions)
	close(client.pending_commands)
}
