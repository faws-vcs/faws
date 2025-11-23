package repo

import (
	"encoding/json"
	"os"

	"github.com/faws-vcs/faws/faws/fs"
)

const Version = 1

type Config struct {
	// The Faws repository version
	Version uint8 `json:"faws_version"`
	// URL pointing to the original location of the repository
	Remote string `json:"remote,omitempty"`
}

func ReadConfig(filename string, config *Config) (err error) {
	var data []byte
	data, err = os.ReadFile(filename)
	if err != nil {
		return
	}

	err = json.Unmarshal(data, config)
	return
}

func WriteConfig(filename string, config *Config) (err error) {
	var data []byte
	data, err = json.Marshal(config)
	if err != nil {
		return
	}

	err = os.WriteFile(filename, data, fs.DefaultPublicPerm)
	return
}
