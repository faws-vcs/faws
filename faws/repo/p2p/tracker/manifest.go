package tracker

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/faws-vcs/faws/faws/identity"
	"github.com/faws-vcs/faws/faws/repo/cas"
	"github.com/faws-vcs/faws/faws/repo/revision"
	"github.com/faws-vcs/faws/faws/validate"
)

const (
	mime_faws_manifest     = "application/x-faws-manifest"
	min_manifest_size      = identity.IDSize + identity.SignatureSize
	min_manifest_info_size = 1 + 8 + sha256.Size
)

// Manifest stores entry points into a published/seeded repository
type Manifest struct {
	// The one who originally published the repository
	Publisher identity.ID
	// The signature proving authenticity of the manifest
	Signature identity.Signature
	// The ManifestInfo
	// You must verify this yourself before using it
	Info []byte
}

// The date is the part of the Info that doesn't need to be decrypted
// This method is used by the tracker server when determining whether to replace an existing manifest
func (m *Manifest) Time() (t time.Time) {
	if len(m.Info) < 9 {
		return
	}
	ts := int64(binary.LittleEndian.Uint64(m.Info[1:9]))
	t = time.Unix(ts, 0).UTC()
	return
}

// The signed portion of the manifest
type ManifestInfo struct {
	// Must = 1
	Reserved uint8
	// The date of publication in unix seconds
	Date int64
	// The publisher's attributes
	PublisherAttributes identity.Attributes
	// All tags and associated commit hashes are kept secret
	// Sorted by name
	Tags []revision.Tag
}

func DecodeManifest(b []byte, m *Manifest) (err error) {
	if len(b) < min_manifest_size {
		err = ErrMalformedManifest
		return
	}

	copy(m.Publisher[:], b[:identity.IDSize])
	b = b[identity.IDSize:]

	copy(m.Signature[:], b[:identity.SignatureSize])
	b = b[identity.SignatureSize:]

	m.Info = make([]byte, len(b))
	copy(m.Info, b)

	return
}

func EncodeManifest(m *Manifest) (out []byte, err error) {
	out = make([]byte, 0, min_manifest_size+len(m.Info))
	out = append(out, m.Publisher[:]...)
	out = append(out, m.Signature[:]...)
	out = append(out, m.Info[:]...)
	return
}

// b is partially ciphertext; the publisher identity and repo UUID are needed to decrypt it
func DecodeManifestInfo(b []byte, topic Topic, out *ManifestInfo) (err error) {
	if len(b) < min_manifest_info_size {
		err = ErrMalformedManifest
		return
	}

	out.Reserved = b[0]
	b = b[1:]
	if out.Reserved != 1 {
		err = ErrMalformedManifest
		return
	}

	// decode date of manifest
	out.Date = int64(binary.LittleEndian.Uint64(b[:8]))
	b = b[8:]
	ciphertext := b

	// decrypt the manifest info
	cleartext := make([]byte, len(ciphertext))
	decrypt(topic.Key(), cleartext, ciphertext)

	// verify the integrity of the manifest info (i.e. have we properly decrypted this?)
	purported_checksum := cleartext[:sha256.Size]
	cleartext = cleartext[sha256.Size:]
	h := sha256.New()
	h.Write(cleartext)
	actual_checksum := h.Sum(nil)
	if !bytes.Equal(purported_checksum, actual_checksum) {
		err = ErrMalformedManifest
		return
	}

	// read the publisher's attributes
	attributes_size := binary.LittleEndian.Uint32(cleartext)
	cleartext = cleartext[4:]
	if attributes_size > 9999 || len(cleartext) < int(attributes_size) {
		err = ErrMalformedManifest
		return
	}
	attributes_data := cleartext[:attributes_size]
	cleartext = cleartext[attributes_size:]
	if err = identity.UnmarshalAttributes(attributes_data, &out.PublisherAttributes); err != nil {
		err = fmt.Errorf("%w: %s", err, spew.Sdump(attributes_data))
		return
	}

	// start to read the tags
	num_tags := int(binary.LittleEndian.Uint32(cleartext))
	cleartext = cleartext[4:]
	// refuse to process absurdly large manifests
	if num_tags > 10000000 {
		err = ErrMalformedManifest
		return
	}
	if len(cleartext) < num_tags*cas.ContentIDSize {
		err = ErrMalformedManifest
		return
	}
	// read the hashes for each tag
	tag_hashes := make([]cas.ContentID, num_tags)
	for i := range num_tags {
		copy(tag_hashes[i][:], cleartext[:cas.ContentIDSize])
		cleartext = cleartext[cas.ContentIDSize:]
	}
	// read the lengths of each tag name
	tag_name_total_length := uint32(0)
	if len(cleartext) < num_tags*4 {
		err = ErrMalformedManifest
		return
	}
	tag_name_lengths := make([]uint32, num_tags)
	for i := range num_tags {
		tag_name_length := binary.LittleEndian.Uint32(cleartext)
		tag_name_lengths[i] = tag_name_length
		cleartext = cleartext[4:]
		tag_name_total_length += tag_name_length
	}
	// finally, read the tag names
	if int(tag_name_total_length) > len(cleartext) {
		err = ErrMalformedManifest
		return
	}
	tag_names := make([]string, num_tags)
	for i := range num_tags {
		tag_name := string(cleartext[:tag_name_lengths[i]])
		if err = validate.CommitTag(tag_name); err != nil {
			err = ErrMalformedManifest
			return
		}
		tag_names[i] = tag_name
		cleartext = cleartext[tag_name_lengths[i]:]
	}
	// tie hashes and tags together
	out.Tags = make([]revision.Tag, num_tags)
	for i := range num_tags {
		out.Tags[i].Name = tag_names[i]
		out.Tags[i].CommitHash = tag_hashes[i]
	}
	return
}

func EncodeManifestInfo(topic Topic, info *ManifestInfo) (b []byte, err error) {
	var attributes_data []byte
	attributes_data, err = identity.MarshalAttributes(&info.PublisherAttributes)
	if err != nil {
		return
	}

	cleartext := make([]byte, sha256.Size)

	var attributes_size [4]byte
	binary.LittleEndian.PutUint32(attributes_size[:], uint32(len(attributes_data)))
	cleartext = append(cleartext, attributes_size[:]...)

	cleartext = append(cleartext, attributes_data...)

	var tag_count [4]byte
	binary.LittleEndian.PutUint32(tag_count[:], uint32(len(info.Tags)))
	cleartext = append(cleartext, tag_count[:]...)

	for i := range info.Tags {
		cleartext = append(cleartext, info.Tags[i].CommitHash[:]...)
	}
	for i := range info.Tags {
		var name_length [4]byte
		binary.LittleEndian.PutUint32(name_length[:], uint32(len(info.Tags[i].Name)))
		cleartext = append(cleartext, name_length[:]...)
	}
	for i := range info.Tags {
		if err = validate.CommitTag(info.Tags[i].Name); err != nil {
			err = ErrMalformedManifest
			return
		}
		cleartext = append(cleartext, []byte(info.Tags[i].Name)...)
	}

	h := sha256.New()
	h.Write(cleartext[sha256.Size:])
	copy(cleartext[:sha256.Size], h.Sum(nil))

	ciphertext := make([]byte, len(cleartext))
	encrypt(topic.Key(), ciphertext, cleartext)

	b = make([]byte, 9)
	b[0] = 1 // Reserved
	binary.LittleEndian.PutUint64(b[1:], uint64(info.Date))
	b = append(b, ciphertext...)

	return
}
