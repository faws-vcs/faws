package event

import "github.com/faws-vcs/faws/faws/repo/cas"

type Notification uint8

const (
	NotifyCacheFile Notification = iota
	NotifyCacheFilePart
	NotifyCacheUsedLazySignature
	NotifyPullTag
	// ( object cas.ContentID, size int )
	NotifyPullObject
	// ( count int )
	NotifyPullQueueCount
	// ( prefix cas.Prefix, object cas.ContentID)
	NotifyCorruptedObject
	// ( prefix cas.Prefix, object cas.ContentID)
	NotifyRemovedCorruptedObject
	NotifyBeginStage
	NotifyCompleteStage
)

type Stage uint8

const (
	//
	//
	StageNone Stage = iota
	StagePullTags
	StagePullObjects
)

type NotifyParams struct {
	Stage  Stage
	Prefix cas.Prefix
	// Local
	Object1 cas.ContentID
	// Remote
	Object2 cas.ContentID
	Count   int
	// Path, Tag
	Name1 string
	// Origin
	Name2   string
	Success bool
}

type NotifyFunc func(n Notification, params *NotifyParams)
