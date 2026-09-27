package config

import (
	"encoding/json"
	"os"
	"path"
)

type Config struct {
	DbUrl string `json:"db_url"`
}

const CONFIG_FILE_NAME = ".rssgatorconfig.json"

// Read the config file and return the corresponding Config instance.
func Read() Config {
	homePath, err := os.UserHomeDir()
	if err != nil {
		panic("cannot find home directory")
	}

	configPath := path.Join(homePath, CONFIG_FILE_NAME)

	data, err := os.ReadFile(configPath)
	if err != nil {
		panic("cannot read file at " + configPath)
	}

	configObj := Config{}
	err = json.Unmarshal(data, &configObj)
	if err != nil {
		panic("error while parsing config JSON at " + configPath)
	}

	return configObj
}
