package peernet

import (
	"encoding/binary"
	"fmt"
	"strconv"

	"github.com/faws-vcs/faws/faws/repo/cas"
)

// MessageID is the prefix to any message sent from one peer to another
type MessageID uint8

const (
	// Please shut up
	Choke = iota
	// You can start talking again
	Unchoke
	// I want this object, but I don't want you to send it to me yet: (content ID)
	WantObject
	// I have this object: (content ID)
	HaveObject
	// I'm ready to receive this object from you: (content ID)
	RequestObject
	// Here's that object you wanted: (content ID) (content)
	Object
	// Text message, used for debug purposes
	Chat
	// The number of message IDs
	NumMessageID
)

type size_bound struct {
	lower uint32
	upper uint32
}

var (
	message_id_strings []string

	message_id_size_bounds []size_bound
)

func init() {
	message_id_strings = make([]string, NumMessageID)
	for i := range message_id_strings {
		message_id_strings[i] = strconv.Itoa(i)
	}
	message_id_strings[Choke] = "choke"
	message_id_strings[Unchoke] = "unchoke"
	message_id_strings[WantObject] = "want_object"
	message_id_strings[HaveObject] = "have_object"
	message_id_strings[RequestObject] = "request_object"
	message_id_strings[Object] = "object"
	message_id_strings[Chat] = "chat"

	message_id_size_bounds = make([]size_bound, NumMessageID)
	message_id_size_bounds[WantObject] = size_bound{cas.ContentIDSize, cas.ContentIDSize}
	message_id_size_bounds[HaveObject] = size_bound{cas.ContentIDSize, cas.ContentIDSize}
	message_id_size_bounds[RequestObject] = size_bound{cas.ContentIDSize, cas.ContentIDSize}
	message_id_size_bounds[Chat] = size_bound{1, 8096}
	message_id_size_bounds[Object] = size_bound{cas.ContentIDSize + cas.PrefixSize, cas.ContentIDSize + cas.PrefixSize + 16777217}
}

func (message_id MessageID) String() (s string) {
	if message_id < NumMessageID {
		s = message_id_strings[message_id]
	}
	return
}

func validate_fragment(fragment *fragment) (err error) {
	message_id := fragment.Message.MessageID()
	if message_id >= NumMessageID {
		err = fmt.Errorf("%w: %d", ErrInvalidMessageID, message_id)
		return
	}
	payload_size := fragment.Message.PayloadSize()
	bound := message_id_size_bounds[message_id]
	if payload_size < bound.lower {
		err = fmt.Errorf("%w: %s must be at least %d bytes", ErrInvalidMessageSize, message_id, bound.lower)
		return
	}
	if payload_size > bound.upper {
		err = fmt.Errorf("%w: %s must be at least %d bytes", ErrInvalidMessageSize, message_id, bound.lower)
		return
	}
	fragment_count := (payload_size + (fragment_max_data_size - 1)) / fragment_max_data_size
	if fragment.Fragment >= uint16(fragment_count) {
		err = fmt.Errorf("%w: fragment %d is not within range of expected fragments (0-%d) which is implied by a payload size of %d", ErrInvalidFragment, fragment.Fragment, fragment_count-1, payload_size)
		return
	}
	fragment_offset := fragment_max_data_size * uint32(fragment.Fragment)
	if (fragment_offset + uint32(len(fragment.Data))) > payload_size {
		err = fmt.Errorf("%w: fragment %d writes data at offset out of bounds defined by payload size", ErrInvalidFragment, fragment.Fragment)
		return
	}
	return
}

const fragment_max_data_size = 16384 - (16 + 2)

// a message is made up of one or more fragments.
// once all fragments are received, and the timestamp is within ttl
// the message will be delivered
type fragment struct {
	Message  message_guid
	Fragment uint16
	Data     []byte
}

func decode_fragment(b []byte, fragment *fragment) (err error) {
	if len(b) < 18 {
		err = ErrInvalidFragment
		return
	}

	copy(fragment.Message[:], b[:16])
	b = b[16:]

	fragment.Fragment = binary.LittleEndian.Uint16(b[:2])
	b = b[2:]

	fragment.Data = b
	return
}

func encode_fragment(fragment *fragment) (b []byte, err error) {
	b = append(b, fragment.Message[:]...)
	var fragment_id [2]byte
	binary.LittleEndian.PutUint16(fragment_id[:], fragment.Fragment)
	b = append(b, fragment_id[:]...)
	b = append(b, fragment.Data...)

	return
}
