package peernet

import "fmt"

var (
	ErrPeerNotFound       = fmt.Errorf("faws/repo/p2p/peernet: peer not found")
	ErrInvalidMessageID   = fmt.Errorf("faws/repo/p2p/peernet: message ID is not valid")
	ErrInvalidMessageSize = fmt.Errorf("faws/repo/p2p/peernet: message has an invalid size for its ID")
	ErrInvalidFragment    = fmt.Errorf("faws/repo/p2p/peernet: message fragment is badly formed")
)
