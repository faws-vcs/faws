package repository

import (
	"github.com/faws-vcs/faws/faws/app"
	"github.com/faws-vcs/faws/faws/identity"
	"github.com/faws-vcs/faws/faws/repo/p2p/tracker"
)

// PublishParams are the input parameters to the command "faws publish", [Publish]
type PublishParams struct {
	// The repository being published
	Directory string
	//
	TrackerURL string
	// Which identity to use to publish the repository manifest
	Sign string
}

// Publish is the implementation of the command "faws publish"
func Publish(params *PublishParams) {
	if params.TrackerURL != "" {
		TrackerURL = params.TrackerURL
	}

	app.Open()
	defer func() {
		app.Close()
	}()

	if err := Open(params.Directory); err != nil {
		app.Fatal(err)
	}

	var (
		signing_identity     identity.Pair
		publisher_attributes identity.Attributes
	)

	ring := app.Configuration.Ring()

	var err error
	if params.Sign == "" {
		err = ring.GetPrimaryPair(&signing_identity, &publisher_attributes)
		if err != nil {
			app.Warning("You don't seem to have a signing identity yet")
			app.Quote("faws id create")
			app.Info("to create one")
		}
	} else {
		err = ring.GetNametagPair(params.Sign, &signing_identity, &publisher_attributes)
	}
	if err != nil {
		app.Fatal(err)
	}

	var p identity.Pair
	if signing_identity == p {
		panic(signing_identity)
	}

	if params.TrackerURL == "" {
		params.TrackerURL = tracker.DefaultURL
	}

	var topic_uri string
	topic_uri, err = Repo.Publish(&signing_identity, &publisher_attributes)
	if err != nil {
		Close()
		app.Fatal(err)
	}

	app.Info("An updated manifest of the repository was generated and published to the tracker.")
	app.Info("To distribute your copy of the repository to other peers, please run:")

	app.Quote("faws seed ", topic_uri)

	app.Info("Other peers can help you distribute the repository by mirroring a faws topic URI.")
	app.Info("Mirroring means first downloading the entire repository, and then indefinitely serving it up to other peers")

	// topic:4ad45483dda770bdeead7583bf60cca14f3f141287af46f91a53f3d86dd2e091/5527ec56-b2c7-48d9-aa11-90ea4895bcc3

	app.Quote("faws mirror ", topic_uri)

	app.Info("Users may also use this URI as an origin and leech copies from the network:")

	app.Quote("faws clone ", topic_uri)

	Close()

}
