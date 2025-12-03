package about

import (
	"bytes"
	"fmt"
	"runtime/debug"
	"text/tabwriter"

	"github.com/faws-vcs/console"
	"github.com/faws-vcs/faws/faws/app"
)

var (
	version  = "dev"
	commit   = ""
	date     = ""
	built_by = ""
)

func init() {
	if commit == "" {
		// not built with
		info, ok := debug.ReadBuildInfo()
		if ok {
			for _, build_setting := range info.Settings {
				if build_setting.Key == "vcs.revision" {
					commit = build_setting.Value
					break
				}
			}
		}
	}
}

type VersionParams struct{}

func Version(params *VersionParams) {
	console.Open()
	defer console.Close()

	var w bytes.Buffer
	var tw tabwriter.Writer
	tw.Init(&w, 0, 0, 8, ' ', 0)

	fmt.Fprintf(&tw, "faws version:\t%s\n", version)
	fmt.Fprintf(&tw, "commit:\t%s\n", commit)
	if date != "" {
		fmt.Fprintf(&tw, "build date:\t%s\n", date)
	}
	if built_by != "" {
		fmt.Fprintf(&tw, "built by:\t%s\n", built_by)
	}
	tw.Flush()

	app.Info(w.String())
}
