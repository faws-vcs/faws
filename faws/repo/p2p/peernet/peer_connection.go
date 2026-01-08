package peernet

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"sync"
	"sync/atomic"

	"github.com/faws-vcs/console"
	"github.com/faws-vcs/faws/faws/identity"
	"github.com/faws-vcs/faws/faws/repo/p2p/tracker"
	"github.com/pion/webrtc/v4"
)

type ChannelState uint8

const (
	ChannelDisconnected = iota
	ChannelActive
	ChannelClosed
	NumChannelStates
)

func (channel_state ChannelState) String() (s string) {
	if channel_state > NumChannelStates {
		return "?"
	}
	s = []string{"disconnected", "active", "closed"}[channel_state]
	return
}

var (
	channel_id            uint16 = 1440
	channel_is_negotiated bool   = true
	channel_is_ordered    bool   = true

	default_ice_servers = []webrtc.ICEServer{
		{
			URLs: []string{"stun:stun.l.google.com:19302"},
		},
	}
)

type peer_connection struct {
	// each peer connection is created on the basis of a topic (a repository published by a publisher)
	topic_channel *topic_channel
	// the identity of the peer we are communicating with
	peer identity.ID
	// our WebRTC connection to the peer
	connection *webrtc.PeerConnection
	//
	write_lock   sync.Mutex
	read_lock    sync.Mutex
	data_channel *webrtc.DataChannel
	// data_channel datachannel.ReadWriteCloser
	expected_bytes  uint32
	current_message *bytes.Buffer
	//
	is_choked bool
	//
	state atomic.Int32
	//
}

func (peer_connection *peer_connection) handle_sdp_offer(topic tracker.Topic, data []byte) (err error) {
	var (
		offer  webrtc.SessionDescription
		answer webrtc.SessionDescription
	)
	offer, err = decode_sdp(data)
	if err != nil {
		return
	}

	if err = peer_connection.connection.SetRemoteDescription(offer); err != nil {
		return
	}

	// create answer
	var answer_options webrtc.AnswerOptions
	answer_options.ICETricklingSupported = true
	answer, err = peer_connection.connection.CreateAnswer(&answer_options)
	if err != nil {
		return
	}
	peer_connection.connection.SetLocalDescription(answer)

	var answer_data []byte
	answer_data, err = encode_sdp(answer)
	if err != nil {
		return
	}
	err = peer_connection.topic_channel.client.tracker_client.Signal(topic, peer_connection.peer, tracker.AnswerSDP, answer_data)
	return
}

func (peer_connection *peer_connection) sdp_offer(topic tracker.Topic) (err error) {
	var (
		offer         webrtc.SessionDescription
		offer_options webrtc.OfferOptions
		offer_data    []byte
	)
	offer_options.ICETricklingSupported = true
	offer, err = peer_connection.connection.CreateOffer(&offer_options)
	if err != nil {
		return
	}
	if err = peer_connection.connection.SetLocalDescription(offer); err != nil {
		return
	}
	offer_data, err = encode_sdp(offer)
	if err != nil {
		return
	}
	err = peer_connection.topic_channel.client.tracker_client.Signal(topic, peer_connection.peer, tracker.OfferSDP, offer_data)
	if err != nil {
		return
	}
	return
}

func (peer_connection *peer_connection) handle_sdp_answer(data []byte) (err error) {
	var (
		answer webrtc.SessionDescription
	)
	answer, err = decode_sdp(data)
	if err != nil {
		return
	}

	if err = peer_connection.connection.SetRemoteDescription(answer); err != nil {
		return
	}

	return
}

func (peer_connection *peer_connection) handle_ice_candidate(data []byte) (err error) {
	var ice_candidate_init webrtc.ICECandidateInit
	ice_candidate_init, err = decode_ice_candidate(data)
	if err != nil {
		return
	}
	err = peer_connection.connection.AddICECandidate(ice_candidate_init)
	return
}

func decode_sdp(data []byte) (session_description webrtc.SessionDescription, err error) {
	// if len(data) < 4 {
	// 	err = fmt.Errorf("faws/p2p/peernet: invalid sdp message")
	// }
	// session_description.Type = webrtc.SDPType(binary.LittleEndian.Uint32(data[:4]))
	// data = data[4:]
	// session_description.SDP = string(data)
	err = json.Unmarshal(data, &session_description)
	return
}

func encode_sdp(session_description webrtc.SessionDescription) (data []byte, err error) {
	// var sdp_type [4]byte
	// binary.LittleEndian.PutUint32(sdp_type[:], uint32(session_description.Type))
	// data = append(data, sdp_type[:]...)
	// data = append(data, []byte(session_description.SDP)...)
	data, err = json.Marshal(&session_description)
	return
}

func decode_ice_candidate(data []byte) (ice_candidate webrtc.ICECandidateInit, err error) {
	err = json.Unmarshal(data, &ice_candidate)
	return
}

func encode_ice_candidate(ice_candidate webrtc.ICECandidateInit) (data []byte, err error) {
	data, err = json.Marshal(ice_candidate)
	return
}

func (peer_connection *peer_connection) handle_connection_state_change(pcs webrtc.PeerConnectionState) {
	switch pcs {
	case webrtc.PeerConnectionStateClosed:
		peer_connection.topic_channel.close(peer_connection.peer)
	}
}

func (peer_connection *peer_connection) set_state(state int32) {
	old_state := peer_connection.state.Load()
	if peer_connection.state.Load() != state {
		if peer_connection.state.CompareAndSwap(old_state, state) {
			peer_connection.topic_channel.client.channel_update_handler(peer_connection.topic_channel.topic, peer_connection.peer, ChannelState(state))
		}
	}
}

func (peer_connection *peer_connection) create_data_channel() {
	var (
		data_channel_init webrtc.DataChannelInit
		data_channel      *webrtc.DataChannel
		err               error
	)
	data_channel_init.ID = &channel_id
	data_channel_init.Negotiated = &channel_is_negotiated
	data_channel_init.Ordered = &channel_is_ordered
	data_channel, err = peer_connection.connection.CreateDataChannel("faws peernet v1", &data_channel_init)
	if err == nil {
		data_channel.OnOpen(func() {
			peer_connection.set_state(ChannelActive)
		})
		// data channel message arrives in-order, but fragmented.
		// since data channels are limited to a particular size (~16KB)
		// we reassemble our own messages which may be up to 16MB in size
		// theoretically we can just use the pion Detach API but...
		// it's got some annoying quirks I'm not a fan of
		data_channel.OnMessage(func(data_channel_message webrtc.DataChannelMessage) {
			peer_connection.handle_message_packet(data_channel_message.Data)
		})
		data_channel.OnClose(func() {
			peer_connection.set_state(ChannelClosed)
		})
	}
	peer_connection.data_channel = data_channel
}

func (peer_connection *peer_connection) handle_message_packet(fragment []byte) {
	peer_connection.read_lock.Lock()
	// console.Println("received message packet", spew.Sdump(fragment))
	err := peer_connection.handle_message_fragment(fragment)
	if err != nil {
		console.Println("error handling message packet", err)
	}
	peer_connection.read_lock.Unlock()
	return
}

func (peer_connection *peer_connection) handle_message_fragment(fragment []byte) (err error) {
	if len(fragment) == 0 {
		return
	}

	fragment_size := uint32(len(fragment))
	if peer_connection.expected_bytes > 0 {
		if peer_connection.expected_bytes > fragment_size {
			peer_connection.current_message.Write(fragment)
			peer_connection.expected_bytes -= fragment_size
			return
		} else if peer_connection.expected_bytes == fragment_size {
			// received message!
			message_buffer := peer_connection.current_message

			peer_connection.current_message.Write(fragment)

			peer_connection.expected_bytes = 0
			peer_connection.current_message = nil

			message := message_buffer.Bytes()
			message_id := MessageID(message[0])
			message = message[1:]

			peer_connection.topic_channel.client.channel_message_handler(peer_connection.topic_channel.topic, peer_connection.peer, message_id, message)

			return
		} else if peer_connection.expected_bytes < fragment_size {
			// received message + extra fragment
			message_buffer := peer_connection.current_message

			peer_connection.current_message.Write(fragment[:peer_connection.expected_bytes])
			fragment = fragment[peer_connection.expected_bytes:]

			peer_connection.expected_bytes = 0
			peer_connection.current_message = nil

			message := message_buffer.Bytes()
			message_id := MessageID(message[0])
			message = message[1:]

			peer_connection.topic_channel.client.channel_message_handler(peer_connection.topic_channel.topic, peer_connection.peer, message_id, message)

			err = peer_connection.handle_message_fragment(fragment)
			return
		}
	}

	// this is a badly formatted fragment, < 4 is invalid, and each message must have 1 byte for message id
	if fragment_size < 5 {
		err = ErrInvalidMessageSize
		return
	}

	peer_connection.expected_bytes = binary.LittleEndian.Uint32(fragment[:4])
	if peer_connection.expected_bytes < 1 {
		err = ErrInvalidMessageSize
		return
	}

	fragment = fragment[4:]

	peer_connection.current_message = new(bytes.Buffer)
	peer_connection.current_message.Write(fragment[0:1])
	message_id := MessageID(fragment[0])
	fragment = fragment[1:]

	peer_connection.expected_bytes--

	if err = validate_message_header(peer_connection.expected_bytes, message_id); err != nil {
		return
	}

	err = peer_connection.handle_message_fragment(fragment)
	return
}

const maximum_fragment_size = 16384

func (peer_connection *peer_connection) send(message_id MessageID, message []byte) (err error) {
	if peer_connection.state.Load() == ChannelActive {
		peer_connection.write_lock.Lock()
		defer peer_connection.write_lock.Unlock()

		// build full message buffer
		// TODO: a more efficient mechanism would be wise here
		var message_header [5]byte
		binary.LittleEndian.PutUint32(message_header[:], uint32(len(message)+1))
		message_header[4] = byte(message_id)
		var message_buffer bytes.Buffer
		message_buffer.Write(message_header[:])
		message_buffer.Write(message[:])

		message_full := message_buffer.Bytes()

		// break message into fragments
		fragment_count := (uint32(message_buffer.Len()) / maximum_fragment_size) + 1

		for i := uint32(0); i < fragment_count; i++ {
			lower := i * maximum_fragment_size
			upper := (i + 1) * maximum_fragment_size
			if upper > uint32(len(message_full)) {
				upper = uint32(len(message_full))
			}
			if err = peer_connection.data_channel.Send(message_full[lower:upper]); err != nil {
				return
			}
		}
	} else {
		console.Println("not active!")
	}
	return
}

func (peer_connection *peer_connection) close() {
	peer_connection.data_channel.Close()
	peer_connection.connection.Close()
	peer_connection.set_state(ChannelClosed)

}
