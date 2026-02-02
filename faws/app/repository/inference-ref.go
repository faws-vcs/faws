package repository

import (
	"os"
	"strings"

	"github.com/faws-vcs/faws/faws/app"
	"github.com/faws-vcs/faws/faws/validate"
	"github.com/spf13/cobra"
)

type InferenceRefParams struct {
	Directory  string
	Ref        string
	Inferences []string
}

func InferenceRef(params *InferenceRefParams) {
	if err := Open(params.Directory); err != nil {
		app.Fatal(err)
	}

	tags, err := Repo.Tags()
	if err != nil {
		app.Fatal(err)
	}
	for _, tag := range tags {
		if params.Ref == "" || strings.HasPrefix(tag.Name, params.Ref) {
			params.Inferences = append(params.Inferences, tag.Name)
		}
	}
	// if the ref is hexadecimal, you can also try to autocomplete it
	if validate.Hex(params.Ref) {
		parsed_ref, parse_ref_err := Repo.ParseRef(params.Ref)
		if parse_ref_err == nil {
			params.Inferences = append(params.Inferences, parsed_ref.String())
		}
	}
	if err := Close(); err != nil {
		app.Fatal(err)
		return
	}
}

func InferenceRefArg(n int) cobra.CompletionFunc {
	return func(cmd *cobra.Command, args []string, toComplete string) (completion []cobra.Completion, shell_comp_directive cobra.ShellCompDirective) {
		if len(args) != n {
			return
		}
		working_directory, err := os.Getwd()
		if err != nil {
			app.Fatal(err)
			return
		}

		var params InferenceRefParams
		params.Directory = working_directory
		params.Ref = toComplete

		InferenceRef(&params)

		completion = make([]cobra.Completion, len(params.Inferences))
		for i := range params.Inferences {
			completion[i] = cobra.Completion(params.Inferences[i])
		}
		return
	}
}
