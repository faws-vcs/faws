package tracker

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/faws-vcs/faws/faws/fs"
	"github.com/faws-vcs/faws/faws/identity"
	"github.com/pion/webrtc/v4"
)

type ServerConfig struct {
	Listen              string             `json:"listen"`
	ICEServers          []webrtc.ICEServer `json:"ice_servers"`
	WhitelistPublishers bool               `json:"whitelist_publishers"`
	PublisherWhitelist  []identity.ID      `json:"publisher_whitelist"`
}

type Server struct {
	directory            string
	config               ServerConfig
	topic_channels       map[TopicHash]*topic_channel
	guard_topic_channels sync.Mutex
	guard_write          sync.Mutex
}

func (server *Server) Init(directory string) (err error) {
	server.topic_channels = make(map[TopicHash]*topic_channel)

	if _, err = os.Stat(directory); err != nil {
		if err = os.Mkdir(directory, fs.DefaultPrivateDirPerm); err != nil {
			return
		}
	}
	server.directory = directory
	var (
		config_name = filepath.Join(directory, "config")
		config_data []byte
	)
	config_data, err = os.ReadFile(config_name)
	if err != nil {
		server.config.Listen = "0.0.0.0:48775"
		server.config.WhitelistPublishers = true
		server.config.ICEServers = []webrtc.ICEServer{}
		server.config.PublisherWhitelist = []identity.ID{}
		config_data, err = json.MarshalIndent(&server.config, "", "  ")
		if err != nil {
			return
		}
		err = os.WriteFile(config_name, config_data, fs.DefaultPrivatePerm)
		if err != nil {
			return
		}
	} else {
		err = json.Unmarshal(config_data, &server.config)
	}
	return
}

func (server *Server) Handler() http.Handler {
	t := http.NewServeMux()
	t.HandleFunc("GET /tracker/v1/ice_server_list", server.handle_ice_server_list)
	t.HandleFunc("PUT /tracker/v1/manifest/{topic}", server.handle_manifest_upload)
	t.HandleFunc("GET /tracker/v1/manifest/{topic}", server.handle_manifest_download)
	t.HandleFunc("/tracker/v1/signaling", server.handle_signaling)
	return t
}

func (server *Server) Serve() (err error) {
	err = http.ListenAndServe(server.config.Listen, server.Handler())
	return
}
