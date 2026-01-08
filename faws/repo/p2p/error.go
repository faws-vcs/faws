package p2p

import "fmt"

var (
	ErrEvilServer        = fmt.Errorf("faws/repo/p2p: the tracker server is malicious or poorly implemented")
	ErrNotSubscribed     = fmt.Errorf("faws/repo/p2p: agent is not subscribed to this topic")
	ErrAlreadySubscribed = fmt.Errorf("faws/repo/p2p: agent is already subscribed to this topic")
)
