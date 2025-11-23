package repository

import (
	"fmt"
	"sync"

	"github.com/dustin/go-humanize"
	"github.com/faws-vcs/faws/faws/app"
	"github.com/faws-vcs/faws/faws/repo/cas"
	"github.com/faws-vcs/faws/faws/repo/event"
	"github.com/pterm/pterm"
)

var (
	stages_text = map[event.Stage]string{
		event.StagePullObjects: "pulling objects...",
		event.StagePullTags:    "pulling tags...",
	}
)

type activity_screen struct {
	stage_spinner    *pterm.SpinnerPrinter
	in_progress      bool
	objects_received int
	objects_in_queue int
	bytes_received   uint64
}

var scrn activity_screen

func begin_stage(stage event.Stage) {
	var err error
	scrn.stage_spinner, err = pterm.DefaultSpinner.Start(stages_text[stage])
	if err != nil {
		panic(err)
	}
}

func complete_stage(stage event.Stage, success bool) {
	// scrn.stage_spinner.Success()
	scrn.stage_spinner.Stop()
	scrn.stage_spinner = nil
}

func update_pull_info(object_size int, object_prefix cas.Prefix, object_hash cas.ContentID) {
	scrn.bytes_received += uint64(object_size)
	scrn.objects_received++

	// #/#
	progress_text := fmt.Sprintf("%d/%d", scrn.objects_received, scrn.objects_in_queue)
	prefix_text := prefix(object_prefix)
	object_hash_abbreviation := object_hash.String()[:10]
	object_size_human := humanize.Bytes(uint64(object_size))
	bytes_received_human := humanize.Bytes(scrn.bytes_received)

	if scrn.in_progress {
		pterm.DefaultArea.Start()

		progress := 
		msg = fmt.Sprintf("%6s %s %6s %16s objects received %s total",
			prefix(object_prefix),
			object_hash.String()[:10],
			humanize.Bytes(uint64(object_size)),
			progress,
			humanize.Bytes(scrn.bytes_received))
	} else {
		msg = fmt.Sprintf("%6s %s %6s %s total",
			prefix(object_prefix),
			object_hash.String()[:10],
			humanize.Bytes(uint64(object_size)),
			humanize.Bytes(scrn.bytes_received))
	}
	// app.Info(fmt.Sprintf("\r%s", msg))
}

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
	guard sync.Mutex
)

func notify(ev event.Notification, params *event.NotifyParams) {
	guard.Lock()

	switch ev {
	case event.NotifyCacheFile:
		app.Info("caching", params.Name1, params.Name2)
	case event.NotifyCacheUsedLazySignature:
		app.Info("using precached file (--lazy)", params.Name1, params.Name2)
	case event.NotifyPullTag:
		app.Info("pulled tag:", params.Name1)
	case event.NotifyPullObject:
		var (
			object_prefix = params.Prefix
			object_hash   = params.Object1
			object_size   = params.Count
		)

		update_pull_info(object_size, object_prefix, object_hash)
	case event.NotifyPullQueueCount:
		scrn.in_progress = true
		scrn.objects_in_queue = params.Count
	case event.NotifyCorruptedObject:
		app.Warning("corrupted object", params.Prefix, params.Object1)
	case event.NotifyRemovedCorruptedObject:
		app.Warning("removed corrupted object", params.Prefix, params.Object1)
	case event.NotifyBeginStage:
		begin_stage(params.Stage)
	case event.NotifyCompleteStage:
		complete_stage(params.Stage, params.Success)
	}
	guard.Unlock()
}
