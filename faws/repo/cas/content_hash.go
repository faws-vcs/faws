package cas

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
)

const ContentIDSize = 20

// truncated SHA-256 hash of the file/file segment
type ContentID [ContentIDSize]byte

var Nil ContentID

func (id ContentID) String() string {
	return hex.EncodeToString(id[:])
}

func (id ContentID) Less(than ContentID) bool {
	return bytes.Compare(id[:], than[:]) == -1
}

func hash_content(prefix Prefix, data []byte) (id ContentID) {
	hash := sha256.New()
	hash.Write(prefix[:])
	hash.Write(data[:])
	checksum := hash.Sum(nil)
	copy(id[:], checksum[:ContentIDSize])
	return
}
