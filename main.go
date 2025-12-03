package main

import (
	"fmt"

	"github.com/faws-vcs/faws/faws/cmd"
)

var (
	version = "dev"     // Default value, overridden by GoReleaser
	commit  = "none"    // Default value, overridden by GoReleaser
	date    = "unknown" // Default value, overridden by GoReleaser
)

func main() {
	fmt.Printf("My App Version: %s, Commit: %s, Built At: %s\n", version, commit, date)
	cmd.Execute()
}
