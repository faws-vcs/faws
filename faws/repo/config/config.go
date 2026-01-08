package config

import (
	"encoding/json"
	"os"

	"github.com/faws-vcs/faws/faws/fs"
	"github.com/google/uuid"
)

const Format = 1

// Config is the format for the "config" JSON file in the root of the repository
type Config struct {
	AppID string `json:"app_id"`
	// The version of Faws
	// This is purely diagnostic. Breaking changes should be demonstrated by RepositoryFormat
	AppVersion string `json:"app_version"`
	// The Faws repository version
	RepositoryFormat uint8 `json:"repository_format"`
	// The unique identifier of this repository
	UUID uuid.UUID `json:"uuid"`
	// URL pointing to the original location of the repository
	Origin string `json:"origin,omitempty"`
}

// ReadConfig reads a config at the filename
func ReadConfig(filename string, config *Config) (err error) {
	var data []byte
	data, err = os.ReadFile(filename)
	if err != nil {
		return
	}

	err = json.Unmarshal(data, config)
	return
}

// WriteConfig saves a config to the filename
func WriteConfig(filename string, config *Config) (err error) {
	var data []byte
	data, err = json.Marshal(config)
	if err != nil {
		return
	}

	err = os.WriteFile(filename, data, fs.DefaultPublicPerm)
	return
}
