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

type PeerState uint8

const (
	PeerDisconnected = iota
	PeerConnected
	NumChannelStates
)

func (channel_state PeerState) String() (s string) {
	if channel_state > NumChannelStates {
		return "?"
	}
	s = []string{"disconnected", "connected"}[channel_state]
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
	write_lock            sync.Mutex
	read_lock             sync.Mutex
	data_channel          *webrtc.DataChannel
	incoming_data_channel chan []byte
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
	case webrtc.PeerConnectionStateDisconnected:
		peer_connection.set_state(PeerDisconnected)
	case webrtc.PeerConnectionStateConnected:
	}
}

func (peer_connection *peer_connection) set_state(state int32) {
	old_state := peer_connection.state.Load()
	if peer_connection.state.Load() != state {
		if peer_connection.state.CompareAndSwap(old_state, state) {
			if !peer_connection.topic_channel.client.is_shutdown.Load() {
				peer_connection.topic_channel.client.peer_update_handler(peer_connection.topic_channel.topic, peer_connection.peer, PeerState(state))
			}
		}
	}
}

func (peer_connection *peer_connection) handle_incoming_data() {
	var (
		expected_bytes int
		buffer         *bytes.Buffer
	)

	for {
		data, ok := <-peer_connection.incoming_data_channel
		if !ok {
			return
		}

	process_data:
		for {
			if expected_bytes == 0 {
				// read header
				if len(data) < 5 {
					console.Println("peer sent message without valid header", len(data))
					peer_connection.close()
					return
				}
				var header [5]byte
				copy(header[:], data[:5])

				expected_bytes = int(binary.LittleEndian.Uint32(header[:4]))
				if expected_bytes == 0 {
					console.Println("peer sent empty message")
					peer_connection.close()
					return
				}
				message_id := MessageID(header[4])

				if err := validate_message_header(uint32(expected_bytes)-1, message_id); err != nil {
					console.Println("peer sent invalid", err)
					peer_connection.close()
					return
				}

				data = data[5:]
				buffer = new(bytes.Buffer)
				buffer.Grow(expected_bytes + 1)
				buffer.Write(header[4:])
			}

			fragment_size := min(len(data), (expected_bytes - buffer.Len()))
			buffer.Write(data[:fragment_size])
			data = data[fragment_size:]

			if buffer.Len() == expected_bytes {
				buffer_bytes := buffer.Bytes()
				peer_connection.topic_channel.client.channel_message_handler(peer_connection.topic_channel.topic, peer_connection.peer, MessageID(buffer_bytes[0]), buffer_bytes[1:])
				buffer = nil
				expected_bytes = 0
			}

			if len(data) == 0 {
				break process_data
			}
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
			peer_connection.set_state(PeerConnected)
			peer_connection.incoming_data_channel = make(chan []byte)
			go peer_connection.handle_incoming_data()
		})
		// data channel message arrives in-order, but fragmented.
		// since data channels are limited to a particular size (~16KB)
		// we reassemble our own messages which may be up to 16MB in size
		// theoretically we can just use the pion Detach API but...
		// it's got some annoying quirks I'm not a fan of
		data_channel.OnMessage(func(data_channel_message webrtc.DataChannelMessage) {
			peer_connection.handle_message_fragment(data_channel_message.Data)
		})
		data_channel.OnClose(func() {
			peer_connection.set_state(PeerDisconnected)
			close(peer_connection.incoming_data_channel)
		})
	}
	peer_connection.data_channel = data_channel
}

func (peer_connection *peer_connection) handle_message_fragment(fragment []byte) {
	peer_connection.incoming_data_channel <- fragment
}

const maximum_fragment_size = 16384

func (peer_connection *peer_connection) send(message_id MessageID, message []byte) (err error) {
	if peer_connection.state.Load() == PeerConnected {
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
			upper = min(uint32(len(message_full)), upper)
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
	peer_connection.set_state(PeerDisconnected)

}
