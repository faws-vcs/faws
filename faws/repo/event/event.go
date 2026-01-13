package event

import (
	"github.com/faws-vcs/faws/faws/identity"
	"github.com/faws-vcs/faws/faws/repo/cas"
)

// A Notification signifies different types of repository events
type Notification uint8

const (
	NotifyCacheFile Notification = iota
	NotifyCacheFilePart
	NotifyCacheUsedLazySignature
	NotifyPullTag
	// ( object cas.ContentID, size int )
	NotifyPullObject
	NotifyTagQueueCount
	// ( count int )
	NotifyPullQueueCount
	// ( prefix cas.Prefix, object cas.ContentID)
	NotifyCorruptedObject
	// ( prefix cas.Prefix, object cas.ContentID)
	NotifyRemovedCorruptedObject
	NotifyBeginStage
	NotifyCompleteStage
	NotifyCheckoutFile
	NotifyCheckoutFilePart
	// p2p
	NotifyPeerConnected
	NotifyPeerDisconnected
)

// A Stage represents a phase of operations within the repository, typically one that can take quite a long time.
// BeginStage and CompleteStage can be relayed to your user interface to give the user an idea of what is happening
type Stage uint8

const (
	//
	//
	StageNone Stage = iota
	StageCacheFiles
	StageCacheFile
	StageWriteTree
	StagePullTags
	StagePullObjects
	StageCheckout
)

// NotifyParams are extra information parameters shared along with the Notification
// There is a lot of overlap, so the same fields are used.
type NotifyParams struct {
	Stage  Stage
	Prefix cas.Prefix
	// Local
	Object1 cas.ContentID
	// Remote
	Object2 cas.ContentID
	Count   int64
	// Path, Tag
	Name1 string
	// Origin
	Name2 string
	// CompleteStage
	Success bool
	// The stage is a child of the parent stage
	Child bool
	//
	ID identity.ID
}

// A NotifyFunc can be supplied to repo.Repository.Open to get notifications about the repository's actions
type NotifyFunc func(n Notification, params *NotifyParams)
