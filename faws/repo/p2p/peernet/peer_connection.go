package peernet

import (
	"encoding/json"
	"errors"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"github.com/faws-vcs/faws/faws/app"
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

const message_ttl = time.Minute

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
	message_sequence      uint64
	incoming_data_channel chan fragment
	outgoing_data_channel chan fragment
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

const incoming_messages_gc_ttl = 1 * time.Minute

type incoming_message struct {
	fragment_bitfield [64]byte
	buffer            []byte
	bytes_received    uint64
}

func incoming_messages_gc(messages map[message_guid]*incoming_message) {
	moment := time.Now()
	for guid := range messages {
		if moment.Sub(time.UnixMilli(guid.Timestamp())) > incoming_messages_gc_ttl {
			delete(messages, guid)
		}
	}
}

func (incoming_message *incoming_message) set_fragment_bit(fragment_id uint16, value bool) {
	byte_index := fragment_id / 8
	byte_flag := (byte(1) << byte(fragment_id%8))
	if value {
		incoming_message.fragment_bitfield[byte_index] |= byte_flag
	} else {
		incoming_message.fragment_bitfield[byte_index] &= ^byte_flag
	}
}

func (incoming_message *incoming_message) get_fragment_bit(fragment_id uint16) (value bool) {
	value = incoming_message.fragment_bitfield[fragment_id/8]&(1<<(fragment_id%8)) != 0
	return
}

func (peer_connection *peer_connection) handle_incoming_data() {
	messages := make(map[message_guid]*incoming_message)

	gc_ticker := time.NewTicker(incoming_messages_gc_ttl)

message_loop:
	for {
		select {
		case <-gc_ticker.C:
			incoming_messages_gc(messages)
		case fragment, ok := <-peer_connection.incoming_data_channel:
			if !ok {
				break message_loop
			}
			moment := time.Now()
			// if the message is too old by now, drop it
			if moment.Sub(time.UnixMilli(fragment.Message.Timestamp())) > message_ttl {
				delete(messages, fragment.Message)
				continue message_loop
			}

			// if the single fragment contains the entire message payload, receive the message and skip tracking
			if fragment.Message.PayloadSize() == uint32(len(fragment.Data)) {
				peer_connection.topic_channel.client.channel_message_handler(peer_connection.topic_channel.topic, peer_connection.peer, fragment.Message.MessageID(), fragment.Data)
				continue message_loop
			}

			// lookup message by guid
			incoming_message_, message_is_already_tracked := messages[fragment.Message]
			if !message_is_already_tracked {
				incoming_message_ = new(incoming_message)
				// allocate buffer
				incoming_message_.buffer = make([]byte, fragment.Message.PayloadSize())
				// add to map
				messages[fragment.Message] = incoming_message_
			}

			// receive the fragment

			// TODO: discard messages that have a much higher amount of data intake than what they should be
			incoming_message_.bytes_received += 16 + 2 + uint64(len(fragment.Data))

			if incoming_message_.get_fragment_bit(fragment.Fragment) {
				// already received, nothing to do here
				continue message_loop
			}

			// the fragment bitfield should be checked now
			fragment_count := uint16((fragment.Message.PayloadSize() + (fragment_max_data_size - 1)) / fragment_max_data_size)
			fragment_offset := fragment_max_data_size * uint32(fragment.Fragment)

			// offset is already validated
			copy(incoming_message_.buffer[fragment_offset:], fragment.Data)

			incoming_message_.set_fragment_bit(fragment.Fragment, true)

			all_fragments_received := true

			for i := uint16(0); i < fragment_count; i++ {
				all_fragments_received = incoming_message_.get_fragment_bit(i)
				if !all_fragments_received {
					break
				}
			}

			if all_fragments_received {
				delete(messages, fragment.Message)
				peer_connection.topic_channel.client.channel_message_handler(peer_connection.topic_channel.topic, peer_connection.peer, fragment.Message.MessageID(), incoming_message_.buffer)
			}
		}

	}
	gc_ticker.Stop()
}

func (peer_connection *peer_connection) handle_outgoing_data() {
	for {
		fragment, ok := <-peer_connection.outgoing_data_channel
		if !ok {
			return
		}
		encoded_fragment, err := encode_fragment(&fragment)
		if err != nil {
			app.Warning(err)
			return
		}

		if err := peer_connection.data_channel.Send(encoded_fragment); err != nil {
			app.Warning(err)
			return
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
			peer_connection.incoming_data_channel = make(chan fragment, 64)
			peer_connection.outgoing_data_channel = make(chan fragment, 128)
			peer_connection.set_state(PeerConnected)
			go peer_connection.handle_outgoing_data()
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
			peer_connection.write_lock.Lock()
			close(peer_connection.outgoing_data_channel)
			peer_connection.outgoing_data_channel = nil
			peer_connection.write_lock.Unlock()
		})
	}
	peer_connection.data_channel = data_channel
}

func (peer_connection *peer_connection) handle_message_fragment(data []byte) {
	var (
		err      error
		fragment fragment
	)
	err = decode_fragment(data, &fragment)
	if err != nil {
		return
	}
	if err = validate_fragment(&fragment); err != nil {
		return
	}
	peer_connection.incoming_data_channel <- fragment
}

func (peer_connection *peer_connection) send(message_id MessageID, message []byte) (err error) {
	if len(message) > 0x3ffffff {
		err = ErrInvalidMessageSize
		return
	}

	if peer_connection.state.Load() == PeerConnected {
		peer_connection.write_lock.Lock()
		defer peer_connection.write_lock.Unlock()

		if peer_connection.outgoing_data_channel == nil {
			err = errors.New("faws/repo/p2p/peernet: cannot send to disconnected peer")
			return
		}

		message_sequence := peer_connection.message_sequence
		// overflow message sequence
		if message_sequence == 0x3ffffffffffff {
			message_sequence = 0
		}
		peer_connection.message_sequence++

		message_guid := new_message_guid(message_id, message_sequence, time.Now().UnixMilli(), uint32(len(message)))

		// break message into fragments
		var fragment_id uint16
		for fragment_data := range slices.Chunk(message, fragment_max_data_size) {
			var fragment fragment
			fragment.Message = message_guid
			fragment.Fragment = fragment_id
			fragment.Data = fragment_data
			fragment_id++
			peer_connection.outgoing_data_channel <- fragment
		}
	} else {
		err = errors.New("faws/repo/p2p/peernet: cannot send to disconnected peer")
	}
	return
}

func (peer_connection *peer_connection) close() {
	peer_connection.data_channel.Close()
	peer_connection.connection.Close()
	peer_connection.set_state(PeerDisconnected)
}
