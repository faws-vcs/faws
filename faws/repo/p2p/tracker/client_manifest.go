package tracker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/faws-vcs/faws/faws/repo/p2p/tracker/models"
)

func (client *Client) PublishManifest(topic_hash TopicHash, manifest []byte) (err error) {
	var (
		status int
		reply  io.ReadCloser
	)
	status, reply, err = client.put("/tracker/v1/manifest/"+topic_hash.String(), bytes.NewReader(manifest))
	if err != nil {
		return
	}
	defer reply.Close()
	if status != http.StatusOK {
		var generic_response models.GenericResponse
		if err = json.NewDecoder(reply).Decode(&generic_response); err != nil {
			err = fmt.Errorf("faws/p2p/tracker: tracker returned status (%d) and did not give a reason", status)
			return
		}

		err = fmt.Errorf("faws/p2p/tracker: tracker returned status (%d) with the message '%s'", status, generic_response.Message)
		return
	}
	return
}

func (client *Client) FetchManifest(topic_name string) (manifest []byte, err error) {
	var (
		status       int
		reply        io.ReadCloser
		content_type string
	)
	status, content_type, reply, err = client.get("/tracker/v1/manifest/" + topic_name)
	if err != nil {
		return
	}
	defer reply.Close()
	if status != http.StatusOK {
		var generic_response models.GenericResponse
		if err = json.NewDecoder(reply).Decode(&generic_response); err != nil {
			err = fmt.Errorf("faws/p2p/tracker: tracker returned status (%d) and did not give a reason", status)
			return
		}

		err = fmt.Errorf("faws/p2p/tracker: tracker returned status (%d) with the message '%s'", status, generic_response.Message)
		return
	}
	if content_type != mime_faws_manifest {
		err = fmt.Errorf("faws/p2p/tracker: tracker sent you an object that is not a manifest")
		return
	}
	manifest, err = io.ReadAll(reply)
	return
}
