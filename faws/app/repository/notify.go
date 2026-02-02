package repository

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/faws-vcs/console"
	"github.com/faws-vcs/faws/faws/app"
	"github.com/faws-vcs/faws/faws/repo/cas"
	"github.com/faws-vcs/faws/faws/repo/event"
	"github.com/faws-vcs/faws/faws/repo/p2p/peernet"
)

const (
	summarize_pruning = 1 << iota
)

var (
	stages_text = map[event.Stage]string{
		event.StagePullObjects:  "Retrieve objects",
		event.StagePullTags:     "Retrieve tags",
		event.StageCacheFiles:   "Cache files",
		event.StageCacheFile:    "Cache file",
		event.StageWriteTree:    "Write tree",
		event.StageCheckout:     "Checkout",
		event.StageServeObjects: "Distribute objects",
		event.StageVisitObjects: "Visit objects",
		event.StagePackObjects:  "Pack objects",
	}

	scrn activity_screen
)

type activity_stage struct {
	stage    event.Stage
	state    int
	is_child bool
}

type activity_screen struct {
	guard sync.RWMutex

	summary_mode uint8

	stages           []activity_stage
	in_progress      bool
	tags_received    int
	tags_in_queue    int
	objects_received int
	objects_in_queue int
	bytes_received   uint64

	last_object_prefix cas.Prefix
	last_object_hash   cas.ContentID
	last_object_size   uint64

	current_file_size     int64
	current_file_progress int64
	current_file_origin   string
	current_file_name     string

	connected_peers int

	received_messages              int64
	last_message_id                peernet.MessageID
	object_uploads                 int64
	duplicate_object_downloads     int64
	duplicate_object_download_size uint64

	objects_visited        uint64
	objects_in_visit_queue uint64
	objects_pruned         uint64

	verbose bool
}

func begin_stage(stage event.Stage, child bool) {
	scrn.guard.Lock()
	scrn.stages = append(scrn.stages, activity_stage{stage, 0, child})
	scrn.guard.Unlock()
}

func complete_stage(stage event.Stage, success bool) {
	scrn.guard.Lock()
	if stage != scrn.stages[len(scrn.stages)-1].stage {
		panic("stage complete mismatch")
	}
	if success {
		if scrn.stages[len(scrn.stages)-1].is_child {
			scrn.stages = scrn.stages[:len(scrn.stages)-1]
		} else {
			scrn.stages[len(scrn.stages)-1].state = 1
		}
	} else {
		scrn.stages[len(scrn.stages)-1].state = 2
	}
	scrn.guard.Unlock()
}

func update_pull_info(object_size int, object_prefix cas.Prefix, object_hash cas.ContentID) {
	scrn.guard.Lock()
	scrn.bytes_received += uint64(object_size)
	scrn.objects_received++

	scrn.last_object_prefix = object_prefix
	scrn.last_object_hash = object_hash
	scrn.last_object_size = uint64(object_size)
	scrn.guard.Unlock()
}

func update_checkout(destination string, size int64) {
	scrn.guard.Lock()
	scrn.current_file_origin = destination
	scrn.current_file_size = size
	scrn.current_file_progress = 0
	scrn.guard.Unlock()
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

func notify(ev event.Notification, params *event.NotifyParams) {

	switch ev {
	case event.NotifyCacheFile:

		if scrn.verbose {
			app.Info("caching", params.Name1, params.Name2)
		}

		// don't try to call app.Info while modifying the activity screen state : it will lead to DEADLOCK
		// as the hud is already trying to get a lock when an update is being notified, these can never become unlocked
		scrn.guard.Lock()
		scrn.current_file_size = params.Count
		scrn.current_file_progress = 0
		scrn.current_file_origin = params.Name2
		scrn.guard.Unlock()

	case event.NotifyCacheFilePart:

		scrn.guard.Lock()
		scrn.current_file_progress += params.Count
		scrn.guard.Unlock()

	case event.NotifyCacheUsedLazySignature:
		app.Info("using precached file (--lazy)", params.Name1, params.Name2)
	case event.NotifyIndexRemoveFile:
		app.Info(fmt.Sprintf("rm '%s'", params.Name1))
	case event.NotifyPullTag:

		local_hash := params.Object1
		remote_hash := params.Object2
		if local_hash == cas.Nil {
			app.Info("retrieved tag", params.Name1+":", params.Object2)
		} else if local_hash != remote_hash {
			app.Info("updated tag", params.Name1+":", params.Object1, "=>", params.Object2)
		} else if scrn.verbose {
			app.Info("tag", params.Name1+":", params.Object2)
		}

		scrn.guard.Lock()
		scrn.tags_received++
		scrn.guard.Unlock()
	case event.NotifyPullObject:
		var (
			object_prefix = params.Prefix
			object_hash   = params.Object1
			object_size   = params.Count
		)

		update_pull_info(int(object_size), object_prefix, object_hash)
	case event.NotifyTagQueueCount:
		scrn.guard.Lock()
		scrn.tags_in_queue = int(params.Count)
		scrn.guard.Unlock()
	case event.NotifyPullQueueCount:
		scrn.guard.Lock()
		scrn.in_progress = true
		scrn.objects_in_queue = int(params.Count)
		scrn.guard.Unlock()
	case event.NotifyCorruptedObject:
		app.Warning("corrupted object", params.Prefix, params.Object1)
	case event.NotifyRemovedCorruptedObject:
		app.Warning("removed corrupted object", params.Prefix, params.Object1)
	case event.NotifyPruneObject:
		scrn.guard.Lock()
		scrn.objects_pruned++
		scrn.guard.Unlock()
		app.Warning("pruned unreachable object", params.Object1)
	case event.NotifyBeginStage:
		begin_stage(params.Stage, params.Child)
	case event.NotifyCompleteStage:
		complete_stage(params.Stage, params.Success)
	case event.NotifyCheckoutFile:
		if scrn.verbose {
			app.Info(params.Name1)
		}
		update_checkout(params.Name1, params.Count)
	case event.NotifyCheckoutFilePart:
		scrn.guard.Lock()
		scrn.current_file_progress += params.Count
		scrn.guard.Unlock()
	case event.NotifyPeerConnected:
		scrn.guard.Lock()
		scrn.connected_peers++
		// app.Info("connected to peer", params.ID)
		scrn.guard.Unlock()
	case event.NotifyPeerDisconnected:
		// app.Info("disconnected from", params.ID)
		scrn.guard.Lock()
		scrn.connected_peers--
		scrn.guard.Unlock()
	case event.NotifyPeerNetMessage:
		scrn.guard.Lock()
		scrn.received_messages++
		scrn.last_message_id = params.MessageID
		scrn.guard.Unlock()
	case event.NotifyPeerObjectUpload:
		scrn.guard.Lock()
		scrn.object_uploads++
		scrn.guard.Unlock()
	case event.NotifyPeerObjectDuplicateDownload:
		scrn.guard.Lock()
		scrn.duplicate_object_downloads++
		scrn.duplicate_object_download_size += uint64(params.Count)
		scrn.guard.Unlock()
	case event.NotifyVisitObject:
		scrn.guard.Lock()
		scrn.objects_visited++
		scrn.guard.Unlock()
	case event.NotifyVisitQueueCount:
		scrn.guard.Lock()
		scrn.objects_in_visit_queue = uint64(params.Count)
		scrn.guard.Unlock()
	}

	console.SwapHud()
}

func render_activity_screen(hud *console.Hud) {
	scrn.guard.RLock()
	defer scrn.guard.RUnlock()

	if hud.Exiting() {
		if scrn.summary_mode&summarize_pruning != 0 {
			prune_summary(hud)
		}
		return
	}

	var spinner console.Spinner
	spinner.Stylesheet.Sequence[7] = console.Cell{'⡿', console.BrightBlue, 0}
	spinner.Stylesheet.Sequence[6] = console.Cell{'⣟', console.BrightBlue, 0}
	spinner.Stylesheet.Sequence[5] = console.Cell{'⣯', console.BrightBlue, 0}
	spinner.Stylesheet.Sequence[4] = console.Cell{'⣷', console.BrightBlue, 0}
	spinner.Stylesheet.Sequence[3] = console.Cell{'⣾', console.BrightBlue, 0}
	spinner.Stylesheet.Sequence[2] = console.Cell{'⣽', console.BrightBlue, 0}
	spinner.Stylesheet.Sequence[1] = console.Cell{'⣻', console.BrightBlue, 0}
	spinner.Stylesheet.Sequence[0] = console.Cell{'⢿', console.BrightBlue, 0}
	spinner.Frequency = time.Second / 3

	var stage_stack []int
	var stage *activity_stage

	for i := range scrn.stages {
		stage = &scrn.stages[i]
		if len(stage_stack) == 0 || stage.is_child {
			stage_stack = append(stage_stack, 0)
		}
		stage_stack[len(stage_stack)-1]++
		var stage_str = ""
		for i, v := range stage_stack {
			if i > 0 {
				stage_str += "."
			}
			stage_str += strconv.Itoa(v)
		}

		message := fmt.Sprintf("%s) %s ", stage_str, stages_text[stage.stage])
		var stage_text console.Text
		switch stage.state {
		case 0:
			stage_text.Stylesheet.Width = len(message)
			stage_text.Add(message, console.BrightBlue, 0)
			hud.Line(&stage_text, &spinner)
		case 1:
			stage_text.Stylesheet.Width = len(message) + 1
			stage_text.Add(message, 0, 0)
			stage_text.Add("🗸", console.BrightGreen, 0)
			hud.Line(&stage_text)
		case 2:
			stage_text.Stylesheet.Width = len(message) + 1
			stage_text.Add(message, console.Red, 0)
			stage_text.Add("❌", 0, 0)
			hud.Line(&stage_text)
		}
	}

	if stage == nil {
		return
	}

	// if scrn.received_messages > 0 {
	// 	var received_messages_text console.Text
	// 	received_messages_text.Stylesheet.Width = console.Width()
	// 	received_messages_text.Add(fmt.Sprintf("%d messages received. last message id: %s", scrn.received_messages, scrn.last_message_id), 0, 0)
	// 	hud.Line(&received_messages_text)
	// }

	if scrn.connected_peers > 0 {
		var peers_text console.Text
		peers_text.Stylesheet.Margin[console.Left] = 1
		peers_text.Stylesheet.Width = console.Width()
		peers_text.Add(fmt.Sprintf("%d peers connected", scrn.connected_peers), 0, 0)
		hud.Line(&peers_text)
	}

	var progress_bar console.ProgressBar
	progress_bar.Stylesheet.Sequence[console.PbCaseLeft] = console.Cell{'[', 0, 0}
	progress_bar.Stylesheet.Sequence[console.PbCaseRight] = console.Cell{']', 0, 0}
	progress_bar.Stylesheet.Sequence[console.PbFluid] = console.Cell{'#', 0, 0}
	progress_bar.Stylesheet.Sequence[console.PbVoid] = console.Cell{'.', 0, 0}
	progress_bar.Stylesheet.Sequence[console.PbTail] = console.Cell{'#', 0, 0}
	progress_bar.Stylesheet.Sequence[console.PbHead] = console.Cell{'#', 0, 0}

	switch stage.stage {
	case event.StageVisitObjects:
		if scrn.objects_in_visit_queue > 0 {
			var progress_text console.Text
			progress_text.Stylesheet.Width = console.Width()
			progress_text.Add(fmt.Sprintf("%d/%d objects visited", scrn.objects_visited, scrn.objects_in_visit_queue), 0, 0)
			hud.Line(&progress_text)

			progress_bar.Stylesheet.Width = console.Width()
			progress_bar.Progress = float64(scrn.objects_visited) / float64(scrn.objects_in_visit_queue)
			hud.Line(&progress_bar)
		}
	case event.StagePullTags:
		var progress_text console.Text
		progress_text.Stylesheet.Width = console.Width()
		progress_text.Add(fmt.Sprintf("%d/%d tags received", scrn.tags_received, scrn.tags_in_queue), 0, 0)
		hud.Line(&progress_text)

		progress_bar.Stylesheet.Width = console.Width()
		progress_bar.Progress = float64(scrn.tags_received) / float64(scrn.tags_in_queue)
		hud.Line(&progress_bar)
	case event.StagePullObjects:

		if scrn.duplicate_object_downloads > 0 {
			var duplicate_objects_text console.Text
			duplicate_objects_text.Stylesheet.Margin[console.Left] = 1
			duplicate_objects_text.Stylesheet.Width = console.Width()
			duplicate_objects_text.Add(
				fmt.Sprintf("%d duplicate objects downloaded (%s wasted)", scrn.duplicate_object_downloads, humanize.Bytes(scrn.duplicate_object_download_size)),
				console.BrightYellow,
				0,
			)
			hud.Line(&duplicate_objects_text)
		}

		var usage_text console.Text
		usage_text.Stylesheet.Width = console.Width()
		usage_text.Stylesheet.Margin[console.Left] = 1
		usage_text.Add(fmt.Sprintf("%d/%d objects processed, %s total", scrn.objects_received, scrn.objects_in_queue, humanize.Bytes(scrn.bytes_received)), 0, 0)
		hud.Line(&usage_text)

		// if scrn.objects_received > 0 {
		// 	var last_object_text console.Text
		// 	last_object_text.Stylesheet.Margin[console.Left] = 1
		// 	last_object_text.Stylesheet.Width = console.Width()
		// 	last_object_text.Add(prefix(scrn.last_object_prefix), console.Black, console.White)
		// 	last_object_text.Add(fmt.Sprintf(" %s %s", scrn.last_object_hash, humanize.Bytes(scrn.last_object_size)), 0, 0)
		// 	hud.Line(&last_object_text)
		// }

		progress_bar.Stylesheet.Width = console.Width()
		progress_bar.Progress = float64(scrn.objects_received) / float64(scrn.objects_in_queue)
		hud.Line(&progress_bar)
	case event.StageCacheFile:
		var file_name_text console.Text
		file_name_text.Stylesheet.Width = console.Width()
		file_name_text.Add(scrn.current_file_origin, 0, 0)

		var progress_text console.Text
		progress_text.Stylesheet.Width = 16
		progress_text.Add(fmt.Sprintf("%s/%s", humanize.Bytes(uint64(scrn.current_file_progress)), humanize.Bytes(uint64(scrn.current_file_size))), 0, 0)

		progress_bar.Stylesheet.Width = console.Width() - 16
		progress_bar.Stylesheet.Alignment = console.Right
		progress_bar.Progress = float64(scrn.current_file_progress) / float64(scrn.current_file_size)

		hud.Line(&file_name_text)
		hud.Line(&progress_text, &progress_bar)
	case event.StageCheckout:
		var file_name_text console.Text
		file_name_text.Stylesheet.Width = console.Width()
		file_name_text.Add(scrn.current_file_origin, 0, 0)

		var progress_text console.Text
		progress_text.Stylesheet.Width = 16
		progress_text.Add(fmt.Sprintf("%s/%s", humanize.Bytes(uint64(scrn.current_file_progress)), humanize.Bytes(uint64(scrn.current_file_size))), 0, 0)

		progress_bar.Stylesheet.Width = console.Width() - progress_text.Stylesheet.Width
		// progress_bar.Stylesheet.Alignment = console.Right
		progress_bar.Progress = float64(scrn.current_file_progress) / float64(scrn.current_file_size)

		hud.Line(&file_name_text)
		hud.Line(&progress_text, &progress_bar)
	case event.StageServeObjects:
		var object_upload_count_text console.Text
		object_upload_count_text.Stylesheet.Width = console.Width()
		object_upload_count_text.Add(fmt.Sprintf("%d objects uploaded", scrn.object_uploads), 0, 0)
		hud.Line(&object_upload_count_text)
	}
}
