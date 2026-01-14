package peernet

import (
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

func validate_message_header(message_size uint32, message_id MessageID) (err error) {
	if message_id >= NumMessageID {
		err = fmt.Errorf("%w: %d", ErrInvalidMessageID, message_id)
		return
	}
	bound := message_id_size_bounds[message_id]
	if message_size < bound.lower {
		err = fmt.Errorf("%w: %s must be at least %d bytes", ErrInvalidMessageSize, message_id, bound.lower)
		return
	}
	if message_size > bound.upper {
		err = fmt.Errorf("%w: %s must be at least %d bytes", ErrInvalidMessageSize, message_id, bound.lower)
		return
	}
	return
}
