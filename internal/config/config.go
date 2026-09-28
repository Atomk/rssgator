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

// Read the config file and return the corresponding Config instance.
func Read() Config {
	configPath := getConfigFilePath()

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

func (c *Config) SetUser(username string) {
	c.CurrentUserName = username

	// Update config file
	err := c.write()
	if err != nil {
		fmt.Printf("error while updating JSON: %e", err)
	}
}

func (c *Config) write() error {
	data, err := json.Marshal(c)
	if err != nil {
		return err
	}
	configPath := getConfigFilePath()
	err = os.WriteFile(configPath, data, 0777)
	if err != nil {
		return err
	}
	fmt.Println("config file updated at " + configPath)
	return nil
}
