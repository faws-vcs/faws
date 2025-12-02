package repository

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/faws-vcs/faws/faws/app"
	"github.com/faws-vcs/faws/faws/identity"
	"github.com/faws-vcs/faws/faws/repo/cas"
	"github.com/faws-vcs/faws/faws/repo/revision"
	"github.com/faws-vcs/faws/faws/timestamp"
)

type ViewLogParams struct {
	Directory string
	Ref       string
}

func author_name(attr *identity.Attributes) string {
	email := ""
	if attr.Email != "" {
		email = fmt.Sprintf(" <%s>", attr.Email)
	}
	return attr.Nametag + email
}

func display_commit(commit_hash cas.ContentID) {
	now := time.Now()

	var (
		author      identity.ID
		commit_info *revision.CommitInfo
		err         error
	)
	author, commit_info, err = Repo.GetCommit(commit_hash)
	if err != nil {
		app.Fatal(err)
	}

	ring := app.Configuration.Ring()

	attr := &commit_info.AuthorAttributes
	if ring != nil {
		var trusted_attributes identity.Attributes
		if ring.GetTrustedAttributes(author, &trusted_attributes) == nil {
			attr = &trusted_attributes
		}
	}

	app.Header("commit " + commit_hash.String())

	var tw tabwriter.Writer
	tw.Init(os.Stdout, 0, 0, 8, ' ', 0)

	fmt.Fprintf(&tw, "author:\t%s\n", author_name(attr))
	fmt.Fprintf(&tw, "author identity:\t%s\n", author.String())
	fmt.Fprintf(&tw, "tag:\t%s\n", commit_info.Tag)
	fmt.Fprintf(&tw, "tree:\t%s\n", commit_info.Tree)
	fmt.Fprintf(&tw, "tree date:\t%s (%s)\n", timestamp.Format(commit_info.TreeDate), humanize.RelTime(time.Unix(commit_info.TreeDate, 0), now, "ago", "from now"))
	fmt.Fprintf(&tw, "commit date:\t%s (%s)\n", timestamp.Format(commit_info.CommitDate), humanize.RelTime(time.Unix(commit_info.CommitDate, 0), now, "ago", "from now"))

	tw.Flush()
}

func ViewLog(params *ViewLogParams) {
	app.Open()
	defer func() {
		app.Close()
	}()

	err := Open(params.Directory)
	if err != nil {
		app.Fatal(err)
		return
	}

	var commit_hash cas.ContentID
	commit_hash, err = Repo.ParseRef(params.Ref)
	if err != nil {
		app.Fatal(err)
	}

	var log []cas.ContentID
	log = append(log, commit_hash)

	for commit_hash != cas.Nil {
		var (
			commit_info *revision.CommitInfo
		)
		_, commit_info, err = Repo.GetCommit(commit_hash)
		if err != nil {
			app.Fatal(err)
		}
		commit_hash = commit_info.Parent
		if commit_hash != cas.Nil {
			log = append(log, commit_hash)
		}
	}

	for _, commit_hash := range log {
		display_commit(commit_hash)
		app.Info()
	}
}
