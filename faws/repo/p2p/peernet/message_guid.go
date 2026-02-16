package peernet

import (
	"encoding/binary"
	"fmt"
	"time"
)

// 128 bits for identifying a message transferred through peernet
//
//		000-004  message id   4 bits,  protocol message type
//		004-054  sequence     50 bits  increasing number
//		054-102  timestamp    48 bits, unix milliseconds
//	    102-128  payload size 26 bits, the size of the completed message payload
type MessageGUID [16]byte

func (message_guid MessageGUID) lo_word() (w uint64) {
	return binary.BigEndian.Uint64(message_guid[8:16])
}

func (message_guid MessageGUID) hi_word() (w uint64) {
	return binary.BigEndian.Uint64(message_guid[0:8])
}

func (message_guid *MessageGUID) set_lo_word(w uint64) {
	binary.BigEndian.PutUint64(message_guid[8:], w)
}

func (message_guid *MessageGUID) set_hi_word(w uint64) {
	binary.BigEndian.PutUint64(message_guid[:8], w)
}

func (message_guid MessageGUID) MessageID() (message_id MessageID) {
	//	message_id = message_guid[0] & 0xf
	message_id = MessageID((message_guid.hi_word() >> 60) & 0xf)
	return
}

func (message_guid MessageGUID) Sequence() (sequence uint64) {
	// 0xFFFFFFFFFFFFFFF
	// &  0xfffffffffffffff : remove hi 4 (message id) bits
	// >> 10                : shift out timestamp bits
	sequence = ((message_guid.hi_word() & 0xfffffffffffffff) >> 10)
	return
}

func (message_guid MessageGUID) get_timestamp() (timestamp int64) {
	// hi 10 bits are in the hi word
	timestamp_hi := message_guid.hi_word() & 0x3ff
	timestamp_lo := message_guid.lo_word() >> 26
	timestamp = int64((timestamp_hi << 38) | timestamp_lo)
	return
}

func (message_guid MessageGUID) Time() (t time.Time) {
	t = time.UnixMilli(message_guid.get_timestamp())
	return
}

func (message_guid MessageGUID) PayloadSize() (payload_size uint32) {
	payload_size = uint32(message_guid.lo_word()) & 0x3ffffff
	return
}

func (message_guid MessageGUID) String() string {
	return fmt.Sprintf(
		"%013x-%012x-%07x-%s",
		message_guid.Sequence(),
		message_guid.get_timestamp(),
		message_guid.PayloadSize(),
		message_guid.MessageID(),
	)
}

func NewMessageGUID(message_id MessageID, sequence uint64, t time.Time, payload_size uint32) (message_guid MessageGUID) {
	timestamp := t.UnixMilli()

	var lo, hi uint64
	hi |= ((uint64(message_id) & 0xf) << 60)
	hi |= (sequence & 0x3ffffffffffff) << 10

	timestamp_hi := (uint64(timestamp) & 0xffc000000000) >> 38
	timestamp_lo := (uint64(timestamp) & 0x3fffffffff) << 26

	hi |= timestamp_hi
	lo |= timestamp_lo

	lo |= uint64(payload_size) & 0x3ffffff

	message_guid.set_hi_word(hi)
	message_guid.set_lo_word(lo)
	return
}
