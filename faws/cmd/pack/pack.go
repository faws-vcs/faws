package pack

import (
	"os"

	"github.com/dustin/go-humanize"
	"github.com/faws-vcs/faws/faws/app"
	"github.com/faws-vcs/faws/faws/app/repository"
	"github.com/faws-vcs/faws/faws/cmd/helpinfo"
	"github.com/faws-vcs/faws/faws/cmd/root"
	"github.com/spf13/cobra"
)

var pack_cmd = cobra.Command{
	Use:     "pack pack-name",
	Short:   helpinfo.Text["pack"],
	GroupID: "repo",
	Run:     run_pack_cmd,
}

func init() {
	flags := pack_cmd.Flags()
	flags.StringP("max-archive-size", "n", "", "set the maximum size of a pack archive file (e.g. 10K, 50G)")
	root.RootCmd.AddCommand(&pack_cmd)
}

func run_pack_cmd(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		cmd.Help()
		return
	}
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

	var params repository.PackParams
	params.Directory = working_directory
	params.Name = args[0]
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

	repository.Pack(&params)
}
