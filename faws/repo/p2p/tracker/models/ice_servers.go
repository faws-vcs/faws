package models

import "github.com/pion/webrtc/v4"

type ICEServerList struct {
	ICEServers []webrtc.ICEServer `json:"ice_servers"`
}
