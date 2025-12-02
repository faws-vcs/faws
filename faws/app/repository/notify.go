package repository

import (
	"fmt"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/faws-vcs/console"
	"github.com/faws-vcs/faws/faws/app"
	"github.com/faws-vcs/faws/faws/repo/cas"
	"github.com/faws-vcs/faws/faws/repo/event"
)

var (
	stages_text = map[event.Stage]string{
		event.StagePullObjects: "Retrieve objects",
		event.StagePullTags:    "Retrieve tags",
	}

	scrn activity_screen

	guard sync.Mutex
)

type activity_stage struct {
	stage event.Stage
	state int
}

type activity_screen struct {
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

	verbose bool
}

func begin_stage(stage event.Stage) {
	scrn.stages = append(scrn.stages, activity_stage{stage, 0})
}

func complete_stage(stage event.Stage, success bool) {
	if stage != scrn.stages[len(scrn.stages)-1].stage {
		panic("stage complete mismatch")
	}
	if success {
		scrn.stages[len(scrn.stages)-1].state = 1
	} else {
		scrn.stages[len(scrn.stages)-1].state = 2
	}
}

func update_pull_info(object_size int, object_prefix cas.Prefix, object_hash cas.ContentID) {
	scrn.bytes_received += uint64(object_size)
	scrn.objects_received++

	scrn.last_object_prefix = object_prefix
	scrn.last_object_hash = object_hash
	scrn.last_object_size = uint64(object_size)
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
	guard.Lock()

	switch ev {
	case event.NotifyCacheFile:
		app.Info("caching", params.Name1, params.Name2)
	case event.NotifyCacheUsedLazySignature:
		app.Info("using precached file (--lazy)", params.Name1, params.Name2)
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
	case event.NotifyPullObject:
		var (
			object_prefix = params.Prefix
			object_hash   = params.Object1
			object_size   = params.Count
		)

		update_pull_info(object_size, object_prefix, object_hash)
	case event.NotifyTagQueueCount:
		scrn.tags_in_queue = params.Count
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

	console.SwapHud()
}

func render_activity_screen(hud *console.Hud) {
	guard.Lock()
	defer guard.Unlock()
	var spinner console.Spinner
	spinner.Stylesheet.Sequence[0] = console.Cell{'⡿', console.BrightBlue, 0}
	spinner.Stylesheet.Sequence[1] = console.Cell{'⣟', console.BrightBlue, 0}
	spinner.Stylesheet.Sequence[2] = console.Cell{'⣯', console.BrightBlue, 0}
	spinner.Stylesheet.Sequence[3] = console.Cell{'⣷', console.BrightBlue, 0}
	spinner.Stylesheet.Sequence[4] = console.Cell{'⣾', console.BrightBlue, 0}
	spinner.Stylesheet.Sequence[5] = console.Cell{'⣽', console.BrightBlue, 0}
	spinner.Stylesheet.Sequence[6] = console.Cell{'⣻', console.BrightBlue, 0}
	spinner.Stylesheet.Sequence[7] = console.Cell{'⢿', console.BrightBlue, 0}
	spinner.Frequency = time.Second / 3

	var stage *activity_stage
	for i := range scrn.stages {
		stage = &scrn.stages[i]
		message := fmt.Sprintf("%d) %s ", i+1, stages_text[stage.stage])
		var stage_text console.Text
		switch stage.state {
		case 0:
			stage_text.Add(message, console.BrightBlue, 0)
			stage_text.Stylesheet.Width = len(message)
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

	var progress_bar console.ProgressBar
	progress_bar.Stylesheet.Sequence[console.PbCaseLeft] = console.Cell{'[', 0, 0}
	progress_bar.Stylesheet.Sequence[console.PbCaseRight] = console.Cell{']', 0, 0}
	progress_bar.Stylesheet.Sequence[console.PbFluid] = console.Cell{'#', 0, 0}
	progress_bar.Stylesheet.Sequence[console.PbVoid] = console.Cell{'.', 0, 0}
	progress_bar.Stylesheet.Sequence[console.PbTail] = console.Cell{'#', 0, 0}
	progress_bar.Stylesheet.Sequence[console.PbHead] = console.Cell{'#', 0, 0}

	switch stage.stage {
	case event.StagePullTags:
		var progress_text console.Text
		progress_text.Stylesheet.Width = console.Width()
		progress_text.Add(fmt.Sprintf("%d/%d tags received", scrn.tags_received, scrn.tags_in_queue), 0, 0)
		hud.Line(&progress_text)

		progress_bar.Stylesheet.Width = console.Width()
		progress_bar.Progress = float64(scrn.tags_received) / float64(scrn.tags_in_queue)
		hud.Line(&progress_bar)
	case event.StagePullObjects:
		var usage_text console.Text
		usage_text.Stylesheet.Width = console.Width()
		usage_text.Stylesheet.Margin[console.Left] = 1
		usage_text.Add(fmt.Sprintf("%d/%d objects received, %s total", scrn.objects_received, scrn.objects_in_queue, humanize.Bytes(scrn.bytes_received)), 0, 0)
		hud.Line(&usage_text)

		if scrn.objects_received > 0 {
			var last_object_text console.Text
			last_object_text.Stylesheet.Margin[console.Left] = 1
			last_object_text.Stylesheet.Width = console.Width()
			last_object_text.Add(prefix(scrn.last_object_prefix), console.Black, console.White)
			last_object_text.Add(fmt.Sprintf(" %s %s", scrn.last_object_hash, humanize.Bytes(scrn.last_object_size)), 0, 0)
			hud.Line(&last_object_text)
		}

		progress_bar.Stylesheet.Width = console.Width()
		progress_bar.Progress = float64(scrn.objects_received) / float64(scrn.objects_in_queue)
		hud.Line(&progress_bar)
	}
}
