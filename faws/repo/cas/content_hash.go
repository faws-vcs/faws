package cas

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
)

// MaxObjectSize objects may not exceed this size constraint.
const MaxObjectSize = 0x1000000

const ContentIDSize = 20

// ContentID is a truncated SHA-256 hash of content
// This is used widely across Faws to identify objects. It may also be referred to as an "object hash"
type ContentID [ContentIDSize]byte

// Nil is the zero value of a ContentID. Its presence symbolizes a lack of content. For instance, if a commit is initial in nature, its parent ContentID is Nil.
var Nil ContentID

// String returns the hexadecimal representation of the ContentID
func (id ContentID) String() string {
	if id == Nil {
		return "nil"
	}
	return hex.EncodeToString(id[:])
}

// Less returns true if the ID is less than the method argument "than"
// This is used to sort a list of ContentIDs.
func (id ContentID) Less(than ContentID) bool {
	return bytes.Compare(id[:], than[:]) == -1
}

// hash_content is used to generate ContentIDs
func hash_content(prefix Prefix, data []byte) (id ContentID) {
	hash := sha256.New()
	hash.Write(prefix[:])
	hash.Write(data[:])
	checksum := hash.Sum(nil)
	copy(id[:], checksum[:ContentIDSize])
	return
}
