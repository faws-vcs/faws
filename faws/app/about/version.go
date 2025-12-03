package about

import (
	"fmt"
	"runtime/debug"

	"github.com/faws-vcs/console"
	"github.com/faws-vcs/faws/faws/app"
)

type VersionParams struct{}

func Version(params *VersionParams) {
	console.Open()
	defer console.Close()

	info, ok := debug.ReadBuildInfo()
	if !ok {
		app.Warning("ReadBuildInfo failed")
	} else {
		app.Info(fmt.Sprintf("%+v", info))
	}
}
