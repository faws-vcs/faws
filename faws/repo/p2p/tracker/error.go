package tracker

import "fmt"

var (
	ErrMalformedManifest = fmt.Errorf("faws/p2p/tracker: manifest is badly formed")
	ErrBadTopicName      = fmt.Errorf("faws/p2p/tracker: topic name is not of the correct format")
	ErrBadTopicMessage   = fmt.Errorf("faws/p2p/tracker: signaling message is badly formed")
	ErrBadTopicURI       = fmt.Errorf("faws/p2p/tracker: topic URI is badly formed")

	ErrBadLogin         = fmt.Errorf("faws/p2p/tracker: bad protocol in login stage")
	ErrBadCommand       = fmt.Errorf("faws/p2p/tracker: a bad command was received")
	ErrClientIsShutdown = fmt.Errorf("faws/p2p/tracker: client is shutdown")
)
