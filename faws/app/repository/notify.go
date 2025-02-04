package repository

import (
	"fmt"

	"github.com/dustin/go-humanize"
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
	in_progress      bool
	objects_received int
	objects_in_queue int
	bytes_received   uint64
)

func notify(ev repo.Ev, args ...any) {
	switch ev {
	case repo.EvPullTag:
	case repo.EvPullObject:
		var (
			object_prefix = args[0].(cas.Prefix)
			object_hash   = args[1].(cas.ContentID)
			object_size   = args[2].(int)
		)

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
	case repo.EvPullQueueCount:
		in_progress = true
		objects_in_queue = args[0].(int)
	}
}
