package identity

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
)

const (
	IDSize   = ed25519.PublicKeySize
	PairSize = ed25519.PrivateKeySize
)

type ID [IDSize]byte

type Pair [PairSize]byte

var Nobody ID

func (pair Pair) ID() (id ID) {
	copy(id[:], pair[32:])
	return id
}

func New() (pair Pair, err error) {
	var new_key ed25519.PrivateKey
	_, new_key, err = ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return
	}
	copy(pair[:], new_key)
	return
}

func (id ID) String() string {
	return hex.EncodeToString(id[:])
}

func (id ID) Less(than ID) bool {
	return bytes.Compare(id[:], than[:]) == -1
}
