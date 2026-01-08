package tracker

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"github.com/faws-vcs/faws/faws/identity"
	"github.com/google/uuid"
)

const (
	TopicHashSize = sha256.Size
	TopicKeySize  = sha256.Size
)

type (
	TopicHash [TopicHashSize]byte
	TopicKey  [TopicKeySize]byte
)

var manifest_iv = []byte{'m', 'a', 'n', 'i', 'f', 'e', 's', 't', ' ', 's', 'e', 'c', 'r', 'e', 't', ' '}

func (topic_hash TopicHash) String() string {
	return hex.EncodeToString(topic_hash[:])
}

func (topic *Topic) Hash() (hash TopicHash) {
	h := sha256.New()
	h.Write(topic.Repository[:])
	h.Write(topic.Publisher[:])
	h.Write([]byte("public topic name"))
	h.Sum(hash[:0])
	return
}

// Key returns the symmetric encryption key for topic signals
func (topic *Topic) Key() (key TopicKey) {
	h := sha256.New()
	h.Write(topic.Repository[:])
	h.Write(topic.Publisher[:])
	h.Write([]byte("secret topic key"))
	h.Sum(key[:0])
	return
}

type Topic struct {
	// The publisher who signed the manifest.
	// Note that the publisher is not required to be the author of any commits.
	Publisher identity.ID
	// The UUID for the Faws repository.
	// This is unique and is generated once upon 'faws init' being executed
	Repository uuid.UUID
}

// returns the Topic URI
func (topic Topic) String() string {
	return "topic:" + topic.Publisher.String() + "/" + topic.Repository.String()
}

// example
//
//	topic:publisher-identity/repo-uuid
func ParseTopicURI(s string, topic *Topic) (err error) {
	var (
		has_scheme bool
	)
	s, has_scheme = strings.CutPrefix(s, "topic:")
	if !has_scheme {
		err = ErrBadTopicURI
		return
	}

	parts := strings.SplitN(s, "/", 2)
	publisher := parts[0]
	repo := parts[1]

	topic.Repository, err = uuid.Parse(repo)
	if err != nil {
		return
	}

	topic.Publisher, err = identity.Parse(publisher)
	if err != nil {
		return
	}

	return
}

func IsTopicURI(s string) (valid bool) {
	valid = strings.HasPrefix(s, "topic:")
	if !valid {
		return
	}
	var topic Topic
	err := ParseTopicURI(s, &topic)
	if err != nil {
		valid = false
		return
	}
	if topic.Publisher == identity.Nobody {
		valid = false
		return
	}
	if topic.Repository == uuid.Nil {
		valid = false
		return
	}

	return
}

func encrypt(key TopicKey, ciphertext, cleartext []byte) {
	block, err := aes.NewCipher(key[:])
	if err != nil {
		panic(err)
	}

	stream := cipher.NewCFBEncrypter(block, manifest_iv)
	stream.XORKeyStream(ciphertext, cleartext)
}

func decrypt(key TopicKey, cleartext, ciphertext []byte) {
	block, err := aes.NewCipher(key[:])
	if err != nil {
		panic(err)
	}
	stream := cipher.NewCFBDecrypter(block, manifest_iv)
	stream.XORKeyStream(cleartext, ciphertext)
}
