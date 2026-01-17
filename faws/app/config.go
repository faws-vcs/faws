package app

import (
	"os"
	"path/filepath"
	"runtime/pprof"

	"github.com/faws-vcs/console"
	"github.com/faws-vcs/faws/faws/config"
)

// The user's local Faws configuration (not accessible until [Open] is called)
var Configuration config.Configuration

var (
	cpu_profile_active bool
	cpu_profile_file   *os.File
)

// Opens the console, and loads the user's local Faws configuration
func Open() {
	console.Open()

	directory := os.Getenv("FAWS_CONFIG")

	if directory == "" {
		user_config_directory, err := os.UserConfigDir()
		if err != nil {
			Fatal(err)
		}
		directory = filepath.Join(user_config_directory, "faws")
	}

	cpu_profile_name := os.Getenv("FAWS_CPU_PROFILE")
	if cpu_profile_name != "" {
		cpu_profile_active = true
		var err error
		cpu_profile_file, err = os.Create(cpu_profile_name)
		if err != nil {
			Fatal(err)
		}
		pprof.StartCPUProfile(cpu_profile_file)
	}

	if err := Configuration.Open(directory); err != nil {
		Fatal(err)
	}

}

// Close terminates the program and saves any changes made to the configuration
func Close() {
	if err := Configuration.Close(); err != nil {
		Fatal(err)
	}

	// save CPU profile
	if cpu_profile_active {
		pprof.StopCPUProfile()
		cpu_profile_file.Close()
	}

	// write a heap profile
	heap_profile_name := os.Getenv("FAWS_HEAP_PROFILE")
	if heap_profile_name != "" {
		heap_profile_file, err := os.Create(heap_profile_name)
		if err != nil {
			Fatal(err)
		}
		if err = pprof.WriteHeapProfile(heap_profile_file); err != nil {
			return
		}
		heap_profile_file.Close()
	}

	console.Close()
}
