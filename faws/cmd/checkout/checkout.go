package commit_tree

import (
	"os"

	"github.com/faws-vcs/faws/faws/app"
	"github.com/faws-vcs/faws/faws/app/repository"
	"github.com/faws-vcs/faws/faws/cmd/helpinfo"
	"github.com/faws-vcs/faws/faws/cmd/root"
	"github.com/spf13/cobra"
)

var checkout_cmd = cobra.Command{
	Use:               "checkout [-w] object-hash destination",
	Short:             helpinfo.Text["checkout"],
	GroupID:           "repo",
	Run:               run_checkout_cmd,
	ValidArgsFunction: repository.InferenceRefArg(0),
}

func init() {
	flags := checkout_cmd.Flags()
	flags.BoolP("overwrite", "w", false, "overwrite any files that may exist at the destination")
	root.RootCmd.AddCommand(&checkout_cmd)
}

func run_checkout_cmd(cmd *cobra.Command, args []string) {
	if len(args) < 2 {
		cmd.Help()
		os.Exit(1)
	}

	flags := cmd.Flags()
	overwrite, err := flags.GetBool("overwrite")
	if err != nil {
		app.Fatal(err)
	}

	// use working directory as default repository location
	working_directory, err := os.Getwd()
	if err != nil {
		app.Fatal(err)
		return
	}

	var params = repository.CheckoutParams{
		Directory:   working_directory,
		Ref:         args[0],
		Destination: args[1],
		Overwrite:   overwrite,
	}

	repository.Checkout(&params)
}
