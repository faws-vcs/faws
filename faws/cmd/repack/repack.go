package repack

import (
	"os"

	"github.com/dustin/go-humanize"
	"github.com/faws-vcs/faws/faws/app"
	"github.com/faws-vcs/faws/faws/app/repository"
	"github.com/faws-vcs/faws/faws/cmd/helpinfo"
	"github.com/faws-vcs/faws/faws/cmd/root"
	"github.com/spf13/cobra"
)

var repack_cmd = cobra.Command{
	Use:     "repack",
	Short:   helpinfo.Text["repack"],
	GroupID: "repo",
	Run:     run_repack_cmd,
}

func init() {
	flags := repack_cmd.Flags()
	flags.StringP("max-archive-size", "n", "", "set the maximum size of a pack archive file (e.g. 10K, 50G)")
	root.RootCmd.AddCommand(&repack_cmd)
}

func run_repack_cmd(cmd *cobra.Command, args []string) {
	flags := cmd.Flags()
	var (
		err               error
		working_directory string
	)
	// use working directory as default repository location
	working_directory, err = os.Getwd()
	if err != nil {
		app.Fatal(err)
		return
	}

	var params repository.RepackParams
	params.Directory = working_directory

	var max_archive_size string
	max_archive_size, err = flags.GetString("max-archive-size")
	if err != nil {
		app.Fatal(err)
	}
	if max_archive_size == "" {
		params.MaxArchiveSize = -1
	} else {
		var max_archive_size_u64 uint64
		max_archive_size_u64, err = humanize.ParseBytes(max_archive_size)
		if err != nil {
			app.Fatal(err)
		}
		params.MaxArchiveSize = int64(max_archive_size_u64)
	}

	repository.Repack(&params)
}
