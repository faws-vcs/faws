package repository

import (
	"github.com/faws-vcs/faws/faws/app"
	"github.com/faws-vcs/faws/faws/identity"
	"github.com/faws-vcs/faws/faws/repo/p2p/tracker"
)

type SeedParams struct {
	Directory  string
	Sign       string
	TrackerURL string
	TopicURI   string
	Quiet      bool
}

func Seed(params *SeedParams) {
	var topic tracker.Topic
	if err := tracker.ParseTopicURI(params.TopicURI, &topic); err != nil {
		return
	}

	if params.TrackerURL != "" {
		TrackerURL = params.TrackerURL
	}

	quiet = params.Quiet

	app.Open()
	defer func() {
		app.Close()
	}()

	if err := Open(params.Directory); err != nil {
		app.Fatal(err)
	}

	var (
		signing_identity identity.Pair
	)

	ring := app.Configuration.Ring()

	var err error
	if params.Sign != "" {
		err = ring.GetPair(params.Sign, &signing_identity, nil)
		if err != nil {
			app.Fatal(err)
		}
	}

	err = Repo.Seed(topic, signing_identity)
	if err != nil {
		Close()
		app.Fatal(err)
	}

	Close()
}
