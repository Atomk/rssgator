package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
)

type Config struct {
	DbUrl           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

const CONFIG_FILE_NAME = ".rssgatorconfig.json"

func getConfigFilePath() string {
	homePath, err := os.UserHomeDir()
	if err != nil {
		panic("cannot find home directory")
	}

	return path.Join(homePath, CONFIG_FILE_NAME)
}

// Read the config file and return a pointer to the corresponding Config instance.
func Read() (*Config, error) {
	configPath := getConfigFilePath()

	data, err := os.ReadFile(configPath)
	if os.IsNotExist(err) {
		// Give the caller the possibility to detect this file does not exist
		return nil, err
	} else if err != nil {
		return nil, fmt.Errorf("cannot read file at %s: %s", configPath, err)
	}

	configObj := Config{}
	err = json.Unmarshal(data, &configObj)
	if err != nil {
		return nil, fmt.Errorf("error while parsing config JSON at %s", configPath)
	}

	return &configObj, nil
}

func (c *Config) SetUser(username string) error {
	c.CurrentUserName = username

	// Update config file
	err := c.Write()
	if err != nil {
		return fmt.Errorf("error while updating JSON: %e", err)
	}
	return nil
}

// Write the configuration file to disk.
// The file will be updated, or created if it does not exist.
func (c *Config) Write() error {
	data, err := json.Marshal(c)
	if err != nil {
		return err
	}
	configPath := getConfigFilePath()

	alreadyExisted := true
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		alreadyExisted = false
	}
	err = os.WriteFile(configPath, data, 0777)
	if err != nil {
		return err
	}
	if alreadyExisted {
		fmt.Println("config file updated at " + configPath)
	} else {
		fmt.Println("config file created at " + configPath)
	}
	return nil
}
