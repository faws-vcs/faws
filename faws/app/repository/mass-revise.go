package repository

import (
	"bytes"
	"io"
	"regexp"
	"time"

	"github.com/faws-vcs/faws/faws/app"
	"github.com/faws-vcs/faws/faws/identity"
	"github.com/faws-vcs/faws/faws/repo/cas"
	"github.com/faws-vcs/faws/faws/repo/revision"
)

// MassReviseParams are the input parameters to the command "faws mass-revise", [MassRevise]
type MassReviseParams struct {
	Directory string
	// if true, remove old commits and tree objects
	// if false, create new commits based on old commits while retaining old information
	Destructive bool
	// if len() != 0, match bytes beginning with this signature
	MatchFileMagic []byte
	MatchTag       *regexp.Regexp
	SetFileMode    bool
	// the new file mode for all files
	NewFileMode revision.FileMode
}

func rewrite_tree(params *MassReviseParams, tree_hash cas.ContentID) (new_tree_hash cas.ContentID) {
	tree, err := Repo.Tree(tree_hash)
	if err != nil {
		app.Fatal("getting tree", err, tree_hash)
	}

	new_tree := *tree

	for i := range new_tree.Entries {
		entry := &new_tree.Entries[i]
		if entry.Prefix == cas.Tree {
			entry.Mode = 0
			entry.Content = rewrite_tree(params, entry.Content)
		} else if entry.Prefix == cas.File {
			var match = true
			if len(params.MatchFileMagic) != 0 {
				file, err := Repo.OpenFile(entry.Content)
				if err != nil {
					app.Fatal("opening file", entry.Content, err)
				}

				file_magic := make([]byte, len(params.MatchFileMagic))
				if _, err = io.ReadFull(file, file_magic); err != nil {
					match = false
				} else {
					if !bytes.Equal(params.MatchFileMagic, file_magic[:]) {
						match = false
					}
				}
				file.Close()
			}

			if match {
				if params.SetFileMode {
					entry.Mode = params.NewFileMode
				}
			}
		}
	}

	new_tree_data, err := revision.MarshalTree(&new_tree)
	if err != nil {
		app.Fatal(err)
	}

	_, new_tree_hash, err = Repo.StoreObject(cas.Tree, new_tree_data)
	if err != nil {
		app.Fatal(err)
	}

	return
}

func rewrite_tag(params *MassReviseParams, tag string) {
	app.Info("rewriting tag", tag)

	commit_hash, err := Repo.ParseRef(tag)
	if err != nil {
		app.Fatal("parse ref", err)
	}

	_, commit_info, err := Repo.GetCommit(commit_hash)
	if err != nil {
		app.Fatal("get commit", err)
	}

	new_commit_info := *commit_info

	new_tree_hash := rewrite_tree(params, commit_info.Tree)
	if new_tree_hash == commit_info.Tree {
		// nothing to be done
		return
	}
	new_commit_info.Tree = new_tree_hash

	var signing_identity identity.Pair
	if err = app.Configuration.Ring().GetPrimaryPair(&signing_identity, &new_commit_info.AuthorAttributes); err != nil {
		app.Fatal(err)
	}

	if params.Destructive {
		Repo.RemoveObject(commit_hash)
	} else {
		new_commit_info.Parent = commit_hash
		new_commit_info.CommitDate = time.Now().Unix()
	}

	var new_commit_hash cas.ContentID
	if new_commit_hash, err = Repo.CommitTree(&signing_identity, &new_commit_info); err != nil {
		return
	}

	app.Info("rewrite", tag, commit_hash, "=>", new_commit_hash)
}

// MassRevise is the implementation of the command "faws mass-revise"
//
// It edits each tag, replacing commits with fixed versions based on the user's match conditions
func MassRevise(params *MassReviseParams) {
	app.Open()
	defer func() {
		app.Close()
	}()

	if !params.SetFileMode {
		// nothing to do
		return
	}

	if err := Open(params.Directory); err != nil {
		app.Fatal(err)
	}

	tags, err := Repo.Tags()
	if err != nil {
		app.Fatal(err)
	}

	for _, tag := range tags {
		if params.MatchTag == nil {
			rewrite_tag(params, tag.Name)
		} else if params.MatchTag.MatchString(tag.Name) {
			rewrite_tag(params, tag.Name)
		}
	}

	Close()
}
