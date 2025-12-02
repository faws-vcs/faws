package identity

import (
	"encoding/binary"
	"fmt"
	"iter"
	"os"
	"slices"
	"sort"
	"strings"

	"github.com/faws-vcs/faws/faws/fs"
	"github.com/faws-vcs/faws/faws/validate"
)

type ring_secret_entry struct {
	Primary    bool
	Pair       Pair
	Attributes Attributes
}

type ring_public_entry struct {
	ID         ID
	Attributes Attributes
}

// Ring implements a keyring of identities
type Ring struct {
	// identities that we trust
	public []ring_public_entry
	// identities that belong to us
	secret []ring_secret_entry
}

func (ring *Ring) Deabbreviate(s string) (id ID, err error) {
	if s == "" {
		err = ErrAbbreviationAmbiguous
		return
	}

	if validate.Hex(s) {
		for i := range ring.secret {
			secret_id := ring.secret[i].Pair.ID()
			if strings.HasPrefix(secret_id.String(), s) {
				if id != Nobody {
					err = ErrAbbreviationAmbiguous
					return
				}
				id = secret_id
			}
		}

		for i := range ring.public {
			public_id := ring.public[i].ID
			if strings.HasPrefix(public_id.String(), s) {
				if id != Nobody {
					err = ErrAbbreviationAmbiguous
					return
				}
				id = public_id
			}
		}

		if id != Nobody {
			return
		}
	}
	id, err = ring.Lookup(s)

	return
}

// if attributes are the most recent,
func (ring *Ring) TrustAttributes(id ID, attributes *Attributes) (err error) {
	i := sort.Search(len(ring.secret), func(i int) bool {
		list_id := ring.secret[i].Pair.ID()
		return !list_id.Less(id)
	})
	if i < len(ring.secret) && ring.secret[i].Pair.ID() == id {
		err = ErrRingCannotTrustSecretKey
		return
	}

	i = sort.Search(len(ring.public), func(i int) bool {
		list_id := ring.public[i].ID
		return !list_id.Less(id)
	})
	if i < len(ring.public) && ring.public[i].ID == id {
		current_entry := &ring.public[i]
		if current_entry.Attributes.Date < attributes.Date {
			current_entry.Attributes = *attributes
		}
		return
	}
	var new_entry ring_public_entry
	new_entry.ID = id
	new_entry.Attributes = *attributes
	ring.public = slices.Insert(ring.public, i, new_entry)
	return
}

func (ring *Ring) Lookup(nametag string) (id ID, err error) {
	if nametag == "" {
		err = ErrRingNoNametag
		return
	}
	for _, secret_entry := range ring.secret {
		if secret_entry.Attributes.Nametag == nametag {
			id = secret_entry.Pair.ID()
			return
		}
	}

	for _, public_entry := range ring.public {
		if public_entry.Attributes.Nametag == nametag {
			id = public_entry.ID
			return
		}
	}

	err = ErrRingKeyNotFound
	return
}

func (ring *Ring) GetAttributesSecret(id ID, attributes *Attributes) (err error) {
	i := sort.Search(len(ring.secret), func(i int) bool {
		list_id := ring.secret[i].Pair.ID()
		return !list_id.Less(id)
	})
	if i < len(ring.secret) && ring.secret[i].Pair.ID() == id {
		if attributes != nil {
			*attributes = ring.secret[i].Attributes
		}
		return
	}

	err = ErrRingKeyNotFound
	return
}

func (ring *Ring) SetAttributesSecret(id ID, attributes *Attributes) (err error) {
	i := sort.Search(len(ring.secret), func(i int) bool {
		list_id := ring.secret[i].Pair.ID()
		return !list_id.Less(id)
	})
	if i < len(ring.secret) && ring.secret[i].Pair.ID() == id {
		if attributes != nil {
			ring.secret[i].Attributes = *attributes
		}
		return
	}

	err = ErrRingKeyNotFound
	return
}

func (ring *Ring) GetTrustedAttributes(id ID, attributes *Attributes) (err error) {
	if err = ring.GetAttributesSecret(id, attributes); err == nil {
		return
	}
	err = nil

	i := sort.Search(len(ring.secret), func(i int) bool {
		list_id := ring.secret[i].Pair.ID()
		return !list_id.Less(id)
	})
	if i < len(ring.secret) && ring.secret[i].Pair.ID() == id {
		if attributes != nil {
			*attributes = ring.secret[i].Attributes
		}
		return
	}

	for _, public_entry := range ring.public {
		if public_entry.ID == id {
			if attributes != nil {
				*attributes = public_entry.Attributes
			}
			return
		}
	}

	err = ErrRingKeyNotFound
	return
}

func (ring *Ring) SetPrimary(id ID) (err error) {
	err = ErrRingKeyNotFound

	for i := 0; i < len(ring.secret); i++ {
		if ring.secret[i].Pair.ID() == id {
			ring.secret[i].Primary = true
			err = nil
		} else {
			ring.secret[i].Primary = false
		}
	}

	return
}

func (ring *Ring) CreateIdentity(attributes *Attributes) (new_id ID, primary bool, err error) {
	if err = validate.Nametag(attributes.Nametag); err != nil {
		return
	}
	if err = validate.Email(attributes.Email); err != nil {
		return
	}
	if err = validate.Description(attributes.Description); err != nil {
		return
	}

	if attributes.Nametag != "" {
		if current_id, name_not_found_err := ring.Lookup(attributes.Nametag); name_not_found_err == nil {
			err = fmt.Errorf("%w: %s", ErrRingNametagInUse, current_id)
			return
		}
	}

	var entry ring_secret_entry
	entry.Attributes = *attributes
	entry.Pair, err = New()
	if err != nil {
		return
	}
	new_id = entry.Pair.ID()

	if len(ring.secret) == 0 {
		primary = true
		entry.Primary = true
		ring.secret = []ring_secret_entry{entry}
		return
	}

	i := sort.Search(len(ring.secret), func(i int) bool {
		secret_entry_id := ring.secret[i].Pair.ID()
		return !secret_entry_id.Less(new_id)
	})

	ring.secret = slices.Insert(ring.secret, i, entry)

	return
}

func (ring *Ring) RemoveIdentity(id ID) (err error) {
	i := sort.Search(len(ring.secret), func(i int) bool {
		secret_entry_id := ring.secret[i].Pair.ID()
		return !secret_entry_id.Less(id)
	})
	if i < len(ring.secret) && ring.secret[i].Pair.ID() == id {
		ring.secret = slices.Delete(ring.secret, i, i+1)
		return
	}

	i = sort.Search(len(ring.public), func(i int) bool {
		public_entry_id := ring.public[i].ID
		return !public_entry_id.Less(id)
	})
	if i < len(ring.public) && ring.public[i].ID == id {
		ring.public = slices.Delete(ring.public, i, i+1)
		return
	}

	err = ErrRingKeyNotFound
	return
}

func ReadRing(filename string, ring *Ring) (err error) {
	var ring_data []byte
	ring_data, err = os.ReadFile(filename)
	if err != nil {
		return
	}

	if len(ring_data) < 8 {
		return
	}

	public_len := binary.LittleEndian.Uint32(ring_data[:4])
	ring_data = ring_data[4:]
	secret_len := binary.LittleEndian.Uint32(ring_data[:4])
	ring_data = ring_data[4:]

	if public_len > 100000 || secret_len > 100000 {
		err = ErrRingTooManyKeys
		return
	}

	ring.public = make([]ring_public_entry, public_len)
	ring.secret = make([]ring_secret_entry, secret_len)

	// TODO: validate against duplicates and incorrect sorting
	for i := uint32(0); i < public_len; i++ {
		public_entry := &ring.public[i]

		copy(public_entry.ID[:], ring_data[:32])
		ring_data = ring_data[32:]

		attributes_length := binary.LittleEndian.Uint32(ring_data[:4])
		ring_data = ring_data[4:]
		if err = UnmarshalAttributes(ring_data[:attributes_length], &public_entry.Attributes); err != nil {
			return
		}
		ring_data = ring_data[attributes_length:]
	}

	for i := uint32(0); i < secret_len; i++ {
		secret_entry := &ring.secret[i]

		secret_entry.Primary = ring_data[0] == 1
		if ring_data[0] > 1 {
			err = fmt.Errorf("faws/identity: bad Primary column in secret keyring")
			return
		}
		ring_data = ring_data[1:]

		copy(secret_entry.Pair[:], ring_data[:64])
		ring_data = ring_data[64:]

		attributes_length := binary.LittleEndian.Uint32(ring_data[:4])
		ring_data = ring_data[4:]
		if err = UnmarshalAttributes(ring_data[:attributes_length], &secret_entry.Attributes); err != nil {
			return
		}
		ring_data = ring_data[attributes_length:]
	}

	return
}

func WriteRing(filename string, ring *Ring) (err error) {
	data := make([]byte, 8)
	binary.LittleEndian.PutUint32(data[0:4], uint32(len(ring.public)))
	binary.LittleEndian.PutUint32(data[4:8], uint32(len(ring.secret)))

	for _, public_entry := range ring.public {
		data = append(data, public_entry.ID[:]...)

		var attributes_length [4]byte
		var attributes_data []byte
		attributes_data, err = MarshalAttributes(&public_entry.Attributes)
		if err != nil {
			return
		}
		binary.LittleEndian.PutUint32(attributes_length[:], uint32(len(attributes_data)))
		data = append(data, attributes_length[:]...)
		data = append(data, attributes_data...)
	}

	for _, secret_entry := range ring.secret {
		if secret_entry.Primary {
			data = append(data, 1)
		} else {
			data = append(data, 0)
		}

		data = append(data, secret_entry.Pair[:]...)

		var attributes_length [4]byte
		var attributes_data []byte
		attributes_data, err = MarshalAttributes(&secret_entry.Attributes)
		if err != nil {
			return
		}
		binary.LittleEndian.PutUint32(attributes_length[:], uint32(len(attributes_data)))
		data = append(data, attributes_length[:]...)
		data = append(data, attributes_data...)
	}

	err = os.WriteFile(filename, data, fs.DefaultPrivatePerm)
	return
}

type RingEntry struct {
	ID         ID
	Secret     bool
	Primary    bool
	Attributes *Attributes
}

func (ring *Ring) Entries() iter.Seq[*RingEntry] {
	return func(yield func(*RingEntry) bool) {
		var entry RingEntry
		entry.Secret = false
		for i := 0; i < len(ring.public); i++ {
			entry.ID = ring.public[i].ID
			entry.Attributes = &ring.public[i].Attributes
			if !yield(&entry) {
				return
			}
		}
		entry.Secret = true
		for i := 0; i < len(ring.secret); i++ {
			entry.ID = ring.secret[i].Pair.ID()
			entry.Attributes = &ring.secret[i].Attributes
			entry.Primary = ring.secret[i].Primary
			if !yield(&entry) {
				return
			}
		}
	}
}

func (ring *Ring) GetPrimaryPair(pair *Pair, attributes *Attributes) (err error) {
	for i := range ring.secret {
		secret_entry := &ring.secret[i]
		if secret_entry.Primary {
			if pair != nil {
				*pair = secret_entry.Pair
			}
			if attributes != nil {
				*attributes = secret_entry.Attributes
			}
			return
		}
	}

	err = ErrRingKeyNotFound
	return
}

func (ring *Ring) GetNametagPair(nametag string, pair *Pair, attributes *Attributes) (err error) {
	for i := range ring.secret {
		secret_entry := &ring.secret[i]
		if secret_entry.Attributes.Nametag == nametag {
			if pair != nil {
				*pair = secret_entry.Pair
			}
			if attributes != nil {
				*attributes = secret_entry.Attributes
			}
			return
		}
	}

	err = ErrRingKeyNotFound
	return
}

func (ring *Ring) GetPair(name string, pair *Pair, attributes *Attributes) (err error) {
	if validate.Hex(name) {
		var deabbreviated ID
		if deabbreviated, err = ring.Deabbreviate(name); err == nil {
			for i := range ring.secret {
				secret_entry := &ring.secret[i]
				if secret_entry.Pair.ID() == deabbreviated {
					if pair != nil {
						*pair = secret_entry.Pair
					}
					if attributes != nil {
						*attributes = secret_entry.Attributes
					}
					return
				}
			}
		}
		err = nil
	}

	err = ring.GetNametagPair(name, pair, attributes)
	return
}
