package tracker

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/faws-vcs/faws/faws/repo/p2p/tracker/models"
)

func (client *Client) GetICEServers() (list models.ICEServerList, err error) {
	var (
		status       int
		content      io.ReadCloser
		content_type string
	)
	status, content_type, content, err = client.get("/tracker/v1/ice_server_list")
	if err != nil {
		return
	}
	if status != http.StatusOK {
		err = fmt.Errorf("faws/p2p/tracker: server returned %d", status)
		return
	}
	if content_type != "application/json; charset=utf-8" {
		err = fmt.Errorf("faws/p2p/tracker: invalid ice server format")
		return
	}
	err = json.NewDecoder(content).Decode(&list)
	content.Close()
	return
}
