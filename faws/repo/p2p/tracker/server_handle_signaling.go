package tracker

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"time"

	"github.com/faws-vcs/faws/faws/identity"
	"github.com/faws-vcs/faws/faws/repo/p2p/tracker/models"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type peer_signal struct {
	source identity.ID
	target identity.ID
	data   []byte
}

type topic_channel struct {
	hash              TopicHash
	guard             sync.Mutex
	add_connection    chan *signaling_connection
	remove_connection chan *signaling_connection
	signal            chan peer_signal
}

func (s *Server) launch_topic_channel(channel *topic_channel) {
	channel.add_connection = make(chan *signaling_connection)
	channel.remove_connection = make(chan *signaling_connection)
	channel.signal = make(chan peer_signal)
	go func() {
		var (
			connections = make(map[identity.ID]*signaling_connection)
		)

	process:
		for {

			select {
			case new_conn := <-channel.add_connection:
				fmt.Println("new conn added", new_conn.id)
				connections[new_conn.id] = new_conn
				fmt.Println("notfying", new_conn.id, "of", len(connections), "connections")
				for _, peer := range connections {
					if peer.id != new_conn.id {
						fmt.Println("notfying", new_conn.id, "of", peer.id)
						new_conn.command(sp_peer, append(channel.hash[:], peer.id[:]...))
					}
				}
			case dead_conn := <-channel.remove_connection:
				delete(connections, dead_conn.id)
			case signal := <-channel.signal:
				if peer, ok := connections[signal.target]; ok {
					var argument bytes.Buffer
					argument.Write(channel.hash[:])
					argument.Write(signal.source[:])
					argument.Write(signal.data)

					peer.command(sp_signal, argument.Bytes())
				}
			case <-time.After(1 * time.Minute):
				// loop through and try to discard channel if unused
				// we attempt to discard the topic channel,
				// if no one is trying to access the list of channels
				if s.guard_topic_channels.TryLock() {
					if len(connections) == 0 {
						delete(s.topic_channels, channel.hash)
						close(channel.add_connection)
						close(channel.signal)
						s.guard_topic_channels.Unlock()
						break process
					}
					s.guard_topic_channels.Unlock()
				}
			}
		}
	}()
}

func (s *Server) find_topic_channel(topic_hash TopicHash) (channel *topic_channel, err error) {
	topic_name := topic_hash.String()
	// ensure that topic exists by looking at manifest
	name := filepath.Join(s.directory, "manifests", topic_name[0:2], topic_name[2:])
	// check if manifest already exists
	if _, err = os.Stat(name); err != nil {
		err = ErrBadTopicName
		return
	}

	var ok bool
	s.guard_topic_channels.Lock()
	channel, ok = s.topic_channels[topic_hash]
	if !ok {
		channel = new(topic_channel)
		channel.hash = topic_hash
		s.topic_channels[topic_hash] = channel
		s.launch_topic_channel(channel)
	}
	s.guard_topic_channels.Unlock()

	return
}

// routes a topic message from source connection to target peer
func (s *Server) route_signal(signaling_connection *signaling_connection, topic_hash TopicHash, target identity.ID, data []byte) (err error) {
	var topic_channel *topic_channel
	topic_channel, err = s.find_topic_channel(topic_hash)
	if err != nil {
		return
	}

	topic_channel.signal <- peer_signal{
		source: signaling_connection.id,
		target: target,
		data:   data,
	}
	return
}

func (s *Server) handle_signaling(rw http.ResponseWriter, r *http.Request) {
	// // validate the name of the topic
	// topic_name := r.PathValue("topic")
	// topic_channel, err := s.find_topic_channel(topic_name)
	// if err != nil {
	// 	s.respond(rw, http.StatusBadRequest, models.GenericResponse{Message: err.Error()})
	// 	return
	// }

	fmt.Println("new connection")

	var err error
	var signaling_connection signaling_connection
	signaling_connection.conn, err = upgrader.Upgrade(rw, r, nil)
	if err != nil {
		s.respond(rw, http.StatusBadRequest, models.GenericResponse{Message: err.Error()})
		return
	}

	if err = s.handle_signaling_login(&signaling_connection); err != nil {
		fmt.Println("failed login", err)
		signaling_connection.command(sp_kick, []byte(err.Error()))
		signaling_connection.conn.Close()
		return
	}

	var (
		command       signaling_protocol_command
		content       []byte
		subscriptions []TopicHash
	)

process:
	for {
		command, content, err = signaling_connection.receive_command()
		if err != nil {
			break process
		}
		fmt.Println("received command", command)
		switch command {
		case sp_subscribe:
			if len(subscriptions) > 256 {
				break process
			}
			if len(content) != TopicHashSize {
				err = ErrBadTopicName
				break process
			}
			var topic_hash TopicHash
			copy(topic_hash[:], content)

			if !slices.Contains(subscriptions, topic_hash) {
				if err = s.subscribe(&signaling_connection, topic_hash); err != nil {
					break process
				}
				subscriptions = append(subscriptions, topic_hash)
			}
		case sp_unsubscribe:
			if len(content) != TopicHashSize {
				err = ErrBadTopicName
				break process
			}
			var topic_hash TopicHash
			copy(topic_hash[:], content)

			index := slices.Index(subscriptions, topic_hash)
			if index != -1 {
				subscriptions = slices.Delete(subscriptions, index, index+1)
				if err = s.unsubscribe(&signaling_connection, topic_hash); err != nil {
					return
				}
			}
		case sp_signal:
			if len(content) < TopicHashSize+identity.IDSize {
				err = ErrBadCommand
				break process
			}
			var (
				topic_hash TopicHash
				target     identity.ID
			)
			copy(topic_hash[:], content[:TopicHashSize])
			content = content[TopicHashSize:]
			copy(target[:], content[:identity.IDSize])
			content = content[identity.IDSize:]
			if !slices.Contains(subscriptions, topic_hash) {
				err = ErrBadCommand
				break process
			}

			if err = s.route_signal(&signaling_connection, topic_hash, target, content); err != nil {
				break process
			}
		}
	}

	if err != nil {
		signaling_connection.command(sp_kick, []byte(err.Error()))
	}
	signaling_connection.conn.Close()

	for _, subscription := range subscriptions {
		s.unsubscribe(&signaling_connection, subscription)
	}
}

func (s *Server) subscribe(signaling_connection *signaling_connection, topic_hash TopicHash) (err error) {
	var topic_channel *topic_channel
	topic_channel, err = s.find_topic_channel(topic_hash)
	if err != nil {
		return
	}
	topic_channel.add_connection <- signaling_connection
	return
}

func (s *Server) unsubscribe(signaling_connection *signaling_connection, topic_hash TopicHash) (err error) {
	var topic_channel *topic_channel
	topic_channel, err = s.find_topic_channel(topic_hash)
	if err != nil {
		return
	}
	topic_channel.remove_connection <- signaling_connection
	return
}

// the client must issue a cryptographic identity to login
func (s *Server) handle_signaling_login(signaling_connection *signaling_connection) (err error) {
	// TODO:
	// introduce delays from low-trust IPs to frustrate flood attacks
	// or proof of work
	// also, we can bypass
	// otherwise, denial of service is likely here

	var (
		command signaling_protocol_command
		content []byte
	)

	command, content, err = signaling_connection.receive_command()
	if err != nil {
		return
	}
	if command != sp_authenticate {
		err = fmt.Errorf("%w: sp_authenticate expected", ErrBadLogin)
		return
	}
	if len(content) != identity.IDSize {
		err = fmt.Errorf("%w: authenticate command does not relay an identity", ErrBadLogin)
		return
	}
	copy(signaling_connection.id[:], content)
	if signaling_connection.id == identity.Nobody {
		err = fmt.Errorf("%w: authenticate command uses nobody as an identity", ErrBadLogin)
		return
	}
	fmt.Println("received id", signaling_connection.id)

	// send at least one (easy) challenge to prove that the user actually owns their purported identity
	var challenge_buffer bytes.Buffer
	// reserved byte
	challenge_buffer.WriteByte(2)

	// calculate size of challenge
	// minimum 100K
	// maximum 1MB
	var (
		minimum      int64 = 1e5
		maximum      int64 = 1e6
		random_value *big.Int
	)
	var max big.Int
	max.SetInt64(maximum - minimum)
	random_value, err = rand.Int(rand.Reader, &max)
	if err != nil {
		err = fmt.Errorf("%w: failed to generate challenge data length", ErrBadLogin)
		return
	}
	challenge_size := minimum + random_value.Int64()
	// generate challenge data
	if _, err = io.CopyN(&challenge_buffer, rand.Reader, challenge_size); err != nil {
		err = fmt.Errorf("%w: failed to generate challenge data", ErrBadLogin)
		return
	}
	// send challenge data to client
	signaling_connection.command(sp_challenge, challenge_buffer.Bytes()[1:])

	command, content, err = signaling_connection.receive_command()
	if err != nil {
		return
	}
	if command != sp_challenge {
		err = fmt.Errorf("%w: sp_challenge expected", ErrBadLogin)
		return
	}
	if len(content) != identity.SignatureSize {
		err = fmt.Errorf("%w: a signature of the challenge data was expected", ErrBadLogin)
		return
	}

	var (
		signature identity.Signature
	)
	copy(signature[:], content[:identity.SignatureSize])

	if !identity.Verify(signaling_connection.id, &signature, challenge_buffer.Bytes()) {
		err = fmt.Errorf("%w: signature was invalid", ErrBadLogin)
		return
	}

	if slices.Contains(s.config.PublisherWhitelist, signaling_connection.id) {
		goto after_pow
	}

	// TODO: proof of work challenge here
after_pow:
	signaling_connection.command(sp_authenticate, nil)
	return
}
