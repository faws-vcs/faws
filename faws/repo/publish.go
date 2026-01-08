package repo

import (
	"time"

	"github.com/faws-vcs/faws/faws/identity"
	"github.com/faws-vcs/faws/faws/repo/p2p/tracker"
)

// Publish generates a manifest and uploads it to the tracker
func (repo *Repository) Publish(signing_identity *identity.Pair, publisher_attributes *identity.Attributes) (topic_uri string, err error) {
	var manifest_info tracker.ManifestInfo
	manifest_info.Date = time.Now().Unix()
	if publisher_attributes != nil {
		manifest_info.PublisherAttributes = *publisher_attributes
	}

	manifest_info.Tags, err = repo.Tags()
	if err != nil {
		return
	}

	// encode manifest info
	var topic tracker.Topic
	topic.Publisher = signing_identity.ID()
	topic.Repository = repo.config.UUID
	var manifest_info_data []byte
	manifest_info_data, err = tracker.EncodeManifestInfo(topic, &manifest_info)
	if err != nil {
		return
	}

	var manifest tracker.Manifest
	manifest.Info = manifest_info_data
	manifest.Publisher = signing_identity.ID()
	// sign manifest info
	identity.Sign(signing_identity, manifest.Info, &manifest.Signature)

	// encode manifest into a bundle
	var manifest_data []byte
	manifest_data, err = tracker.EncodeManifest(&manifest)
	if err != nil {
		return
	}

	var tracker_client tracker.Client
	tracker_client.Init(repo.tracker_url, nil)
	if err = tracker_client.PublishManifest(topic.Hash(), manifest_data); err != nil {
		return
	}

	topic_uri = topic.String()

	return
}
