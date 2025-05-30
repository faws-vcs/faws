package repository

import (
	"fmt"
	"sync"

	"github.com/dustin/go-humanize"
	"github.com/faws-vcs/faws/faws/app"
	"github.com/faws-vcs/faws/faws/repo"
	"github.com/faws-vcs/faws/faws/repo/cas"
)

func prefix(p cas.Prefix) string {
	switch p {
	case cas.Commit:
		return "commit"
	case cas.Tree:
		return "tree"
	case cas.File:
		return "file"
	case cas.Part:
		return "part"
	default:
		return ""
	}
}

var (
	guard            sync.Mutex
	in_progress      bool
	objects_received int
	objects_in_queue int
	bytes_received   uint64
)

func notify(ev repo.Ev, args ...any) {
	switch ev {
	case repo.EvCacheFile:
		app.Info("caching", args[0], args[1])
	case repo.EvCacheUsedLazySignature:
		app.Info("using precached file (--lazy)", args[0], args[1])
	case repo.EvPullTag:
	case repo.EvPullObject:
		var (
			object_prefix = args[0].(cas.Prefix)
			object_hash   = args[1].(cas.ContentID)
			object_size   = args[2].(int)
		)

		guard.Lock()
		bytes_received += uint64(object_size)
		objects_received++
		var msg string
		if in_progress {
			progress := fmt.Sprintf("%d/%d", objects_received, objects_in_queue)
			msg = fmt.Sprintf("%6s %s %6s %16s objects received %s total",
				prefix(object_prefix),
				object_hash.String()[:10],
				humanize.Bytes(uint64(object_size)),
				progress,
				humanize.Bytes(bytes_received))
		} else {
			msg = fmt.Sprintf("%6s %s %6s %s total",
				prefix(object_prefix),
				object_hash.String()[:10],
				humanize.Bytes(uint64(object_size)),
				humanize.Bytes(bytes_received))
		}
		fmt.Printf("\r%s", msg)
		guard.Unlock()
	case repo.EvPullQueueCount:
		guard.Lock()
		in_progress = true
		objects_in_queue = args[0].(int)
		guard.Unlock()
	case repo.EvCorruptedObject:
		app.Warning("corrupted object", args[0].(cas.Prefix), args[1].(cas.ContentID))
	case repo.EvRemovedCorruptedObject:
		app.Warning("removed corrupted object", args[0].(cas.Prefix), args[1].(cas.ContentID))
	}
}
