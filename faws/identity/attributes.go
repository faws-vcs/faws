package identity

import (
	"encoding/binary"
	"fmt"
)

const (
	max_nametag_length     = 63
	max_description_length = 1024
	max_email_length       = 320
)

// Attributes for a particular ID
type Attributes struct {
	Reserved uint8 `json:"reserved"`
	// Max-63-byte human-readable text ID in lowercase
	// e.g. "joe.schmoe"
	Nametag string `json:"nametag,omitempty"`
	// Long arbitrary description. Maximum 1024 bytes
	Description string `json:"description,omitempty"`
	// Email address
	Email string `json:"email,omitempty"`
	// Unix seconds for when these attributes were changed
	// When the user is trusted, newer attributes are automatically read from repositories
	Date int64 `json:"date"`
}

// Serialize Attributes
func MarshalAttributes(a *Attributes) (data []byte, err error) {
	nametag_length := len(a.Nametag)
	description_length := len(a.Description)
	email_length := len(a.Email)
	if nametag_length > max_nametag_length {
		err = fmt.Errorf("faws/identity: nametag attribute is too long %d/%d", nametag_length, max_nametag_length)
		return
	}
	if description_length > max_description_length {
		err = fmt.Errorf("faws/identity: description attribute is too long %d/%d", description_length, max_description_length)
		return
	}
	if email_length > max_email_length {
		err = fmt.Errorf("faws/identity: email attribute is too long %d/%d", email_length, max_email_length)
		return
	}
	var (
		nametag_length_data     [2]byte
		description_length_data [2]byte
		email_length_data       [2]byte
		date_data               [8]byte
	)
	data = append(data, a.Reserved)
	binary.LittleEndian.PutUint16(nametag_length_data[:], uint16(nametag_length))
	binary.LittleEndian.PutUint16(description_length_data[:], uint16(description_length))
	binary.LittleEndian.PutUint16(email_length_data[:], uint16(email_length))
	binary.LittleEndian.PutUint64(date_data[:], uint64(a.Date))
	data = append(data, nametag_length_data[:]...)
	data = append(data, description_length_data[:]...)
	data = append(data, email_length_data[:]...)
	data = append(data, date_data[:]...)
	data = append(data, []byte(a.Nametag)...)
	data = append(data, []byte(a.Description)...)
	data = append(data, []byte(a.Email)...)
	return
}

func UnmarshalAttributes(data []byte, a *Attributes) (err error) {
	if len(data) < 15 {
		err = fmt.Errorf("faws/identity: bad attributes format: %d", len(data))
		return
	}
	a.Reserved = data[0]
	data = data[1:]
	nametag_length := binary.LittleEndian.Uint16(data[:2])
	data = data[2:]
	description_length := binary.LittleEndian.Uint16(data[:2])
	data = data[2:]
	email_length := binary.LittleEndian.Uint16(data[:2])
	data = data[2:]
	a.Date = int64(binary.LittleEndian.Uint64(data[:8]))
	data = data[8:]

	if nametag_length > max_nametag_length {
		err = fmt.Errorf("faws/identity: nametag attribute is too long %d/%d", nametag_length, max_nametag_length)
		return
	}
	if description_length > max_description_length {
		err = fmt.Errorf("faws/identity: description attribute is too long %d/%d", description_length, max_description_length)
		return
	}
	if email_length > max_email_length {
		err = fmt.Errorf("faws/identity: email attribute is too long %d/%d", email_length, max_email_length)
		return
	}
	a.Nametag = string(data[:nametag_length])
	data = data[nametag_length:]
	a.Description = string(data[:description_length])
	data = data[description_length:]
	a.Email = string(data[:email_length])
	return
}
