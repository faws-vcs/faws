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

// ID is a public cryptographic identity.
// It can be used to verify the authenticity of commits and manifests
type ID [IDSize]byte

// Pair is a secret counterpart for an ID. This should not be shared with anybody.
type Pair [PairSize]byte

var Nobody ID

var Nil Pair

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

func (id *ID) UnmarshalText(text []byte) (err error) {
	var n int
	n, err = hex.Decode(id[:], text)
	if err != nil {
		return
	}
	if n != IDSize {
		err = ErrIDStringTooShort
	}
	return
}

func Parse(s string) (id ID, err error) {
	err = id.UnmarshalText([]byte(s))
	return
}
