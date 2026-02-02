package repository

import (
	"strconv"

	"github.com/faws-vcs/console"
	"github.com/faws-vcs/faws/faws/app"
)

// PruneCacheParams are the input parameters to the command "faws prune", [PruneCache]
type PruneCacheParams struct {
	// The directory where the repository is located
	Directory string
}

// PruneCache is the implementation of the command "faws prune"
//
// It removes all unreachable objects from the cache.
// Unreachable objects may still persist in the repository, if they are part of a pack.
func PruneCache(params *PruneCacheParams) {
	app.Open()
	defer func() {
		app.Close()
	}()

	// get a report before exit
	scrn.summary_mode |= summarize_pruning

	if err := Open(params.Directory); err != nil {
		app.Fatal(err)
	}

	if err := Repo.PruneCache(); err != nil {
		app.Fatal(err)
	}

	Close()
}

func prune_summary(hud *console.Hud) {
	objects_visited_string := strconv.FormatUint(scrn.objects_visited, 10)
	objects_pruned_string := strconv.FormatUint(scrn.objects_pruned, 10)

	num_width := max(len(objects_pruned_string), len(objects_visited_string))

	var visit_summary console.Text
	// var objects_visited_pad strings.Builder
	objects_visited_padding := num_width - len(objects_visited_string)
	// for i := 0; i < objects_visited_padding; i++ {
	// 	visit_summary.Add(" ")
	// }
	visit_summary.Stylesheet.Margin[console.Left] = objects_visited_padding
	visit_summary.Stylesheet.Width = console.Width()
	visit_summary.Add(objects_visited_string, 0, 0)
	visit_summary.Add(" objects ", 0, 0)
	visit_summary.Add("visited", 0, 0)
	// visit_summary.Add("visited", console.Black, console.Blue)

	var prune_summary console.Text
	objects_pruned_padding := num_width - len(objects_pruned_string)
	prune_summary.Stylesheet.Margin[console.Left] = objects_pruned_padding
	prune_summary.Stylesheet.Width = console.Width()
	prune_summary.Add(objects_pruned_string, 0, 0)
	prune_summary.Add(" objects ", 0, 0)
	prune_summary.Add("pruned", 0, 0)
	// prune_summary.Add("pruned", console.Black, console.Red)

	hud.Line(&visit_summary)
	hud.Line(&prune_summary)
}
