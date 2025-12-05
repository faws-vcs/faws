package about

import (
	"fmt"
	"os"
	"runtime/debug"
	"text/tabwriter"
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

// VersionParams are the input parameters to the "faws version" command [Version]
type VersionParams struct{}

// Version is the implementation of the "faws version" command
//
// It displays information about the version of Faws currently being used, such as:
//
//	version     (git tag, provided by GoReleaser)
//	commit hash (either provided by GoReleaser or by Go debug.BuildInfo)
//	build date  (provided by GoReleaser)
//	built by    (go or goreleaser)
func Version(params *VersionParams) {
	var tw tabwriter.Writer
	tw.Init(os.Stdout, 0, 0, 8, ' ', 0)

	fmt.Fprintf(&tw, "faws version:\t%s\n", version)
	fmt.Fprintf(&tw, "commit:\t%s\n", commit)
	if date != "" {
		fmt.Fprintf(&tw, "build date:\t%s\n", date)
	}
	if built_by != "" {
		fmt.Fprintf(&tw, "built by:\t%s\n", built_by)
	}
	tw.Flush()
}

// GetVersionString returns Faws's version tag  (included by Goreleaser), or "dev" if it was built with only Go
func GetVersionString() (s string) {
	s = version
	return
}
