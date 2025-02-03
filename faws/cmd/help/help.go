package help

import (
	"fmt"

	"github.com/faws-vcs/faws/faws/cmd/helpinfo"
	"github.com/spf13/cobra"
)

const help_text = `Faws is a toy version control system (VCS).

view and edit repository sources
  source import    import repositories from a source URI
  source ls        list all sources
  source update    update list of repositories
  source rm        remove a repository source

manage authorship identities
  id import        import identities from a source URI
  id create        generate an identity key
  id ls            list all identity keys

locate and retrieve repositories
  repos            list available repositories
  pull             download entire repository into current directory
  export           using a pulled repository, export the files contained at a particular revision into an external directory
  get              download only the files contained at a particular revision into the current directory

manage repository state
  init             create an empty repository in the current repository
  commit           write staged file operations into revision history
  checkout         switch the current revision (aka HEAD) (has no effect on faws export)
  history          view the revision history of the repository up to HEAD
	rebase           

work on the current change
  add              add/update a file/directory into the staging area
  rm               stage a file/directory for removal
	status           
`

var HelpCmd = cobra.Command{
	Use:   "faws help",
	Short: "Faws is a toy version control system (VCS).",
	Run:   run_help_cmd,
}

func run_help_cmd(cmd *cobra.Command, args []string) {
	fmt.Println("Faws is an experimental version control system for hosting large collections of files.")
	fmt.Println()

	max_command_length := int(0)

	// get max command length
	for _, category := range helpinfo.Categories {
		for _, command := range category.Commands {
			if len(command) > max_command_length {
				max_command_length = len(command)
			}
		}
	}

	for _, category := range helpinfo.Categories {
		fmt.Println(category.Description)

		minimum_width := max_command_length + 2

		for _, command := range category.Commands {
			fmt.Print("  ")
			command_length := len(command)
			fmt.Print(command)
			for i := 0; i < (minimum_width - command_length); i++ {
				fmt.Print(" ")
			}
			fmt.Print(helpinfo.Text[command])
			fmt.Println()
		}

		fmt.Println()
	}
}
